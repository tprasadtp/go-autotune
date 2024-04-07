// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package testutils

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

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
