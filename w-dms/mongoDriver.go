package main

import (
	"gopkg.in/mgo.v2"
	"log"
	"gopkg.in/mgo.v2/bson"
)

type mongoDriver struct {
	session  *mgo.Session
	database string
}

func newMongoDriver(dbName string) *mongoDriver {
	mongoSession, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal("Cannot connect to mongo database", err)
	}
	return &mongoDriver{mongoSession, dbName}
}

func (d *mongoDriver) Create(t Task) error {
	var c *mgo.Collection = d.session.DB(d.database).C("Tasks")
	return c.Insert(&t)
}
func (d *mongoDriver) ReadByID(id *int64) (TaskList, error) {
	var c *mgo.Collection = d.session.DB(d.database).C("Tasks")
	var tasks TaskList
	if id == nil {
		err := c.Find(nil).All(&tasks)
		if err != nil {
			return nil, err
		}
	} else {
		var task Task
		err := c.Find("").One(&task)
		if err != nil {
			return nil, err
		}
		tasks = make(TaskList, 1)
		tasks[0] = task
	}
	return tasks, nil
}
func (d *mongoDriver) ReadByAlias(alias *string) (TaskList, error) {
	var c *mgo.Collection = d.session.DB(d.database).C("Tasks")
	var tasks TaskList
	if alias == nil {
		err := c.Find(nil).All(&tasks)
		if err != nil {
			return nil, err
		}
	} else {
		err := c.Find(bson.M{"alias":"FT"}).All(&tasks)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil }
func (d *mongoDriver) Update(t Task) error                         { return nil }
func (d *mongoDriver) Delete(t Task) error {
	return nil
}
func (d *mongoDriver) Close() error {
	d.session.Close()
	return nil
}
