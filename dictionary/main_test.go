package main

import (
	"errors"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/Clever/leakybucket/memory"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

type mockDao struct{}

func createRequest(method, path string, reader io.Reader, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, path, reader)
	if err != nil {
		t.Error(err.Error())
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req
}

func TestAddWord(t *testing.T) {
	println("TestAddWord")
	e := echo.New()
	// Test positive scenario
	var req = createRequest("echo.POST", "/word", strings.NewReader("{\"Word\" : \"Cat\", \"Translations\" : [\"Chat\"], \"Idioms\": []}"), t)
	wordsDb = new(mockDao)
	recorder := httptest.NewRecorder()
	cont := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
	err := addWord(cont)
	if err != nil {
		t.Error(err.Error())
	}
	if recorder.Code != http.StatusCreated {
		t.Error("Wrong response code")
	}
	if recorder.Body.Len() != 0 {
		t.Error("Content is not expected")
	}
	// Test negative scenario
	req = createRequest("echo.POST", "/word", strings.NewReader("{\"Word\" : \"Bomb\", \"Translations\" : [\"Bomba\"], \"Idioms\": []}"), t)
	recorder = httptest.NewRecorder()
	cont = e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
	err = addWord(cont)
	if err != nil {
		t.Error(err.Error())
	}
	if recorder.Code != http.StatusBadRequest {
		t.Error("Wrong response code")
	}
	responseBody, _ := recorder.Body.ReadString('\n')
	println(responseBody)
	if responseBody != "{\"error\":\"Word already exists\"}" {
		t.Error("Content is wrong")
	}
}

func TestUpdateWord(t *testing.T) {
	println("TestUpdateWord")
	e := echo.New()
	// Test positive scenario
	var req = createRequest("echo.PUT", "/word", strings.NewReader("{\"Word\" : \"Cat\", \"Translations\" : [\"Chat\"], \"Idioms\": []}"), t)
	wordsDb = new(mockDao)
	recorder := httptest.NewRecorder()
	cont := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
	err := updateWord(cont)
	if err != nil {
		t.Error(err.Error())
	}
	if recorder.Code != http.StatusOK {
		t.Error("Wrong response code")
	}
	if recorder.Body.Len() != 0 {
		t.Error("Content is not expected")
	}
}

func TestDeleteWord(t *testing.T) {
	println("TestDeleteWord")
	e := echo.New()
	// Test positive scenario
	wordsDb = new(mockDao)
	var req = createRequest("echo.DELETE", "/word/Hello", nil, t)
	recorder := httptest.NewRecorder()
	cont := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
	cont.SetParamNames("w")
	cont.SetParamValues("Hello")
	err := deleteWord(cont)
	if err != nil {
		t.Error(err.Error())
	}
	if recorder.Code != http.StatusNoContent {
		t.Error("Wrong response code")
	}
	if recorder.Body.Len() != 0 {
		t.Error("Content is not expected")
	}
	// Test negative scenario
	req = createRequest("echo.DELETE", "/word/Bomb", nil, t)
	recorder = httptest.NewRecorder()
	cont = e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
	cont.SetParamNames("w")
	cont.SetParamValues("Bomb")
	err = deleteWord(cont)
	if err != nil {
		t.Error(err.Error())
	}
	if recorder.Code != http.StatusBadRequest {
		t.Error("Wrong response code")
	}
	responseBody, _ := recorder.Body.ReadString('\n')
	if responseBody != "{\"error\":\"Word does not exist\"}" {
		t.Error("Content is wrong")
	}
}

func TestFindWord(t *testing.T) {
	println("TestFindWord")
	e := echo.New()
	wordsDb = new(mockDao)
	// Test positive scenario
	var req = createRequest("echo.GET", "/word/Hello", nil, t)
	recorder := httptest.NewRecorder()
	cont := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
	cont.SetParamNames("w")
	cont.SetParamValues("Hello")
	err := findWord(cont)
	if err != nil {
		t.Error(err.Error())
	}
	if recorder.Code != http.StatusOK {
		t.Error("Wrong response code")
	}
	responseBody, _ := recorder.Body.ReadString('\n')
	if responseBody != "{\"Word\":\"Hello\",\"Translations\":[\"Salut\"],\"Idioms\":[]}" {
		t.Error("Content is wrong")
	}
	// Test negative scenario
	req = createRequest("echo.GET", "/word/Bomb", nil, t)
	req.Header.Add("User", "Bill")
	recorder = httptest.NewRecorder()
	cont = e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
	cont.SetParamNames("w")
	cont.SetParamValues("Bomb")
	err = findWord(cont)
	if err != nil {
		t.Error(err.Error())
	}
	if recorder.Code != http.StatusBadRequest {
		t.Error("Wrong response code")
	}
	responseBody, _ = recorder.Body.ReadString('\n')
	if responseBody != "{\"error\":\"Word does not exist\"}" {
		t.Error("Content is wrong")
	}
}

func TestRateLimiter(t *testing.T) {
	println("TestRateLimiter")
	e := echo.New()
	wordsDb = new(mockDao)
	var storage = memory.New()
	var filter, err = storage.Create("Request Filter", 5, time.Second*5)
	if err != nil {
		log.Fatal("Cannot create rate limiter", err)
	}
	var responseCodes [7]int
	for i := 0; i < 7; i++ {
		req, err := http.NewRequest("echo.POST", "/word", strings.NewReader("{\"Word\" : \"Cat\", \"Translations\" : [\"Chat\"], \"Idioms\": []}"))
		if err != nil {
			t.Error(err.Error())
		}
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()
		cont := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
		err = limitRequest(filter)(addWord)(cont)
		if err != nil {
			t.Error(err.Error())
		}
		responseCodes[i] = recorder.Code
	}
	var i int
	for i = 0; i < 5; i++ {
		if responseCodes[i] != http.StatusCreated {
			t.Error("Wrong error code. Expected 201")
		}
	}
	for i = 5; i < 7; i++ {
		if responseCodes[i] != http.StatusServiceUnavailable {
			t.Error("Wrong error code. Expected 503")
		}
	}
}

func (di *mockDao) addDictEntry(user string, dictEntry dictionaryEntry) error {
	if dictEntry.Word == "Bomb" {
		return errors.New("Word already exists")
	}
	return nil
}

func (di *mockDao) updateDictEntry(user string, dictEntry dictionaryEntry) error {
	return nil
}

func (di *mockDao) deleteDictEntry(user string, dictEntry dictionaryEntry) error {
	if dictEntry.Word == "Bomb" {
		return errors.New("Word does not exist")
	}
	return nil
}

func (di *mockDao) checkTranslation(user, word, translation string) (bool, error) {
	return true, nil
}

func (di *mockDao) getAllWords(user string) ([]string, error) {
	return []string{"Hello", "World"}, nil
}

func (di *mockDao) getDictEntry(user, word string) (*dictionaryEntry, error) {
	if word == "Bomb" && user == "Bill" {
		return nil, errors.New("Word does not exist")
	}
	var entry = newDictEntry()
	entry.Word = "Hello"
	entry.Translations = append(entry.Translations, "Salut")
	return entry, nil
}
