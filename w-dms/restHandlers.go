// Package classification Task API.
//
// Provides API to operate with tasks
//
// swagger:meta
package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type swaggerTask struct {
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
type swaggerTaskResponse struct {
	//in: body
	Body []swaggerTask
}

// Query params
//swagger:parameters readTasks deleteTask
type swaggerReadTaskParameters struct {
	// Id of the task to be found (skip to find all)
	//
	//in: query
	//required: false
	ID string `json:"id"`

	// Alias of the task to be found (skip to find all)
	// in: query
	//required: false
	Alias string `json:"alias"`
}

// Task
//swagger:parameters saveTask updateTask
type swaggerTaskRequest struct {
	//in: body
	Body swaggerTask
}

// Task was not found
//swagger:response taskNotFound
type swaggerTaskNotFound struct{}

// Task created
//swagger:response taskCreated
type swaggerTaskCreated struct{}

// Task updated
//swagger:response taskUpdated
type swaggerTaskUpdated struct{}

// Task deleted
//swagger:response taskDeleted
type swaggerTaskDeleted struct{}

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
	if err == errTaskNotFound || err == errUnknownURL {
		wr.WriteHeader(http.StatusNotFound)
	} else if err == errMissingID || err == errInvalidID {
		wr.WriteHeader(http.StatusBadRequest)
	} else {
		wr.WriteHeader(http.StatusInternalServerError)
	}
	wr.Write([]byte(err.Error()))
}

func dispatchTaskRequest(respWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		readTask(respWriter, request)
	case http.MethodPost:
		addTask(respWriter, request)
	case http.MethodDelete:
		deleteTask(respWriter, request)
	case http.MethodPut:
		updateTask(respWriter, request)
	default:
		writeError(respWriter, errUnknownURL)
	}
}

// addTask is REST API to save a task
func addTask(respWriter http.ResponseWriter, request *http.Request) {
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

// readTask is REST API to read tasks
func readTask(respWriter http.ResponseWriter, request *http.Request) {
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
		writeError(respWriter, errMissingID)
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
