// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

var testVerboseCache bool
var testVerboseOnce sync.Once

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

var goCoverDirCache string
var testCoverDirOnce sync.Once

// TestingCoverDir coverage data directory. Returns empty if coverage is not
// enabled or if test.gocoverdir flag or GOCOVERDIR env variable is not specified.
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
		var gocoverdirFlag = flag.Lookup("test.gocoverdir")
		if goCoverDir == "" && gocoverdirFlag != nil {
			goCoverDir = gocoverdirFlag.Value.String()
		}

		var goCoverDirEnv = strings.TrimSpace(os.Getenv("GOCOVERDIR"))
		if goCoverDir == "" && goCoverDirEnv != "" {
			goCoverDir = goCoverDirEnv
		}

		// Return empty string
		if goCoverDir != "" {
			goCoverDirCache = goCoverDir
		}
	})
	t.Helper()

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

// SystemdRun runs test function fn via systemd-run.
func SystemdRun(t *testing.T, flags []string, fn func(t *testing.T)) {
	t.Helper()
	if fn == nil {
		t.Fatalf("fn function is nil")
	}

	if _, err := exec.LookPath("systemd-run"); err != nil {
		t.Fatalf("systemd-run binary is not available")
	}

	trampoline := strings.TrimSpace(strings.ToLower(os.Getenv("GO_TEST_EXEC_TRAMPOLINE")))

	// If trampoline is true, run the given test function.
	if trampoline == "true" {
		fn(t)
		return
	}

	osOrSystem := "--user"

	if syscall.Geteuid() == 0 {
		osOrSystem = "--system"
	}

	// Build arguments to re-exec this test.
	args := []string{
		osOrSystem,
		"--no-ask-password",
		"--wait",
	}

	// If test is not verbose, hide systemd-run logs via --quiet flag
	if !TestingIsVerbose() {
		args = append(args, "--quiet")
	}

	// Args specified by tests.
	args = append(args, flags...)

	// Trampoline args.
	args = append(args,
		// Always override GO_TEST_EXEC_TRAMPOLINE env set by args.
		"--setenv=GO_TEST_EXEC_TRAMPOLINE=true",
		// Pass other arguments to test binary.
		"--",
		os.Args[0],
		fmt.Sprintf("-test.run=^%s$", t.Name()),
	)

	// Add verbose flag if test also mentions it.
	if TestingIsVerbose() {
		args = append(args, "-test.v=true")
	}

	// The return value will be empty if test coverage is not enabled.
	if TestingCoverDir(t) != "" {
		args = append(args, fmt.Sprintf("--test.gocoverdir=%s", TestingCoverDir(t)))
	}

	// Set timeouts.
	var ctx context.Context
	var cancel context.CancelFunc
	if ts, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(context.Background(), ts)
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*30)
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemd-run", args...)
	t.Logf("Running via : systemd-run %v", cmd.Args)
	buf, err := cmd.CombinedOutput()

	t.Logf("Output of systemd-run:\n%s", string(buf))
	if err != nil {
		t.Fatalf("Failed to re-exec test: %s", err)
	}
}
