package main

import (
	"log"
	"sync"
)

const (
	typeString      = "string"
	typeInt         = "int"
	typeBool        = "bool"
	typeFloat       = "float"
	typeStringArray = "[string]"
	typeIntArray    = "[int]"
	typeBoolArray   = "[bool]"
	typeFloatArray  = "[float]"
	typeObject      = "object"
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

	response struct {
		BodyTpl          string `xml:"template"`
		SkipVerification bool   `xml:"skip-verification,attr"`
		Format           string `xml:"format,attr"`
	}

	apiDefList struct {
		DefList []apiCallDef `xml:"api-def"`
	}

	apiCallDef struct {
		RequestBodyTpl string      `xml:"request-body-template"`
		Address        reqAddr     `xml:"address"`
		Response       response    `xml:"response"`
		Status         int         `xml:"status"`
		Params         []formParam `xml:"parameters>param"`
		JSONInResponse bool        `xml:"json-in-response"`
	}
)

var buffers objectPool

func main() {
	generator = randValueGenerator{}
	buffers = newBufferPool()
	defs, err := readAPICallDefinitions("endpoints.xml")
	if err != nil {
		log.Fatal("Error occurred: ", err.Error())
	}
	var callChain = make(chan apiCallDef, 100)
	var waitGroup = &sync.WaitGroup{}
	for _, d := range defs {
		waitGroup.Add(1)
		callChain <- d
	}
	for i := 0; i < 2; i++ {
		startNewWorker(callChain, waitGroup)
	}
	waitGroup.Wait()
}
