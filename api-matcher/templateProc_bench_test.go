package main

import "testing"

func BenchmarkComposingFromInvalidTemplate(b *testing.B) {
	generator = randValueGenerator{}
	for i := 0; i < b.N; i++ {
		_, err := composeRequestBody(taskJSONSpec)
		if err != nil {
			b.Error("Error in benchmark", err)
		}
	}
}

func BenchmarkFieldParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parseResponseFields(taskJSONSpec)
		if err != nil {
			b.Error("Error in benchmark", err)
		}
	}
}
