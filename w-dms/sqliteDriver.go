package main

import (
	"database/sql"
	"log"
	"time"
)

const selectCaluse = `select t.id, t.alias, t.description, t.task_timestamp, t.estimate_time, t.real_time, ty.type, r.reminder, ta.tag
from Tasks t
inner join Types ty on t.type_id = ty.id
left join Tags ta on ta.task_id = t.id
left join Reminders r on r.task_id = t.id`

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

func (d *sqliteDriver) findTypeID(typeName string) (int, error) {
	row := d.db.QueryRow("select id from Types where type = ?", typeName)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *sqliteDriver) Create(t Task) error {
	transaction, err := d.db.Begin()
	if err != nil {
		return nil
	}
	typeID, err := d.findTypeID(t.Type)
	if err != nil {
		transaction.Rollback()
		return err
	}
	var timestamp = time.Now().Unix()
	res, err := transaction.Exec(`insert into Tasks(alias, description, type_id, task_timestamp, estimate_time, real_time)
   values(?, ?, ?, ?, ?, ?)`, t.Alias, t.Description, typeID, timestamp, t.EstimateTime, t.RealTime)
	if err != nil {
		transaction.Rollback()
		return err
	}
	taskID, err := res.LastInsertId()
	if err != nil {
		transaction.Rollback()
		return err
	}
	for _, tag := range t.Tags {
		_, err = transaction.Exec("insert into Tags(task_id, tag) values(?, ?)", taskID, tag)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}
	for _, reminder := range t.Reminders {
		_, err = transaction.Exec("insert into Reminders(task_id, reminder) values(?, ?)", taskID, reminder)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}
	transaction.Commit()
	return nil
}

func (d *sqliteDriver) ReadByID(id *int64) (TaskList, error) {
	var taskMap = make(map[int]Task, 0)
	var rows *sql.Rows
	var err error
	if id == nil {
		rows, err = d.db.Query(selectCaluse)
	} else {
		rows, err = d.db.QueryRow(selectCaluse+" where t.id = ?", id)
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		// rows.Scan(dest)
	}
	return nil, nil
}

func (d *sqliteDriver) ReadByAlias(alias *string) (TaskList, error) {
	return nil, nil
}

func (d *sqliteDriver) Update(t Task) error {
	return nil
}

func (d *sqliteDriver) Delete(t Task) error {
	return nil
}
