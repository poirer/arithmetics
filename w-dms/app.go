package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

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
		ID           int      `json:"id"`
		Alias        string   `json:"alias"`
		Description  string   `json:"desc"`
		Type         string   `json:"type"`
		Tags         []string `json:"tags"`
		Timestamp    int64    `json:"ts"`
		EstimateTime string   `json:"etime"`
		RealTime     string   `json:"rtime"`
		Reminders    []string `json:"reminders"`
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
)

var db dbDriver // Yes, I temeber that global variable is bad. But I don't know yet what is good

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

func addTask(respWriter http.ResponseWriter, request *http.Request) {
	task, err := readTaskFromRequest(request)
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		respWriter.Write([]byte(err.Error()))
	}
	err = db.Create(*task)
	if err != nil {
		log.Println(err.Error())
		respWriter.WriteHeader(http.StatusInternalServerError)
		respWriter.Write([]byte(err.Error()))
	}
	var res = TaskList{*task}
	writeTaskListToResponse(res, respWriter)
}

func main() {
	db = newSqliteDriver("/home/zhenya/Development/data/tasks.db")
	var sm = http.NewServeMux()
	sm.HandleFunc("/add", addTask)
	http.ListenAndServe(":8080", sm)
}
