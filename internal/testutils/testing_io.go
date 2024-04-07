// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package testutils

import (
	"bytes"
	"io"
	"testing"
)

var _ io.Writer = (*testLogWriter)(nil)

type testLogWriter struct {
	t testing.TB
}

// NewTestLogWriter returns an [io.Writer], which writes to
// Log method of [testing.TB]. If tb is nil it panics.
func NewTestLogWriter(tb testing.TB) io.Writer {
	if tb == nil {
		panic("NewTestLogWriter: t is nil")
	}
	return &testLogWriter{
		t: tb,
	}
}

func (w *testLogWriter) Write(b []byte) (int, error) {
	w.t.Logf("%s", b)
	return len(b), nil
}

func NewOutputLogger(t *testing.T, prefix string) *OutputLogger {
	return &OutputLogger{
		t:      t,
		prefix: prefix,
		buf:    make([]byte, 0, 1024),
	}
}

// Writes to t.Log when new lines are found.
type OutputLogger struct {
	t      *testing.T
	buf    []byte
	prefix string
}

func (l *OutputLogger) LogOutput(b []byte) {
	if len(b) == 0 {
		return
	}
	l.t.Helper()
	l.buf = append(l.buf, b...)
	var n int
	for {
		n = bytes.IndexByte(l.buf, '\n')
		if n < 0 {
			break
		}
		l.t.Logf("(%s) %s", l.prefix, l.buf[:n])
		if n+1 > len(l.buf) {
			l.buf = l.buf[0:]
		} else {
			l.buf = l.buf[n+1:]
		}
	}
}

func (l *OutputLogger) Logf(format string, args ...any) {
	l.t.Helper()
	l.t.Logf(format, args...)
}

func (l *OutputLogger) Errorf(format string, args ...any) {
	l.t.Helper()
	l.t.Errorf(format, args...)
}

func (l *OutputLogger) Write(b []byte) (int, error) {
	l.LogOutput(b)
	return len(b), nil
}
