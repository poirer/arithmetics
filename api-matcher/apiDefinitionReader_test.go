package main

import (
	"strings"
	"testing"
)

const (
	defXML = `<definitions>
    <api-def>
      <address method="GET">
        http://localhost:8080/task
      </address>
      <request-body-template/>
      <status>200</status>
      <response format="text" skip-verification="true"/>
      <parameters>
        <param>
          <name>id</name>
          <value>0</value>
        </param>
        <param>
          <name>alias</name>
          <value>golang</value>
        </param>
      </parameters>
    </api-def>
    <api-def>
      <address method="DELETE">
        http://localhost:8080/task
      </address>
      <request-body-template>
        {"field1": [int], "field2": [string]}
      </request-body-template>
      <status>400</status>
			<response format="json" skip-verification="false">
				<template>
					{"f1": string, "f2": int, "f3": bool}
				</template>
			</response>
    </api-def>
  </definitions>
`
	invalidXML = `<definitions>
    <api>
      <address method="GET">
        http://localhost:8080/task
      </address>
      <parameters>
        <param>
          <name>id</name>
          <value>0</value>
        </param>
        <param>
          <name>alias</name>
          <value>golang</value>
        </param>
      </parameters>
    </api-def>
  </definitions>
`
)

func TestReading(t *testing.T) {
	var source = strings.NewReader(defXML)
	definitions, err := readDefinitionsFromReader(source)
	if err != nil {
		t.Error("Error occurred while reading definitions from Xml", err)
	}
	expextedDefs := []apiCallDef{
		{
			RequestBodyTpl: "",
			Address:        reqAddr{Method: "GET", URL: "http://localhost:8080/task"},
			Response:       response{"", true, "text"},
			Status:         200,
			Params:         []formParam{{Key: "id", Value: "0"}, {Key: "alias", Value: "golang"}},
		},
		{
			RequestBodyTpl: `{"field1": [int], "field2": [string]}`,
			Address:        reqAddr{Method: "DELETE", URL: "http://localhost:8080/task"},
			Response:       response{"{\"f1\": string, \"f2\": int, \"f3\": bool}", false, "json"},
			Status:         400,
			Params:         nil,
		},
	}
	if len(definitions) != len(expextedDefs) {
		t.Errorf("Wrong number of read definitions. Expected %d but was %d", len(expextedDefs), len(definitions))
		t.FailNow()
	}
	for i, d := range expextedDefs {
		compareAPIDefinitions(t, d, definitions[i])
	}
}

func TestReadFromInvalidXml(t *testing.T) {
	var source = strings.NewReader(invalidXML)
	_, err := readDefinitionsFromReader(source)
	if err == nil {
		t.Error("Error had to be raised, but was not")
	}
}

func TestObjectValuesTrim(t *testing.T) {
	var d = createObjectToTrim()
	trimAPIDefFileds(&d)
	if d.RequestBodyTpl != "Two\nlines" {
		t.Errorf("Wrong request body template after trim. Expected \"Two\nlines\" but was \"%s\"", d.RequestBodyTpl)
	}
	if d.Response.BodyTpl != "Test with		 tabulation" {
		t.Errorf("Wrong response body template after trim. Expected \"Test with		 tabulation\" but was \"%s\"", d.Response.BodyTpl)
	}
	if d.Response.Format != "text" {
		t.Errorf("Wrong response body template after trim. Expected \"text\" but was \"%s\"", d.Response.BodyTpl)
	}
	if d.Response.SkipVerification {
		t.Errorf("Wrong verification flag after trim. Expected false but was %t", d.Response.SkipVerification)
	}
	if d.Address.Method != "POST" {
		t.Errorf("Wrong address method after trim. Expected \"POST\" but was \"%s\"", d.Address.Method)
	}
	if d.Address.URL != "localhost" {
		t.Errorf("Wrong address url after trim. Expected \"localhost\" but was \"%s\"", d.Address.URL)
	}
	if d.Status != 100 {
		t.Errorf("Wrong status after trim. Expected 100 but was %d", d.Status)
	}
	var expectedParams = []formParam{
		{"Name1", "Value1"},
		{"Name2", "Value 2"},
		{"Name3", "Value 3"},
		{"", ""},
	}
	if len(d.Params) != 4 {
		t.Errorf("Wrong number of parameters after trim. Expected 4 but was %d", len(d.Params))
	}
	for i, p := range d.Params {
		if p.Key != expectedParams[i].Key {
			t.Errorf("Wrong parameter key after trim. Expected <%s> but was <%s>", expectedParams[i].Key, p.Key)
		}
		if p.Value != expectedParams[i].Value {
			t.Errorf("Wrong parameter value after trim. Expected %s but was %s", expectedParams[i].Value, p.Value)
		}
	}
}

func compareAPIDefinitions(t *testing.T, expectedDef, actualDef apiCallDef) {
	var expectedAddr, actualAddr = expectedDef.Address, actualDef.Address
	if expectedAddr.Method != actualAddr.Method {
		t.Errorf("Wrong method. Expected %s but was %s", expectedAddr.Method, actualAddr.Method)
	}
	if expectedAddr.URL != actualAddr.URL {
		t.Errorf("Wrong url. Expected %s but was %s", expectedAddr.URL, actualAddr.URL)
	}
	if expectedDef.RequestBodyTpl != actualDef.RequestBodyTpl {
		t.Error("Wrong request body template")
	}
	if expectedDef.Response.BodyTpl != actualDef.Response.BodyTpl {
		t.Error("Wrong response body template")
	}
	if expectedDef.Response.Format != actualDef.Response.Format {
		t.Error("Wrong response body format")
	}
	if expectedDef.Response.SkipVerification != actualDef.Response.SkipVerification {
		t.Errorf("Wrong flag to skip verification. Expected %t but was %t", expectedDef.Response.SkipVerification, actualDef.Response.SkipVerification)
	}
	if expectedDef.Status != actualDef.Status {
		t.Errorf("Wrong response status. Expected %d but was %d", expectedDef.Status, actualDef.Status)
	}
	if len(expectedDef.Params) != len(actualDef.Params) {
		t.Errorf("Wrong number of parameters. Expected %d but was %d", len(expectedDef.Params), len(actualDef.Params))
	} else {
		for i, ep := range expectedDef.Params {
			ap := actualDef.Params[i]
			if ep.Key != ap.Key {
				t.Errorf("Wrong parameter key. Expected %s but was %s", ep.Key, ap.Key)
			}
			if ep.Value != ap.Value {
				t.Errorf("Wrong parameter value for parameter %s. Expected %s but was %s", ep.Key, ep.Value, ap.Value)
			}
		}
	}
}

func createObjectToTrim() apiCallDef {
	return apiCallDef{
		RequestBodyTpl: ` Two
lines `,
		Response: response{"	Test with		 tabulation	", false, " text "},
		Address: reqAddr{
			Method: "	  POST   ",
			URL: `
			localhost
		`,
		},
		Status: 100,
		Params: []formParam{
			{" Name1", "	Value1"},
			{`
				Name2
			`, `
				Value 2
			`},
			{"Name3", "Value 3"},
			{"    ", "     "},
		},
	}
}
