// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package trampoline

import (
	"bytes"
	"io"
	"testing"
)

var _ io.Writer = (*writer)(nil)

// NewWriter returns an [io.Writer] which writes to [testing.TB.Log],
// Optionally with a prefix. Only handles unix new lines.
func NewWriter(tb testing.TB, prefix string) io.Writer {
	return &writer{
		tb:     tb,
		prefix: prefix,
		buf:    make([]byte, 0, 1024),
	}
}

// Writes to t.Log when new lines are found.
type writer struct {
	prefix string
	tb     testing.TB
	buf    []byte
}

func (l *writer) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	l.buf = append(l.buf, b...)
	var n int
	for {
		n = bytes.IndexByte(l.buf, '\n')
		if n < 0 {
			break
		}

		if l.prefix != "" {
			l.tb.Logf("(%s) %s", l.prefix, l.buf[:n])
		} else {
			l.tb.Log(string(l.buf[:n]))
		}

		if n+1 > len(l.buf) {
			l.buf = l.buf[0:]
		} else {
			l.buf = l.buf[n+1:]
		}
	}
	return len(b), nil
}
