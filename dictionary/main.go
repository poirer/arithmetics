package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"net/http"
)

type dictionaryEntry struct {
	Word         string
	Translations []string
	Idioms       []string
}

var db *sql.DB

// Creates new instance of dictionaryEntry and initializes its fields
func NewDictEntry() *dictionaryEntry {
	var de = new(dictionaryEntry)
	de.Idioms = make([]string, 0, 10)
	de.Translations = make([]string, 0, 10)
	return de
}

func addWord(c echo.Context) error {
	var dictEntry = NewDictEntry()
	c.Bind(dictEntry)
	var user = c.Request().Header().Get("User")
	err := addDictEntry(db, user, *dictEntry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("{\"error\" : \"%s\"}", err.Error()))
	}
	return c.NoContent(http.StatusCreated)
}

func updateWord(c echo.Context) error {
	var dictEntry = NewDictEntry()
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

func main() {
	db = connect("/home/zhenya/Development/go-test-data/dict")
	defer db.Close()
	var e *echo.Echo = echo.New()
	e.POST("/word", addWord)
	e.PUT("/word", updateWord)
	e.DELETE("/word/:w", deleteWord)
	e.GET("/word/:w", findWord)
	e.GET("/word", loadAllWords)
	e.Run(standard.New(":8083"))
}
