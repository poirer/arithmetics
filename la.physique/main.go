// Application publishes REST API to perform basic arithmetic operations using different HTTP methods
// It uses http.Server and custom handler to dispatch requests. There is no special tests, standart REST client was used to verify
package main1

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	logFile       *os.File
	accessLogFile *os.File
)

func init() {
	err := os.MkdirAll("logs", 0777)
	if err != nil {
		println("Cannot create folder for logs")
		panic(err)
	}
	logFile, err = os.Create("logs/server.log")
	if err != nil {
		println("Cannot create log file")
		panic(err)
	}
	accessLogFile, err = os.Create("logs/access.log")
	if err != nil {
		println("Cannot create access log file")
	}
}

type handlerMap map[string]func(http.ResponseWriter, *http.Request)

type requestDispatcher struct {
	pathHandlerMap handlerMap
	defaultHandler func(http.ResponseWriter, *http.Request)
}

func (rd requestDispatcher) ServeHTTP(respWriter http.ResponseWriter, request *http.Request) {
	var path = request.URL.Path
	var method = request.Method
	accessLogFile.WriteString(fmt.Sprintf("Methd: %s, Path:%s\n", method, path))
	handlerFunc, exists := rd.pathHandlerMap[method+":"+path]
	if !exists {
		rd.defaultHandler(respWriter, request)
	} else {
		// Question: I would like to handle request in separate goroutine. But if I add key word go before the function call, it will stop to work :-(
		handlerFunc(respWriter, request)
	}
}

func (rd requestDispatcher) register(path string, handleFunc func(http.ResponseWriter, *http.Request)) {
	rd.pathHandlerMap[path] = handleFunc
}

func voidHandler(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.WriteHeader(http.StatusNotFound)
	respWriter.Write([]byte("Unknown request"))
}

func add(respWriter http.ResponseWriter, request *http.Request) {
	// var xStr = request.PostFormValue("x")
	doOperation(respWriter, request, "+")
}

func substract(respWriter http.ResponseWriter, request *http.Request) {
	doOperation(respWriter, request, "-")
}

func multiply(respWriter http.ResponseWriter, request *http.Request) {
	doOperation(respWriter, request, "*")
}

func divide(respWriter http.ResponseWriter, request *http.Request) {
	doOperation(respWriter, request, "/")
}

func doOperation(respWriter http.ResponseWriter, request *http.Request, operation string) {
	var xStr = request.FormValue("x")
	var yStr = request.FormValue("y")
	x, err := strconv.ParseInt(xStr, 0, 0)
	if err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		respWriter.Write([]byte("Invalid argument X"))
		return
	}
	y, err := strconv.ParseInt(yStr, 0, 0)
	if err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		respWriter.Write([]byte("Invalid argument Y"))
		return
	}
	var result int64
	switch operation {
	case "+":
		result = x + y
	case "-":
		result = x - y
	case "*":
		result = x * y
	case "/":
		result = x / y
	}
	respWriter.Write([]byte(strconv.FormatInt(result, 10)))
}

func main() {
	var dispatcher = requestDispatcher{
		pathHandlerMap: make(handlerMap, 100),
		defaultHandler: voidHandler,
	}
	dispatcher.register("POST:/add", add)
	dispatcher.register("DELETE:/substract", substract)
	dispatcher.register("PUT:/multiply", multiply)
	dispatcher.register("GET:/divide", divide)
	var server = &http.Server{
		Addr:         ":8083", // Question: how can I set some specific path for my server? For example, I want that <localhot:8083/A/path-and-parameters> and <localhost:8083/B/path-and-parameters> are being processed by different instances of http.Server. Addr: :8083/A didn't work :-(
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		ErrorLog:     log.New(logFile, "", 0),
		Handler:      dispatcher,
	}
	err := server.ListenAndServe()
	if err != nil {
		println(err.Error())
	}
	// Question: not sure it is the best way to close files. What if ListenAndServe panics? Or what if logFile was created successfully, but accessLogFile pnics? This code will never run then...
	logFile.Close()
	accessLogFile.Close()
}
