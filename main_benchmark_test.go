package main

import (
	"testing"
)

func BenchmarkLint(b *testing.B) {
	var writer noOpWriter
	for b.Loop() {
		lint(writer, writer, false, []string{"testdata/benchmark/*.sql"}, nil, formatText)
	}
}

type noOpWriter struct{}

func (noOpWriter) Write(n []byte) (int, error) {
	return len(n), nil
}
