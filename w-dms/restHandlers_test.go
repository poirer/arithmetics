package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validNewTaskRequestBody = `{
  "alias": "Task 1",
  "desc": "Description 1",
  "etime": "1h",
  "reminders": ["10m", "11m"],
  "rtime": "1h 20m",
  "tags": ["golang", "test", "dms"],
  "ts": 1234567890,
  "type": "normal"
  }`

	invalidNewTaskRequestBody = `{
  "alias": "Raise Error",
  "desc": "Description 1",
  "etime": "1h",
  "reminders": ["10m", "11m"],
  "rtime": "1h 20m",
  "tags": ["golang", "test", "dms"],
  "ts": 1234567890,
  "type": "normal"
  }`
)

var (
	emptyParams = make(map[string]string, 0)

	validNewTask = Task{
		Alias:        "Task 1",
		Description:  "Description 1",
		EstimateTime: "1h",
		Reminders:    []string{"10m", "11m"},
		RealTime:     "1h 20m",
		Tags:         []string{"golang", "test", "dms"},
		Timestamp:    1234567890,
		Type:         "normal",
	}

	invalidNewTask = Task{
		Alias:        "Raise Error",
		Description:  "Description 1",
		EstimateTime: "1h",
		Reminders:    []string{"10m", "11m"},
		RealTime:     "1h 20m",
		Tags:         []string{"golang", "test", "dms"},
		Timestamp:    1234567890,
		Type:         "normal",
	}
)

type dbDriverMock struct {
	mock.Mock
}

func (dm *dbDriverMock) Create(t Task) error {
	ar := dm.Called(t)
	return ar.Error(0)
}

func (dm *dbDriverMock) ReadByID(id interface{}) (TaskList, error) {
	ar := dm.Called(id)
	res := ar.Get(0)
	if res == nil {
		return nil, ar.Error(1)
	}
	return res.(TaskList), ar.Error(1)
}

func (dm *dbDriverMock) ReadByAlias(alias *string) (TaskList, error) {
	ar := dm.Called(alias)
	return ar.Get(0).(TaskList), ar.Error(1)
}

func (dm *dbDriverMock) Update(t Task) error {
	ar := dm.Called(t)
	return ar.Error(0)
}

func (dm *dbDriverMock) Delete(t Task) error {
	ar := dm.Called(t)
	return ar.Error(0)
}

func (dm *dbDriverMock) Close() error {
	return nil
}

func (dm *dbDriverMock) Init() error {
	return nil
}

func createRequest(method, body string, parameters map[string]string) (*http.Request, error) {
	var reader io.Reader = strings.NewReader(body)
	var form = make(url.Values)
	for p := range parameters {
		form[p] = []string{parameters[p]}
	}
	req, err := http.NewRequest(method, "/task", reader)
	req.Form = form
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func prepareRequest(t *testing.T, method, body string, parameters map[string]string) (*http.Request, *httptest.ResponseRecorder) {
	request, err := createRequest(method, body, parameters)
	assert.Nil(t, err)
	recorder := httptest.NewRecorder()
	return request, recorder
}

var dbMock *dbDriverMock

func resetMock() {
	dbMock = new(dbDriverMock)
	db = dbMock
}

func TestAddTask(t *testing.T) {
	// Successful task creation
	resetMock()
	dbMock.On("Create", validNewTask).Return(nil)
	request, recorder := prepareRequest(t, http.MethodPost, validNewTaskRequestBody, emptyParams)
	addTask(recorder, request)
	assert.Equal(t, http.StatusCreated, recorder.Code, "Unexpected response status")
	assert.Equal(t, 0, recorder.Body.Len(), "Unexpected response length")
	dbMock.AssertExpectations(t)

	// Unseccessful task creation
	resetMock()
	dbMock.On("Create", invalidNewTask).Return(errors.New("Terrible error! We all will die!"))
	request, recorder = prepareRequest(t, http.MethodPost, invalidNewTaskRequestBody, emptyParams)
	addTask(recorder, request)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code, "Unexpected response status")
	responseBody, _ := recorder.Body.ReadString('\n')
	assert.Equal(t, "Terrible error! We all will die!", responseBody, "Unexpected error message")
	dbMock.AssertExpectations(t)

	// Invalid task in request
	request, recorder = prepareRequest(t, http.MethodPost, "Not a task", emptyParams)
	addTask(recorder, request)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code, "Unexpected response status")
	responseBody, _ = recorder.Body.ReadString('\n')
	assert.NotEmpty(t, responseBody, "Empty error message")
}

func TestReadTask(t *testing.T) {
	// Read all tasks when no parameters are specified
	resetMock()
	dbMock.On("ReadByID", nil).Return(TaskList{Task{ID: 1}, Task{ID: 2}}, nil)
	request, recorder := prepareRequest(t, http.MethodGet, "", emptyParams)
	readTask(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code, "Unexpected response status")
	// It would be better to check result here. But it is too annoying
	dbMock.AssertExpectations(t)

	// Read task with concrete id when only id is specified in query
	resetMock()
	var param = "1"
	dbMock.On("ReadByID", &param).Return(TaskList{Task{ID: 1}}, nil)
	request, recorder = prepareRequest(t, http.MethodGet, "", map[string]string{"id": "1"})
	readTask(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code, "Unexpected response status")
	// It would be better to check result here. But it is too annoying
	dbMock.AssertExpectations(t)

	// Read task with concrete alias when only alias is specified in query
	resetMock()
	param = "Task 1"
	dbMock.On("ReadByAlias", &param).Return(TaskList{Task{ID: 1}}, nil)
	request, recorder = prepareRequest(t, http.MethodGet, "", map[string]string{"alias": "Task 1"})
	readTask(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code, "Unexpected response status")
	// It would be better to check result here. But it is too annoying
	dbMock.AssertExpectations(t)

	// Read task with concrete id when both id and alias are specified in query
	resetMock()
	param = "1"
	dbMock.On("ReadByID", &param).Return(TaskList{Task{ID: 1}}, nil)
	request, recorder = prepareRequest(t, http.MethodGet, "", map[string]string{"alias": "Task 1", "id": "1"})
	readTask(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code, "Unexpected response status")
	// It would be better to check result here. But it is too annoying
	dbMock.AssertExpectations(t)
	dbMock.AssertNotCalled(t, "ReadByAlias")

	// Read task which does not exist
	resetMock()
	param = "3"
	dbMock.On("ReadByID", &param).Return(nil, errTaskNotFound)
	request, recorder = prepareRequest(t, http.MethodGet, "", map[string]string{"id": "3"})
	readTask(recorder, request)
	assert.Equal(t, recorder.Code, http.StatusNotFound, "Unexpected response status")
	responseBody, _ := recorder.Body.ReadString('\n')
	assert.Equal(t, "Task not found", responseBody)
	dbMock.AssertExpectations(t)
}
