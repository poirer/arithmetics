// Application publishes REST API to perform basic arithmetic operations using different HTTP methods
// It uses http.ServeMux and middleware. There is no special tests, standart REST client was used to verify
package main

import (
	"net/http"
	"strconv"
	"os"
	"fmt"
)

var (
	accessLogFile *os.File
)

func init() {
	err := os.MkdirAll("logs", 0777)
	if err != nil {
		println("Cannot create folder for logs")
		panic(err)
	}
	accessLogFile, err = os.Create("logs/access.log")
	if err != nil {
		println("Cannot create access log file")
	}
}


func add(respWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPost {
		doOperation(respWriter, request, "+")
	} else {
		invalidMethod(respWriter, request)
	}
}

func substract(respWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodDelete {
		doOperation(respWriter, request, "-")
	} else {
		invalidMethod(respWriter, request)
	}
}

func multiply(respWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut {
		doOperation(respWriter, request, "*")
	} else {
		invalidMethod(respWriter, request)
	}
}

func divide(respWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		doOperation(respWriter, request, "/")
	} else {
		invalidMethod(respWriter, request)
	}
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

func invalidMethod(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.WriteHeader(http.StatusMethodNotAllowed)
	respWriter.Write([]byte("Method " + request.Method + " is not allowed for this path"))
}

func logMiddlewareFunc(mainHandler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func (respWriter http.ResponseWriter, request *http.Request) {
		accessLogFile.WriteString(fmt.Sprintf("Method: %s, path: %s\n", request.Method, request.URL.Path))
		mainHandler.ServeHTTP(respWriter, request)
	})
}

func main() {
	var serverMux = http.NewServeMux()
	serverMux.HandleFunc("/add", logMiddlewareFunc(add))
	serverMux.HandleFunc("/substract", logMiddlewareFunc(substract))
	serverMux.HandleFunc("/divide", logMiddlewareFunc(divide))
	serverMux.HandleFunc("/multiply", logMiddlewareFunc(multiply))
	err := http.ListenAndServe(":8083", serverMux)
	if err != nil {
		println(err.Error())
	}
}
