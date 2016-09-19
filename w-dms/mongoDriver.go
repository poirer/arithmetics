package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
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
	var c *mgo.Collection = d.getTasksC()
	return c.Insert(&t)
}

func (d *mongoDriver) ReadByID(id interface{}) (TaskList, error) {
	var c *mgo.Collection = d.session.DB(d.database).C("Tasks")
	var tasks TaskList
	if id == nil {
		err := c.Find(nil).All(&tasks)
		if err != nil && err == mgo.ErrNotFound {
			return nil, taskNotFoundError
		} else if err != nil {
			return nil, err
		}
	} else {
		var task Task
		objIndex := bson.ObjectIdHex(*id.(*string))
		err := c.Find(bson.M{"_id": objIndex}).One(&task)
		if err != nil && err == mgo.ErrNotFound {
			return TaskList{}, nil
		} else if err != nil {
			return nil, err
		}
		tasks = make(TaskList, 1)
		tasks[0] = task
	}
	return tasks, nil
}

func (d *mongoDriver) ReadByAlias(alias *string) (TaskList, error) {
	var c *mgo.Collection = d.getTasksC()
	var tasks TaskList
	if alias == nil {
		err := c.Find(nil).All(&tasks)
		if err != nil && err == mgo.ErrNotFound {
			return TaskList{}, nil
		} else if err != nil {
			return nil, err
		}
	} else {
		err := c.Find(bson.M{"alias": alias}).All(&tasks)
		if err != nil && err == mgo.ErrNotFound {
			return nil, taskNotFoundError
		} else if err != nil {
			return nil, err
		}
	}

	return tasks, nil
}

func (d *mongoDriver) Update(t Task) error {
	var c *mgo.Collection = d.getTasksC()
	err := c.UpdateId(t.ID, t)
	if err == mgo.ErrNotFound {
		return taskNotFoundError
	}
	return err
}

func (d *mongoDriver) Delete(t Task) error {
	var c = d.getTasksC()
	return c.RemoveId(bson.ObjectIdHex(t.ID.(string)))
}

func (d *mongoDriver) Close() error {
	d.session.Close()
	return nil
}

func (d *mongoDriver) getTasksC() *mgo.Collection {
	return d.session.DB(d.database).C("Tasks")
}
