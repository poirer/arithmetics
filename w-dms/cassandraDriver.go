package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"

	"github.com/gocql/gocql"
)

const (
	nullUUIDStr = "00000000-0000-0000-0000-000000000000"

	casInsertClause = `insert into Tasks(id, alias, description, type_id, task_timestamp, estimate_time, real_time, reminders)
  values(?, ?, ?, ?, ?, ?, ?, ?)`
	casDeleteClause = "delete from Tasks where id = ?"
	casUpdateClause = `update Tasks
      set alias = ?,
      description = ?,
      type_id = ?,
      task_timestamp = ?,
      estimate_time = ?,
      real_time = ?,
      reminders = ?
    where id = ?`
	casSelectClause = `select id, alias, description, type_id, task_timestamp, estimate_time, real_time, reminders from Tasks`

	casSelectTagNamesClause = "select tag from Tags where task_id = ?"
	casSelectTagIDsClause   = "select id from Tags where task_id = ?"
	casInsertTagClause      = "insert into Tags(id, task_id, tag) values(uuid(), ?, ?)"
	casDeleteTagsClause     = "delete from Tags where id = ?"

	casSelectTypeByNameClause = `select id from Types where type = ?`
	casSelectTypeByIDClause   = `select type from Types where id = ?`
	casInsertTypeClause       = `insert into Types(id, type) values(?, ?)`
)

var nullUUID gocql.UUID

type cassandraDriver struct {
	session *gocql.Session
}

func newCassandraDriver(keyspace string) *cassandraDriver {
	var err error
	nullUUID, err = gocql.ParseUUID(nullUUIDStr)
	if err != nil {
		log.Fatal("Cannot initialize cassandra driver: ", err)
	}
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = keyspace
	cluster.ProtoVersion = 3
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Cannot connect to cassandra database: ", err)
	}
	return &cassandraDriver{session}
}

func (d *cassandraDriver) Create(t Task) error {
	typeID, err := d.findOrInsertType(t.Type)
	if err != nil {
		return err
	}
	taskUUID := gocql.TimeUUID()
	// TODO: the operation below inserts empty strings instead of nulls
	err = d.session.Query(casInsertClause, taskUUID, t.Alias, t.Description, typeID, t.Timestamp, t.EstimateTime, t.RealTime, t.Reminders).Exec()
	if err != nil {
		return err
	}
	err = d.insertTaskTags(taskUUID, t.Tags)
	if err != nil {
		_ = d.session.Query(casDeleteTagsClause, taskUUID).Exec()
		return err
	}
	return nil
}

func (d *cassandraDriver) ReadByID(id interface{}) (TaskList, error) {
	if id != nil {
		return d.selectTasksByCondition("where id = ?", id)
	}
	return d.selectTasksByCondition("")
}

func (d *cassandraDriver) ReadByAlias(alias *string) (TaskList, error) {
	if alias != nil {
		return d.selectTasksByCondition("where alias = ?", *alias)
	}
	return d.selectTasksByCondition("")
}

func (d *cassandraDriver) Update(t Task) error {
	typeID, err := d.findOrInsertType(t.Type)
	if err != nil {
		return err
	}
	d.session.Query(casUpdateClause, t.Alias, t.Description, typeID, t.Timestamp, t.EstimateTime, t.RealTime, t.Reminders, t.ID).Exec()
	taskID, err := gocql.ParseUUID(t.ID.(string))
	if err != nil {
		return err
	}
	err = d.deleteTaskTags(taskID)
	if err != nil {
		return err
	}
	return d.insertTaskTags(taskID, t.Tags)
}

func (d *cassandraDriver) Delete(t Task) error {
	if t.ID == nil {
		return errMissingID
	}
	taskID, err := gocql.ParseUUID(t.ID.(string))
	if err != nil {
		return err
	}
	err = d.deleteTaskTags(taskID)
	if err != nil {
		return err
	}
	return d.session.Query(casDeleteClause, t.ID).Exec()
}

func (d *cassandraDriver) Close() error {
	d.session.Close()
	return nil
}

func (d *cassandraDriver) Init() error {
	scriptFile, err := os.Open("cassandra-script.cql")
	if err != nil {
		return err
	}
	defer scriptFile.Close()
	scripts, err := ioutil.ReadAll(scriptFile)
	if err != nil {
		return err
	}
	clauses := bytes.Split(scripts, []byte{';'})
	for i := 0; i < len(clauses); i++ {
		clause := bytes.TrimFunc(clauses[i], func(r rune) bool { return r == '\n' || r == '\t' || r == ' ' })
		if len(clause) > 0 {
			err = d.session.Query(string(clause)).Exec()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *cassandraDriver) findOrInsertType(typeName string) (gocql.UUID, error) {
	var typeUUID gocql.UUID
	err := d.session.Query(casSelectTypeByNameClause, typeName).Consistency(gocql.One).Scan(&typeUUID)
	if err == gocql.ErrNotFound {
		typeUUID = gocql.TimeUUID()
		insErr := d.session.Query(casInsertTypeClause, typeUUID, typeName).Exec()
		if insErr != nil {
			return nullUUID, insErr
		}
	} else if err != nil {
		return nullUUID, err
	}
	return typeUUID, nil
}

func (d *cassandraDriver) selectTasksByCondition(condition string, args ...interface{}) (TaskList, error) {
	var id, typeID gocql.UUID
	var alias, desc, eTime, rTime string
	var ts int64
	var reminders []string
	var selectError error
	resultIt := d.session.Query(casSelectClause+" "+condition, args...).Iter()
	var tasks = TaskList{}
	for resultIt.Scan(&id, &alias, &desc, &typeID, &ts, &eTime, &rTime, &reminders) {
		t := Task{ID: id, Alias: alias, Description: desc, Timestamp: ts, RealTime: rTime, EstimateTime: eTime, Reminders: reminders, Tags: []string{}}
		var typeName string
		err := d.session.Query(casSelectTypeByIDClause, typeID).Consistency(gocql.One).Scan(&typeName)
		if err != nil {
			selectError = err
			break
		}
		t.Type = typeName
		var tag string
		tagIt := d.session.Query(casSelectTagNamesClause, id).Iter()
		for tagIt.Scan(&tag) {
			t.Tags = append(t.Tags, tag)
		}
		err = tagIt.Close()
		if err != nil {
			selectError = err
			break
		}
		tasks = append(tasks, t)
	}
	if selectError != nil {
		_ = resultIt.Close()
		return nil, selectError
	}
	return tasks, nil
}

func (d *cassandraDriver) deleteTaskTags(taskID gocql.UUID) error {
	tagIt := d.session.Query(casSelectTagIDsClause, taskID).Iter()
	// Just couldn't delete all tags in one statement
	var tagID gocql.UUID
	for tagIt.Scan(&tagID) {
		err := d.session.Query(casDeleteTagsClause, tagID).Exec()
		if err != nil {
			return err
		}
	}
	return tagIt.Close()
}

func (d *cassandraDriver) insertTaskTags(taskID gocql.UUID, tags []string) error {
	tagBatch := d.session.NewBatch(gocql.LoggedBatch)
	for _, tag := range tags {
		tagBatch.Query(casInsertTagClause, taskID, tag)
	}
	return d.session.ExecuteBatch(tagBatch)
}
