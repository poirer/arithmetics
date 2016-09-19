// Package classification Task API.
//
// Provides API to operate with tasks
//
// swagger:meta
package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

// Task is task
// swagger:model
type Task struct {
	ID interface{} `json:"id" bson:"_id,omitempty"`
	Alias string `json:"alias" bson:"alias"`
	Description string `json:"desc"`
	Type         string   `json:"type"`
	Tags         []string `json:"tags"`
	Timestamp    int64    `json:"ts"`
	EstimateTime string   `json:"etime"`
	RealTime     string   `json:"rtime"`
	Reminders    []string `json:"reminders"`
}

type SwaggerTask struct {
	//Task Id
	ID string `json:"id" bson:"_id,omitempty"`
	// Task Alias
	// required: true
	Alias string `json:"alias" bson:"alias"`
	// Task description
	Description string `json:"desc"`
	// Task type
	// required: true
	Type string `json:"type"`
	// Task tags
	// required: false
	Tags []string `json:"tags"`
	// Timastamp when the task was created
	Timestamp int64 `json:"ts"`
	// Task estimation time
	EstimateTime string `json:"etime"`
	// Task actual time
	RealTime string `json:"rtime"`
	// Times when the task should be reminded
	Reminders []string `json:"reminders"`
}

// Tasks were read successfully
// swagger:response taskResponse
type TaskSwaggerResponse struct {
	//in: body
	Body []SwaggerTask
}

// Query params
//swagger:parameters readTasks deleteTask
type ReadTaskParameters struct {
	// Id of the task to be found (skip to find all)
	//
	//in: query
	//required: false
	Id string `json:"id"`

	// Alias of the task to be found (skip to find all)
	// in: query
	//required: false
	Alias string `json:"alias"`
}

// Task
//swagger:parameters saveTask updateTask
type TaskSwaggerRequest struct {
	//in: body
	Body SwaggerTask
}

type TaskNotFoundError struct {
}

// Task was not found
//swagger:response taskNotFound
type SwaggerTaskNotFound struct {}

// Task created
//swagger:response taskCreated
type SwaggerTaskCreated struct {}

// Task updated
//swagger:response taskUpdated
type SwaggerTaskUpdated struct {}

// Task deleted
//swagger:response taskDeleted
type SwaggerTaskDeleted struct {}

var (
	taskNotFoundError = errors.New("Task not found")
	unknownURLError = errors.New("Not found")
	missingIdError = errors.New("Task id is missing")
	invalidIDError = errors.New("Invalid task id")
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

func closeCompletelyRequestBody(body *io.ReadCloser) {
	defer func() {
		err := (*body).Close()
		if err != nil {
			log.Println(err.Error())
		}
	}()
	_, err := ioutil.ReadAll(*body)
	if err != nil {
		log.Println(err.Error())
	}
}

func readTaskFromRequest(request *http.Request) (*Task, error) {
	defer closeCompletelyRequestBody(&request.Body)
	var decoder = json.NewDecoder(request.Body)
	var t Task
	err := decoder.Decode(&t)
	if err != nil && err != io.EOF {
		log.Println(err.Error())
		return nil, err
	}
	return &t, nil
}

func writeTaskListToResponse(tasks TaskList, respWriter http.ResponseWriter) error {
	var encoder = json.NewEncoder(respWriter)
	respWriter.Header().Add("Content-Type", "application/json")
	err := encoder.Encode(tasks)
	return err
}

func writeError(wr http.ResponseWriter, err error) {
	log.Println(err.Error())
	if err == taskNotFoundError || err == unknownURLError {
		wr.WriteHeader(http.StatusNotFound)
	} else if err == missingIdError || err == invalidIDError {
		wr.WriteHeader(http.StatusBadRequest)
	} else {
		wr.WriteHeader(http.StatusInternalServerError)
	}
	wr.Write([]byte(err.Error()))
}

func dispatchTaskRequest(respWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		ReadTask(respWriter, request)
	case http.MethodPost:
		AddTask(respWriter, request)
	case http.MethodDelete:
		deleteTask(respWriter, request)
	case http.MethodPut:
		updateTask(respWriter, request)
	default:
		writeError(respWriter, unknownURLError)
	}
}

// REST API to save a task
func AddTask(respWriter http.ResponseWriter, request *http.Request) {
	// swagger:route POST /task saveTask
	// Saves a task in the database
	// Consumes:
	// - application/json
	// Produces:
	// - No content
	// Responses:
	// 201: taskCreated

	task, err := readTaskFromRequest(request)
	if err != nil {
		writeError(respWriter, err)
		return
	}
	err = db.Create(*task)
	if err != nil {
		writeError(respWriter, err)
		return
	}
	respWriter.WriteHeader(http.StatusCreated)
}

// REAT API to read tasks
func ReadTask(respWriter http.ResponseWriter, request *http.Request) {
	// swagger:route GET /task readTasks
	// Reads tasks from database
	// Produces:
	// - application/json
	// Responses:
	// 200: taskResponse
	// 404: taskNotFound
	var id = request.FormValue("id")
	var alias = request.FormValue("alias")
	var result TaskList
	var err error
	if id != "" {
		result, err = db.ReadByID(&id)
	} else if alias != "" {
		result, err = db.ReadByAlias(&alias)
	} else {
		result, err = db.ReadByID(nil)
	}
	if err != nil {
		writeError(respWriter, err)
		return
	}
	writeTaskListToResponse(result, respWriter)
}

// REST API to delete task
func deleteTask(respWriter http.ResponseWriter, request *http.Request) {
	//swagger:route DELETE /task deleteTask
	// Deletes task from database
	// Responses:
	// 204: taskDeleted
	// 404: taskNotFound
	var id = request.FormValue("id")
	if id != "" {
		err := db.Delete(Task{ID: id})
		if err != nil {
			writeError(respWriter, err)
			return
		}
	} else {
		writeError(respWriter, missingIdError)
	}
	respWriter.WriteHeader(http.StatusNoContent)
}

// REST API to update task
func updateTask(respWriter http.ResponseWriter, request *http.Request) {
	//swagger:route PUT /task updateTask
	// Updates task in database
	// Consumes:
	// - application/json
	// Responses:
	// 200: taskUpdated
	// 404: taskNotFound
	task, err := readTaskFromRequest(request)
	if err != nil {
		writeError(respWriter, err)
		return
	}
	err = db.Update(*task)
	if err != nil {
		writeError(respWriter, err)
		return
	}
	respWriter.WriteHeader(http.StatusNoContent)
}

func main() {
	//db = newSqыliteDriver("/home/zhenya/Development/data/tasks.db")
	db = newMongoDriver("Tasks")
	defer db.Close()
	var sm = http.NewServeMux()
	sm.HandleFunc("/task", dispatchTaskRequest)
	swaggerDoc := http.FileServer(http.Dir("swagger-ui"))
	sm.Handle("/api-doc/", http.StripPrefix("/api-doc/", swaggerDoc))
	http.ListenAndServe(":8080", sm)
}
