package main

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"

	meterBucket "github.com/Clever/leakybucket"
	"github.com/Clever/leakybucket/memory"

	flowBucket "github.com/imormo/leakyBucket"
	"fmt"
)

type (
	dictionaryEntry struct {
		Word         string
		Translations []string
		Idioms       []string
	}
	errorWrapper struct {
		Error string `json:"error"`
	}
)

var (
	wordsDb wordsDao
	userDb  userDao
	dbConn  connectable
)

func initDb() {
	var dao = new(daoImpl)
	wordsDb = dao
	userDb = dao
	dbConn = dao
	dbConn.connect("/home/zhenya/Development/task-data/words.db")
}

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
	err := wordsDb.addDictEntry(user, *dictEntry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorWrapper{err.Error()})
	}
	return c.NoContent(http.StatusCreated)
}

func updateWord(c echo.Context) error {
	var dictEntry = newDictEntry()
	c.Bind(dictEntry)
	var user = c.Request().Header().Get("User")
	err := wordsDb.updateDictEntry(user, *dictEntry)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorWrapper{err.Error()})
	}
	return c.NoContent(http.StatusOK)
}

func deleteWord(c echo.Context) error {
	var word = c.Param("w")
	var user = c.Request().Header().Get("User")
	var de = dictionaryEntry{word, []string{}, []string{}}
	err := wordsDb.deleteDictEntry(user, de)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorWrapper{err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func findWord(c echo.Context) error {
	var word = c.Param("w")
	var user = c.Request().Header().Get("User")
	de, err := wordsDb.getDictEntry(user, word)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorWrapper{err.Error()})
	}
	return c.JSON(http.StatusOK, de)
}

func loadAllWords(c echo.Context) error {
	var user = c.Request().Header().Get("User")
	words, err := wordsDb.getAllWords(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorWrapper{err.Error()})
	}
	return c.JSON(http.StatusOK, words)
}

func getUsers(c echo.Context) error {
	users, err := userDb.retrieveUsers()
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorWrapper{err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func truncateRequest(c echo.Context) error {
	return c.String(http.StatusServiceUnavailable, "Too many requests")
}

func limitRequest(filter leakybucket.Bucket) func(mainHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(mainHandler echo.HandlerFunc) echo.HandlerFunc {
		state, err := filter.Add(1)
		if err == leakybucket.ErrorFull {
			log.Println("Bucket is full", state)
			return truncateRequest
		} else if err != nil {
			log.Println("Unexpected error", err)
			return nil
		}
		return func(c echo.Context) error {
			return mainHandler(c)
		}
	}
}

var flowLeackyBucket flowBucket.LeakyBucket

func queueRequest(c echo.Context) error {
	err := flowLeackyBucket.Add(struct{})
	if err != nil {
		log.Println("Bucket is full")
		return c.String(http.StatusServiceUnavailable, "Too many requests")
	}
	return c.NoContent(http.StatusAccepted)
}

func flowBucketRequestHandler(m interface{}) (err error) {
	fmt.Println("Handled")
	return nil
}

func main() {
	initDb()
	defer dbConn.close()
	var storage = memory.New()
	var meterLeackyBucket, err = storage.Create("Request Filter", 5, time.Second*5)
	if err != nil {
		log.Fatal("Cannot create rate limiter", err)
	}
	flowLeackyBucket = flowBucket.NewLeakyBucket(10, 5, flowBucketRequestHandler)
	var mainEndpoint = echo.New()
	mainEndpoint.Use(limitRequest(meterLeackyBucket))
	mainEndpoint.POST("/word", addWord)
	mainEndpoint.PUT("/word", updateWord)
	mainEndpoint.DELETE("/word/:w", deleteWord)
	mainEndpoint.GET("/word/:w", findWord)
	mainEndpoint.GET("/word", loadAllWords)
	mainEndpoint.GET("/queue", queueRequest)
	mainEndpoint.Run(standard.New(":8083"))
}
