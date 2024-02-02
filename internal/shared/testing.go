// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
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

// Compile time check to ensure types implement required interfaces.
var (
	_ io.Writer = (*testLogWriter)(nil)
)

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
