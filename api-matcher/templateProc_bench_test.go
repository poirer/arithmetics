package main

import "testing"

func BenchmarkComposingFromInvalidTemplate(b *testing.B) {
	b.ReportAllocs()
	generator = randValueGenerator{}
	buffers = newBufferPool()
	for i := 0; i < b.N; i++ {
		composeRequestBody(taskJSONSpec)
	}
}

func BenchmarkFieldParser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parseResponseFields(taskJSONSpec)
	}
}

func BenchmarkTrimViaReflection(b *testing.B) {
	b.ReportAllocs()
	var d = createObjectToTrim()
	for i := 0; i < b.N; i++ {
		trimAPIDefFileds(&d)
	}
}
