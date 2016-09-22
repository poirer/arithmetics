package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type (
	responseField struct {
		name      string
		typ       string
		nestedObj []responseField
	}

	formParam struct {
		Key   string `xml:"name"`
		Value string `xml:"value"`
	}

	reqAddr struct {
		URL    string `xml:",chardata"`
		Method string `xml:"method,attr"`
	}

	apiDefList struct {
		DefList []apiCallDef `xml:"api-def"`
	}

	apiCallDef struct {
		RequestBodyTpl       string      `xml:"request-body-template"`
		Address              *reqAddr    `xml:"address"` // Declare it as a pointer is the only way to trim values using reflection, that I have found
		ResponseBodyTpl      string      `xml:"response-body-template"`
		SkipRespVerification bool        `xml:"skip-response-verification"`
		Status               int         `xml:"status"`
		Params               []formParam `xml:"parameters>param"`
	}
)

func checkEndpoint(def apiCallDef) error {
	var client = http.Client{}
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
	response, err := client.Do(request)
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

func logCall(def apiCallDef, err error) {
	log.Println("***** Call *****") // Fortran? :-)
	log.Printf("%s\t%s\n", def.Address.Method, def.Address.URL)
	if err != nil {
		log.Println("Failed: ", err.Error())
	} else {
		log.Println("Success")
	}
}

func main() {
	generator = randValueGenerator{}
	defs, err := readAPICallDefinitions("endpoints.xml")
	if err != nil {
		println("Error occurred: ", err.Error())
	} else {
		for _, d := range defs {
			err := checkEndpoint(d)
			logCall(d, err)
		}
	}
}
