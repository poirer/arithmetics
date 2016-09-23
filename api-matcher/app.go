package main

import (
	"log"
	"sync"
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

func main() {
	generator = randValueGenerator{}
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
