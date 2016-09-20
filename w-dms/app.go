package main

import (
	"errors"
	"log"
	"net/http"

	"flag"

	_ "github.com/mattn/go-sqlite3"
)

// Task is task
// swagger:model
type Task struct {
	ID           interface{} `json:"id" bson:"_id,omitempty"`
	Alias        string      `json:"alias" bson:"alias"`
	Description  string      `json:"desc"`
	Type         string      `json:"type"`
	Tags         []string    `json:"tags"`
	Timestamp    int64       `json:"ts"`
	EstimateTime string      `json:"etime"`
	RealTime     string      `json:"rtime"`
	Reminders    []string    `json:"reminders"`
}

var (
	errTaskNotFound = errors.New("Task not found")
	errUnknownURL   = errors.New("Not found")
	errMissingID    = errors.New("Task id is missing")
	errInvalidID    = errors.New("Invalid task id")
)

type (
	// TaskList is tasklist
	TaskList []Task

	dbDriver interface {
		Create(t Task) error
		ReadByID(id interface{}) (TaskList, error)
		ReadByAlias(alias *string) (TaskList, error)
		Update(t Task) error
		Delete(t Task) error
	}

	closeable interface {
		Close() error
	}

	closeableDbDriver interface {
		dbDriver
		closeable
	}
)

var db closeableDbDriver // Yes, I remember that global variable is bad. But I don't know yet what is good

func main() {
	var dbType *string
	dbType = flag.String("db", "mongo", "Database type, either mongo or sqlite")
	flag.Parse()
	switch *dbType {
	case "mongo":
		log.Println("Using mongodb")
		db = newMongoDriver("Tasks")
	case "sqlite":
		log.Println("Using sqlite")
		db = newSqliteDriver("tasks.db")
	default:
		log.Println("Using default (mongo)")
		db = newMongoDriver("Tasks")
	}
	defer db.Close()
	var sm = http.NewServeMux()
	sm.HandleFunc("/task", dispatchTaskRequest)
	swaggerDoc := http.FileServer(http.Dir("swagger-ui"))
	sm.Handle("/api-doc/", http.StripPrefix("/api-doc/", swaggerDoc))
	http.ListenAndServe(":8080", sm)
}
