package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	taskSimpleJSONSpec = `{
    "alias": string,
    "desc": string,
    "etime": string,
    "reminders": [string],
    "rtime": string,
    "tags": [string],
    "ts": int,
    "type": string
  }`

	taskJSONSpec = `{
    "alias": string,
    "desc": string,
    "inner" : {
      "f1":string,
      "f2":[int],
      "f3": {"nf1":string,"nf2":string}
    },
    "etime": string,
    "reminders": [string],
    "rtime": string,
    "tags": [string],
    "ts": int,
    "type": string,
    "done": bool
  }`
)

type responseField struct {
	name      string
	typ       string
	nestedObj []responseField
}

type apiCallDef struct {
	requestBodyTmpl          string
	method                   string
	url                      string
	responseBodyTmpl         string
	skipResponseVerification bool
	status                   int
}

func checkEndpoint(def apiCallDef) error {
	var client = http.Client{}
	var body io.Reader
	if def.requestBodyTmpl != "" {
		content, err := composeRequestBody(def.requestBodyTmpl)
		if err != nil {
			println("Error occurred when trying to compose request body: ", err.Error())
			return err
		}
		body = strings.NewReader(content)
	}
	request, err := http.NewRequest(def.method, def.url, body)
	if err != nil {
		println("Error occurred when trying to create new request: ", err.Error())
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		println("Error occurred when trying to perform request: ", err.Error())
		return err
	}
	if response.StatusCode != def.status {
		return errors.New("Response status code is not equal to expected")
	}
	if !def.skipResponseVerification {
		bodyContent, err := ioutil.ReadAll(response.Body)
		if len(def.responseBodyTmpl) > 0 && len(bodyContent) == 0 {
			return errors.New("Response body expected but was not received")
		}
		if len(def.responseBodyTmpl) == 0 && len(bodyContent) > 0 {
			return errors.New("Response body not expected but was received")
		}
		expectedFields, err := parseResponseFields(def.responseBodyTmpl)
		if err != nil {
			println("Error occurred when trying to parse expected fields: ", err.Error())
			return err
		}
		correct, err := checkResponseAgainstExpectedFields(string(bodyContent), expectedFields)
		if err != nil {
			println("Error occurred when verifying response body: ", err.Error())
			return err
		}
		if !correct {
			return errors.New("Server response does not match to expectation")
		}
	}
	return nil
}

func main() {
	var apiDef = apiCallDef{
		requestBodyTmpl: taskSimpleJSONSpec,
		status:          201,
		skipResponseVerification: true,
		method: http.MethodPost,
		url:    "http://localhost:8080/task",
	}
	res := checkEndpoint(apiDef)
	if res != nil {
		println("Unseccessful ", res.Error())
	} else {
		println("Successful")
	}
}
