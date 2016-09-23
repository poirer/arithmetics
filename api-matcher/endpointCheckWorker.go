package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type checkWorker struct {
	client    http.Client
	workQueue chan apiCallDef
	waitGroup *sync.WaitGroup
}

func startNewWorker(queue chan apiCallDef, waitGroup *sync.WaitGroup) {
	var client = http.Client{}
	var cw = checkWorker{client, queue, waitGroup}
	go cw.run()
}

func (cw *checkWorker) checkEndpoint(def apiCallDef) error {
	var body io.Reader
	if def.RequestBodyTpl != "" {
		content, err := composeRequestBody(def.RequestBodyTpl)
		if err != nil {
			println("Error occurred when trying to compose request body: ", err.Error())
			return err
		}
		body = strings.NewReader(content)
	}
	request, err := http.NewRequest(def.Address.Method, def.Address.URL, body)
	if err != nil {
		println("Error occurred when trying to create new request: ", err.Error())
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	if def.Params != nil {
		parameters := url.Values{}
		for _, p := range def.Params {
			parameters.Add(p.Key, p.Value)
		}
		request.Form = parameters
	}
	response, err := cw.client.Do(request)
	if err != nil {
		println("Error occurred when trying to perform request: ", err.Error())
		return err
	}
	if response.StatusCode != def.Status {
		return fmt.Errorf("Response status (%d) code is not equal to expected (%d)", response.StatusCode, def.Status)
	}
	if !def.SkipRespVerification {
		bodyContent, err := ioutil.ReadAll(response.Body)
		if err != nil {
			println("Error occurred while trying to read response")
			return err
		}
		if len(def.ResponseBodyTpl) > 0 && len(bodyContent) == 0 {
			return errors.New("Response body expected but was not received")
		}
		if len(def.ResponseBodyTpl) == 0 && len(bodyContent) > 0 {
			return errors.New("Response body not expected but was received")
		}
		expectedFields, err := parseResponseFields(def.ResponseBodyTpl)
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

func (checkWorker) logCall(def apiCallDef, err error) {
	var logBuf bytes.Buffer                  // It is needed to orint the whole log entry as one message; otherwise logs from different workers may be mixed since they share one instance of log
	logBuf.WriteString("***** Call *****\n") // Fortran? :-)
	logBuf.WriteString(fmt.Sprintf("%s\t%s\n", def.Address.Method, def.Address.URL))
	if err != nil {
		logBuf.WriteString("Failed: ")
		logBuf.WriteString(err.Error())
		logBuf.WriteByte('\n')
	} else {
		logBuf.WriteString("Success\n")
	}
	log.Print(logBuf.String())
}

func (cw *checkWorker) done() {
	cw.waitGroup.Done()
}

func (cw *checkWorker) run() {
	for {
		def := <-cw.workQueue
		err := cw.checkEndpoint(def)
		cw.logCall(def, err)
		cw.done()
	}
}
