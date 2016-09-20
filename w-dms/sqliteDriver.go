package main

import (
	"database/sql"
	"log"
	"strconv"
)

const (
	insertClause = `insert into Tasks(alias, description, type_id, task_timestamp, estimate_time, real_time)
 values(?, ?, ?, ?, ?, ?)`

	selectCaluse = `select t.id, t.alias, t.description, t.task_timestamp, t.estimate_time, t.real_time, ty.type, r.reminder, ta.tag
from Tasks t
inner join Types ty on t.type_id = ty.id
left join Tags ta on ta.task_id = t.id
left join Reminders r on r.task_id = t.id`

	updateClause = `update Tasks
set alias = ?,
    description = ?,
    type_id = ?,
    task_timestamp = ?,
    estimate_time = ?,
    real_time = ?
where id = ?
`
)

type sqliteDriver struct {
	db *sql.DB
}

func newSqliteDriver(dbURL string) *sqliteDriver {
	db, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database", err)
	}
	return &sqliteDriver{db}
}

func (d *sqliteDriver) Create(t Task) error {
	var canCommit = false
	transaction, err := d.db.Begin()
	if err != nil {
		return nil
	}
	defer func() {
		if canCommit {
			transaction.Commit()
		} else {
			transaction.Rollback()
		}
	}()
	typeID, err := d.findTypeID(t.Type)
	if err != nil {
		return err
	}
	res, err := transaction.Exec(insertClause, t.Alias, t.Description, typeID, t.Timestamp, t.EstimateTime, t.RealTime)
	if err != nil {
		return err
	}
	taskID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = int(taskID)
	if err = addTaskTagsAndRemindersInTransaction(t, transaction); err != nil {
		return err
	}
	canCommit = true
	return nil
}

func (d *sqliteDriver) ReadByID(id interface{}) (TaskList, error) {
	if id != nil {
		iID, err := strconv.ParseInt(id.(string), 10, 0)
		if err != nil {
			return nil, errInvalidID
		}

		return d.readByCondition("where t.id = ?", iID)
	}
	return d.readByCondition("")
}

func (d *sqliteDriver) ReadByAlias(alias *string) (TaskList, error) {
	if alias != nil {
		return d.readByCondition("where t.alias = ?", alias)
	}
	return d.readByCondition("")
}

func (d *sqliteDriver) Update(t Task) error {
	var canCommit = false
	transaction, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if canCommit {
			transaction.Commit()
		} else {
			transaction.Rollback()
		}
	}()
	if err = cleanTagsAndRemindersInTransaction(t, transaction); err != nil {
		return err
	}
	typeID, err := d.findTypeID(t.Type)
	if err != nil {
		return err
	}
	res, err := transaction.Exec(updateClause, t.Alias, t.Description, typeID, t.Timestamp, t.EstimateTime, t.RealTime, t.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err == sql.ErrNoRows || count == 0 {
		return errTaskNotFound
	} else if err != nil {
		return err
	}
	if err = addTaskTagsAndRemindersInTransaction(t, transaction); err != nil {
		return err
	}
	canCommit = true
	return nil
}

func (d *sqliteDriver) Delete(t Task) error {
	id, err := strconv.ParseInt(t.ID.(string), 10, 0)
	if err != nil {
		return err
	}
	t.ID = id
	var canCommit = false
	transaction, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if canCommit {
			transaction.Commit()
		} else {
			transaction.Rollback()
		}
	}()
	if err = cleanTagsAndRemindersInTransaction(t, transaction); err != nil {
		return err
	}
	res, err := transaction.Exec("delete from Tasks where id = ?", t.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err == sql.ErrNoRows || count == 0 {
		return errTaskNotFound
	} else if err != nil {
		return err
	}
	canCommit = true
	return nil
}

func (d *sqliteDriver) findTypeID(typeName string) (int, error) {
	row := d.db.QueryRow("select id from Types where type = ?", typeName)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func addTaskTagsAndRemindersInTransaction(t Task, tx *sql.Tx) error {
	var tagStm, remStm *sql.Stmt
	defer func() {
		if tagStm != nil {
			tagStm.Close()
		}
		if remStm != nil {
			remStm.Close()
		}
	}()
	tagStm, err := tx.Prepare("insert into Tags(task_id, tag) values(?, ?)")
	if err != nil {
		return err
	}
	for _, tag := range t.Tags {
		_, err = tagStm.Exec(t.ID, tag)
		if err != nil {
			return err
		}
	}
	remStm, err = tx.Prepare("insert into Reminders(task_id, reminder) values(?, ?)")
	if err != nil {
		return err
	}
	for _, reminder := range t.Reminders {
		_, err = remStm.Exec(t.ID, reminder)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanTagsAndRemindersInTransaction(task Task, tx *sql.Tx) error {
	_, err := tx.Exec("delete from Tags where task_id = ?", task.ID)
	if err != nil {
		return err
	}
	_, err = tx.Exec("delete from Reminders where task_id = ?", task.ID)
	return err
}

func (d *sqliteDriver) readByCondition(condition string, args ...interface{}) (TaskList, error) {
	var taskMap = make(map[int]*Task, 0)
	var rows *sql.Rows
	var err error
	rows, err = d.db.Query(selectCaluse+" "+condition, args...)
	if err != nil {
		return nil, err
	}
	var id int
	var ts int64
	var alias, typ string
	var descr, rt, et, tag, rem sql.NullString
	for rows.Next() {
		err := rows.Scan(&id, &alias, &descr, &ts, &rt, &et, &typ, &rem, &tag)
		if err != nil {
			return nil, err
		}
		task, exists := taskMap[id]
		if !exists {
			task = &Task{ID: id, Alias: alias, Timestamp: ts, Type: typ, Reminders: []string{}, Tags: []string{}}
			if descr.Valid {
				task.Description = descr.String
			}
			if et.Valid {
				task.EstimateTime = et.String
			}
			if rt.Valid {
				task.RealTime = rt.String
			}
			taskMap[id] = task
		}
		if tag.Valid {
			addToSliceIfAbsent(&task.Tags, tag.String)
		}
		if rem.Valid {
			addToSliceIfAbsent(&task.Reminders, rem.String)
		}
	}
	var list = make(TaskList, len(taskMap))
	var i int
	for id = range taskMap {
		list[i] = *taskMap[id]
		i++
	}
	if len(list) == 0 && condition != "" {
		return nil, errTaskNotFound
	}
	return list, nil
}

func (d *sqliteDriver) Close() error {
	return d.db.Close()
}

func addToSliceIfAbsent(slice *[]string, value string) {
	for _, s := range *slice {
		if s == value {
			return
		}
	}
	*slice = append(*slice, value)
}
