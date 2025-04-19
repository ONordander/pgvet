package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

func configureLogger(verbose bool, w io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(level)

	return slog.New(&handler{w: w, level: level})
}

type handler struct {
	w     io.Writer
	level slog.Level
}

func (h *handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError {
		fmt.Fprintf(h.w, "\033[0;31m%s\033[0m\n", r.Message)
		return nil
	}

	fmt.Fprintf(h.w, "%s\n", r.Message)
	return nil
}

func (h *handler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *handler) WithGroup(_ string) slog.Handler {
	return h
}
