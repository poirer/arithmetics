package main

import (
	"fmt"
	"strings"
	"testing"
)

const (
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

	invalidFieldSpec = `{
"alias": int : string,
"desc": string,
}`

	invalidObjectSpec = `{
"alias" {"field1":string},
"desc": string,
}`

	invalidFieldTypeSpec = `{
"alias" :unknown,
"desc": string,
}`
)

func ExampleComposeRequestBody() {
	generator = testGenerator{}
	res, err := composeRequestBody(taskJSONSpec)
	if err != nil {
		println(err.Error())
	} else {
		fmt.Println(res) // If to change this to simple println, then test fails.
	}
	// Output: {
	//"alias": "A",
	//"desc": "A",
	//"inner" : {
	//"f1":"A",
	//"f2":[1, 2, 3],
	//"f3": {"nf1":"A","nf2":"A"}
	//},
	//"etime": "A",
	//"reminders": ["A", "B"],
	//"rtime": "A",
	//"tags": ["A", "B"],
	//"ts": 1,
	//"type": "A",
	//"done": true
	//}
}

func TestComposingFromInvalidTemplate(t *testing.T) {
	generator = randValueGenerator{}
	_, err := composeRequestBody(invalidFieldTypeSpec)
	if err == nil {
		t.Error("Error expected")
	} else if err.Error() != "Unsupported type <unknown>" {
		t.Error("Wrong error message")
	}
}

func TestFieldParser(t *testing.T) {
	fields, err := parseResponseFields(taskJSONSpec)
	if err != nil {
		t.Error(err.Error())
	}
	var expectedFields = []responseField{
		{name: `"alias"`, typ: "string", nestedObj: nil},
		{name: `"desc"`, typ: "string", nestedObj: nil},
		{name: `"inner"`, typ: "object", nestedObj: []responseField{
			{name: `"f1"`, typ: "string", nestedObj: nil},
			{name: `"f2"`, typ: "[int]", nestedObj: nil},
			{name: `"f3"`, typ: "object", nestedObj: []responseField{
				{name: `"nf1"`, typ: "string", nestedObj: nil},
				{name: `"nf2"`, typ: "string", nestedObj: nil},
			}},
		}},
		{name: `"etime"`, typ: "string", nestedObj: nil},
		{name: `"reminders"`, typ: "[string]", nestedObj: nil},
		{name: `"rtime"`, typ: "string", nestedObj: nil},
		{name: `"tags"`, typ: "[string]", nestedObj: nil},
		{name: `"ts"`, typ: "int", nestedObj: nil},
		{name: `"type"`, typ: "string", nestedObj: nil},
		{name: `"done"`, typ: "bool", nestedObj: nil},
	}
	compareResponseFields(t, expectedFields, fields, "root")
}

func TestParseInvalidTemplate(t *testing.T) {
	_, err := parseResponseFields(invalidFieldSpec)
	if err == nil {
		t.Error("Error expected")
	} else if err.Error() != "Parse error: simple field definition must look like \"field\":type" {
		t.Error("Wrong error message")
	}
	_, err = parseResponseFields(invalidObjectSpec)
	if err == nil {
		t.Error("Error expected")
	} else if err.Error() != "Parse error: object field definition must look like \"file\":{...}" {
		println(err.Error())
		t.Error("Wrong error message")
	}
}

func compareResponseFields(t *testing.T, expectedFields, actualFields []responseField, sourceName string) {
	if len(actualFields) != len(expectedFields) {
		t.Errorf("Wrong number of fields in structure \"%s\". Expected %d but was %d", sourceName, len(expectedFields), len(actualFields))
	}
	for i := 0; i < len(actualFields); i++ {
		af, ef := actualFields[i], expectedFields[i]
		if af.name != ef.name {
			t.Errorf("Wrong field name in structure \"%s\". Expected %s but was %s", sourceName, ef.name, af.name)
		}
		if af.typ != ef.typ {
			t.Errorf("Wrong field type in structure \"%s\". Expected %s but was %s", sourceName, ef.typ, af.typ)
		}
		if ef.nestedObj != nil && af.nestedObj != nil {
			compareResponseFields(t, ef.nestedObj, af.nestedObj, ef.name)
		} else if ef.nestedObj != nil || af.nestedObj != nil {
			t.Errorf("Wrong nested objects in structure \"%s\"", sourceName)
		}

	}
}

type testGenerator struct{}

func (testGenerator) genValue(typeStr string) (string, error) {
	typeStr = strings.Trim(typeStr, "\n\t\r ")
	if typeStr[0] == '[' && typeStr[len(typeStr)-1] == ']' {
		typeStr = typeStr[1 : len(typeStr)-1]
		switch typeStr {
		case "string":
			return `["A", "B"]`, nil
		case "int":
			return "[1, 2, 3]", nil
		case "bool":
			return "[true]", nil
		case "float":
			return "[0.1]", nil
		default:
			return "[]", nil
		}
	}
	switch typeStr {
	case "string":
		return `"A"`, nil
	case "int":
		return "1", nil
	case "bool":
		return "true", nil
	case "float":
		return "0.1", nil
	default:
		return "", nil
	}
}
