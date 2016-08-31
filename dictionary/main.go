package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
)

type dictionaryEntry struct {
	Word         string
	Translations []string
	Idioms       []string
}

var db *sql.DB

// Creates new instance of dictionaryEntry and initializes its fields
func newDictEntry() *dictionaryEntry {
	var de = new(dictionaryEntry)
	de.Idioms = make([]string, 0, 10)
	de.Translations = make([]string, 0, 10)
	return de
}

func addWord(c echo.Context) error {
	var dictEntry = newDictEntry()
	c.Bind(dictEntry)
	var user = c.Request().Header().Get("User")
	err := addDictEntry(db, user, *dictEntry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("{\"error\" : \"%s\"}", err.Error()))
	}
	return c.NoContent(http.StatusCreated)
}

func updateWord(c echo.Context) error {
	var dictEntry = newDictEntry()
	c.Bind(dictEntry)
	var user = c.Request().Header().Get("User")
	err := updateDictEntry(db, user, *dictEntry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("{\"error\" : \"%s\"}", err.Error()))
	}
	return c.NoContent(http.StatusOK)
}

func deleteWord(c echo.Context) error {
	var word = c.Param("w")
	var user = c.Request().Header().Get("User")
	var de = dictionaryEntry{word, []string{}, []string{}}
	err := deleteDictEntry(db, user, de)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("{\"error\" : \"%s\"}", err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func findWord(c echo.Context) error {
	var word = c.Param("w")
	var user = c.Request().Header().Get("User")
	de, err := getDictEntry(db, user, word)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("{\"error\" : \"%s\"}", err.Error()))
	}
	return c.JSON(http.StatusOK, de)
}

func loadAllWords(c echo.Context) error {
	var user = c.Request().Header().Get("User")
	words, err := getAllWords(db, user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("{\"error\" : \"%s\"}", err.Error()))
	}
	return c.JSON(http.StatusOK, words)
}

func getUsers(c echo.Context) error {
	users, err := retrieveUsers(db)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("{\"error\" : \"%s\"}", err.Error()))
	}
	return c.JSON(http.StatusOK, users)
}

func main() {
	db = connect("/home/zhenya/Development/task-data/words.db")
	defer db.Close()
	var mainEndpoint *echo.Echo = echo.New()
	mainEndpoint.POST("/word", addWord)
	mainEndpoint.PUT("/word", updateWord)
	mainEndpoint.DELETE("/word/:w", deleteWord)
	mainEndpoint.GET("/word/:w", findWord)
	mainEndpoint.GET("/word", loadAllWords)
	go mainEndpoint.Run(standard.New(":8083"))
	var secureEndpoint = echo.New()
	secureEndpoint.GET("/user", getUsers)
	var conf engine.Config = engine.Config{
		Address:      ":8483",
		TLSCertFile:  "cert.pem",
		TLSKeyFile:   "key.pem",
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}
	secureEndpoint.Run(standard.WithConfig(conf))
}
