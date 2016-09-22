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
      <response-body-template/>
      <skip-response-verification>
        true
      </skip-response-verification>
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
      <response-body-template/>
      <skip-response-verification>
        true
      </skip-response-verification>
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
			RequestBodyTpl:       "",
			Address:              &reqAddr{Method: "GET", URL: "http://localhost:8080/task"},
			ResponseBodyTpl:      "",
			SkipRespVerification: true,
			Status:               200,
			Params:               []formParam{{Key: "id", Value: "0"}, {Key: "alias", Value: "golang"}},
		},
		{
			RequestBodyTpl:       `{"field1": [int], "field2": [string]}`,
			Address:              &reqAddr{Method: "DELETE", URL: "http://localhost:8080/task"},
			ResponseBodyTpl:      "",
			SkipRespVerification: true,
			Status:               400,
			Params:               nil,
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

func compareAPIDefinitions(t *testing.T, expectedDef, actualDef apiCallDef) {
	var expectedAddr, actualAddr = expectedDef.Address, actualDef.Address
	if expectedAddr != nil && actualAddr != nil {
		if expectedAddr.Method != actualAddr.Method {
			t.Errorf("Wrong method. Expected %s but was %s", expectedAddr.Method, actualAddr.Method)
		}
		if expectedAddr.URL != actualAddr.URL {
			t.Errorf("Wrong url. Expected %s but was %s", expectedAddr.URL, actualAddr.URL)
		}
	} else if expectedAddr != nil && actualAddr == nil {
		t.Error("Expected not nil address but was nil")
	} else if expectedAddr == nil && actualAddr != nil {
		t.Error("Expected nil address but was not nil")
	}
	if expectedDef.RequestBodyTpl != actualDef.RequestBodyTpl {
		t.Error("Wrong request body template")
	}
	if expectedDef.ResponseBodyTpl != actualDef.ResponseBodyTpl {
		t.Error("Wrong response body template")
	}
	if expectedDef.SkipRespVerification != actualDef.SkipRespVerification {
		t.Errorf("Wrong flag to skip verification. Expected %t but was %t", expectedDef.SkipRespVerification, actualDef.SkipRespVerification)
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
