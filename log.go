package main

import (
	"fmt"
	"io"
	"log"
)

type logger struct {
	*log.Logger
}

func newLogger(w io.Writer) logger {
	return logger{log.New(w, "", 0)}
}

func (l *logger) Info(format string, args ...any) {
	l.Printf(format, args...)
}

func (l *logger) Error(format string, args ...any) {
	msg := fmt.Sprintf("\033[0;31m%s\033[0m\n", format)
	l.Printf(msg, args...)
}
