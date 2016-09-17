package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const task = `{
  "id" : 2,
  "alias" : "Second Task",
  ""
}

{"eventType":"session_start","ts":1473837996,"params":{"first":1,"second":"Two"}}
`

// Task is task
type (
	Task struct {
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

	// TaskList is tasklist
	TaskList []Task

	dbDriver interface {
		Create(t Task) error
		ReadByID(id *int64) (TaskList, error)
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

func writeError(wr http.ResponseWriter, status int, err error) {
	log.Println(err.Error())
	wr.WriteHeader(status)
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
		writeError(respWriter, http.StatusNotFound, errors.New("Not found"))
	}
}

func addTask(respWriter http.ResponseWriter, request *http.Request) {
	task, err := readTaskFromRequest(request)
	if err != nil {
		writeError(respWriter, http.StatusInternalServerError, err)
		return
	}
	err = db.Create(*task)
	if err != nil {
		writeError(respWriter, http.StatusInternalServerError, err)
		return
	}
	respWriter.WriteHeader(http.StatusCreated)
}

func readTask(respWriter http.ResponseWriter, request *http.Request) {
	var sID = request.FormValue("id")
	var alias = request.FormValue("alias")
	var result TaskList
	var err error
	if sID != "" {
		id, parseErr := strconv.ParseInt(sID, 10, 0)
		if parseErr != nil {
			writeError(respWriter, http.StatusBadRequest, parseErr)
			return
		}
		result, err = db.ReadByID(&id)
	} else if alias != "" {
		result, err = db.ReadByAlias(&alias)
	} else {
		result, err = db.ReadByID(nil)
	}
	if err != nil {
		writeError(respWriter, http.StatusInternalServerError, err)
		return
	}
	writeTaskListToResponse(result, respWriter)
}

func deleteTask(respWriter http.ResponseWriter, request *http.Request) {
	var sID = request.FormValue("id")
	if sID != "" {
		id, parseErr := strconv.ParseInt(sID, 10, 0)
		if parseErr != nil {
			writeError(respWriter, http.StatusBadRequest, parseErr)
			return
		}
		err := db.Delete(Task{ID: int(id)})
		if err != nil {
			writeError(respWriter, http.StatusInternalServerError, err)
			return
		}
	} else {
		writeError(respWriter, http.StatusBadRequest, errors.New("Task id is not specidied"))
	}
	respWriter.WriteHeader(http.StatusNoContent)
}

func updateTask(respWriter http.ResponseWriter, request *http.Request) {
	task, err := readTaskFromRequest(request)
	if err != nil {
		writeError(respWriter, http.StatusInternalServerError, err)
		return
	}
	err = db.Update(*task)
	if err != nil {
		writeError(respWriter, http.StatusInternalServerError, err)
		return
	}
	respWriter.WriteHeader(http.StatusNoContent)
}

func main() {
	//db = newSqliteDriver("/home/zhenya/Development/data/tasks.db")
	db = newMongoDriver("Tasks")
	defer db.Close()
	var sm = http.NewServeMux()
	sm.HandleFunc("/task", dispatchTaskRequest)
	http.ListenAndServe(":8080", sm)
}
