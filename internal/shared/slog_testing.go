// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"testing"
)

type TestingHandler struct {
	t     testing.TB
	attrs []slog.Attr
}

// NewTestSlogHandler returns an [slog.Handler], which writes to
// Log method of [testing.TB]. If tb is nil it panics.
//
// This is neither fast not efficient, but its sufficient for tests.
// This does not implement WithGroup method.
func NewTestingHandler(tb testing.TB) slog.Handler {
	if tb == nil {
		panic("NewTestLogHandler: tb is nil")
	}
	return &TestingHandler{
		t: tb,
	}
}

func (h *TestingHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *TestingHandler) WithGroup(_ string) slog.Handler {
	panic("NewTestLogHandler: WithGroup not implemented")
}

func (h *TestingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	v := &TestingHandler{
		t:     h.t,
		attrs: slices.Clip(h.attrs),
	}
	v.attrs = append(v.attrs, attrs...)
	return v
}

func (h *TestingHandler) Handle(_ context.Context, e Event) error {
	var buf strings.Builder
	attrsCount := len(h.attrs) + e.NumAttrs()
	if attrsCount > 0 {
		e.Attrs(func(a slog.Attr) bool {
			if buf.Len() > 0 {
				buf.WriteRune(' ')
			}
			buf.WriteString(a.Key)
			buf.WriteRune('=')
			buf.WriteString(a.Value.String())
			return true
		})
	}
	if buf.Len() > 0 {
		h.t.Logf("%s: %s [%s]", e.Level, e.Message, buf.String())
	} else {
		h.t.Logf("%s: %s", e.Level, e.Message)
	}

	return nil
}