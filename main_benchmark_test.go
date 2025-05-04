package main

import (
	"testing"
)

func BenchmarkLint(b *testing.B) {
	var writer noOpWriter
	for b.Loop() {
		lint(writer, writer, []string{"testdata/benchmark/*.sql"}, ptr("testdata/config-all-enabled.yaml"), formatText)
	}
}

type noOpWriter struct{}

func (noOpWriter) Write(n []byte) (int, error) {
	return len(n), nil
}
