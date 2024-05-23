// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package trampoline

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
)

var _ slog.Handler = (*testingHandler)(nil)

type testingHandler struct {
	tb      testing.TB
	handler slog.Handler
}

// Event is an alias for [log/slog.Record].
type Event = slog.Record

// NewTestingHandler returns a [slog.Handler], which writes to
// Log method of [testing.TB]. If t is nil, it panics.
func NewTestingHandler(tb testing.TB) slog.Handler {
	if tb == nil {
		panic("NewTestLogHandler: t is nil")
	}
	handler := slog.NewTextHandler(
		NewWriter(tb, ""), // no prefix because its in process log.
		&slog.HandlerOptions{
			// Log all messages. t.Log will only show them on errors
			// or if using in verbose flag.
			Level: slog.LevelDebug,
			// Disable timestamps.
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if len(groups) == 0 {
					if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
						return slog.Attr{}
					}
				}
				return a
			},
		},
	)
	return &testingHandler{
		tb:      tb,
		handler: handler,
	}
}

func (h *testingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *testingHandler) WithGroup(group string) slog.Handler {
	return &testingHandler{
		tb:      h.tb,
		handler: h.handler.WithGroup(group),
	}
}

func (h *testingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &testingHandler{
		tb:      h.tb,
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *testingHandler) Handle(ctx context.Context, e Event) error {
	err := h.handler.Handle(ctx, e)
	if err != nil {
		return fmt.Errorf("trampoline(slog): %w", err)
	}
	return nil
}
