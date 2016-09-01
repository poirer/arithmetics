package main

import (
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

func TestAddWord(t *testing.T) {
	e := echo.New()
	req, err := http.NewRequest("echo.POST", "/word", strings.NewReader("{\"Word\" : \"Cat\", \"Translations\" : [\"Chat\"], \"Idioms\": []}"))
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	db = new(mockDao)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	cont := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
	err = addWord(cont)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if recorder.Code != http.StatusCreated {
		t.Error("Wrong response code")
		t.Fail()
	}
	if recorder.Body.Len() != 0 {
		t.Error("Content is not expected")
		t.Fail()
	}
}

func TestRateLimiter(t *testing.T) {
	e := echo.New()
	db = new(mockDao)
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
			t.Fail()
		}
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()
		cont := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(recorder, e.Logger()))
		err = limitRequest(filter)(addWord)(cont)
		if err != nil {
			t.Error(err.Error())
			t.Fail()
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

func (di *mockDao) connect(dbURL string) {
}

func (di *mockDao) close() error {
	return nil
}

func (di *mockDao) addDictEntry(user string, dictEntry dictionaryEntry) error {
	return nil
}

func (di *mockDao) updateDictEntry(user string, dictEntry dictionaryEntry) error {
	return nil
}

func (di *mockDao) deleteDictEntry(user string, dictEntry dictionaryEntry) error {
	return nil
}

func (di *mockDao) checkTranslation(user, word, translation string) (bool, error) {
	return true, nil
}

func (di *mockDao) getAllWords(user string) ([]string, error) {
	return []string{"Hello", "World"}, nil
}

func (di *mockDao) getDictEntry(user, word string) (*dictionaryEntry, error) {
	var entry = newDictEntry()
	entry.Word = "Hello"
	entry.Translations = append(entry.Translations, "Salut")
	return entry, nil
}

func (di *mockDao) retrieveUsers() ([]string, error) {
	return []string{"zhenya", "gouser"}, nil
}
