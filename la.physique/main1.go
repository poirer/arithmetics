// Application publishes REST API to perform basic arithmetic operations using different HTTP methods
// It uses http.ServeMux (why it is named Mux?)
package main

import (
	"net/http"
	"strconv"
)

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
	}
	y, err := strconv.ParseInt(yStr, 0, 0)
	if err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		respWriter.Write([]byte("Invalid argument Y"))
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

func main() {
	// Question: Don't know yet how to add pre-handler to log all requests...
	var serverMux = http.NewServeMux()
	serverMux.HandleFunc("/add", add)
	serverMux.HandleFunc("/substract", substract)
	serverMux.HandleFunc("/divide", divide)
	serverMux.HandleFunc("/multiply", multiply)
	err := http.ListenAndServe(":8083", serverMux)
	if err != nil {
		println(err.Error())
	}
	accessLogFile.Close()
}
