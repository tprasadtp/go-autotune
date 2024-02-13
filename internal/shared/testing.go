// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	testVerboseCache bool
	testVerboseOnce  sync.Once
)

// TestingIsVerbose returns true if test.v flag is set.
func TestingIsVerbose() bool {
	testVerboseOnce.Do(func() {
		v := flag.Lookup("test.v")
		if v != nil {
			if v.Value.String() == "true" {
				testVerboseCache = true
			}
		}
	})
	return testVerboseCache
}

var (
	goCoverDirCache  string
	testCoverDirOnce sync.Once
)

// TestingCoverDir coverage data directory. Returns empty if coverage is not
// enabled or if test.gocoverdir flag or GOCOVERDIR env variable is not specified.
// because tests can enable this globally, it is always resolved to absolute path.
//
// This uses Undocumented/Unexported test flag: -test.gocoverdir.
// https://github.com/golang/go/issues/51430#issuecomment-1344711300
func TestingCoverDir(t *testing.T) string {
	testCoverDirOnce.Do(func() {
		// The return value will be empty if test coverage is not enabled.
		if testing.CoverMode() == "" {
			return
		}

		var goCoverDir string
		gocoverdirFlag := flag.Lookup("test.gocoverdir")
		if goCoverDir == "" && gocoverdirFlag != nil {
			goCoverDir = gocoverdirFlag.Value.String()
		}

		goCoverDirEnv := strings.TrimSpace(os.Getenv("GOCOVERDIR"))
		if goCoverDir == "" && goCoverDirEnv != "" {
			goCoverDir = goCoverDirEnv
		}

		// Return empty string
		if goCoverDir != "" {
			goCoverDirCache = goCoverDir
		}
	})

	if goCoverDirCache == "" {
		return ""
	}

	// Get absolute path for GoCoverDir.
	goCoverDirAbs, err := filepath.Abs(goCoverDirCache)
	if err != nil {
		t.Fatalf("Failed to get absolute path of test.gocoverdir(%s):%s",
			goCoverDirCache, err)
	}
	return goCoverDirAbs
}

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
