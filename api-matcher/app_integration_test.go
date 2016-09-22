package main

import (
	"bytes"
	"flag"
	"log"
	"strings"
	"testing"
)

var integration *bool

func TestIntegrationWinEndpoint(t *testing.T) {
	if !*integration {
		t.Skip("Test runs only when -integration flag is specified")
	}
	defs, err := readAPICallDefinitions("endpoints.xml")
	if err != nil {
		t.Error("Error occurred while reading call definitions", err)
	} else {
		if len(defs) != 4 {
			t.Errorf("Wrong number of calls. Expected 4 but was %d", len(defs))
			t.FailNow()
		}
		var logBuffer = bytes.NewBuffer(make([]byte, 0, 2048))
		log.SetOutput(logBuffer)
		generator = randValueGenerator{}
		for _, d := range defs {
			err := checkEndpoint(d)
			logCall(d, err)
		}
		checkLog(t, logBuffer.Bytes(), len(defs))
	}
}

func checkLog(t *testing.T, logContent []byte, expectedCalls int) {
	var s = string(logContent)
	var entries = strings.Split(s, "\n")
	var successCount, failCount int
	for _, entry := range entries {
		if strings.Contains(entry, "Success") {
			successCount++
		} else if strings.Contains(entry, "Failed") {
			failCount++
		}
	}
	t.Logf("In log found %d successful and %d failed calls", successCount, failCount)
	if failCount+successCount != expectedCalls {
		t.Errorf("Wrong number of calls was done. Expected %d but was %d", expectedCalls, failCount+successCount)
	}
	if failCount > 0 {
		t.Log("See log content")
		t.Log(s)
		t.Error("Integration test failed")
	}
}

func TestMain(m *testing.M) {
	integration = flag.Bool("integration", false, "Integration tests")
	flag.Parse()
	m.Run()
}
