// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

var hasCommandSystemdRunCache bool
var hasCommandSystemdRunOnce sync.Once

func HasCommandSystemdRun() bool {
	hasCommandSystemdRunOnce.Do(func() {
		if _, err := exec.LookPath("systemd-run"); err == nil {
			hasCommandSystemdRunCache = true
		}
	})
	return hasCommandSystemdRunCache
}

// SystemdRun runs test function fn via systemd-run.
func SystemdRun(t *testing.T, flags []string, fn func(t *testing.T)) {
	t.Helper()
	if fn == nil {
		t.Fatalf("fn function is nil")
	}

	if !HasCommandSystemdRun() {
		t.Fatalf("systemd-run command is not available")
	}

	// If trampoline is true, run the given test function.
	if IsTrue("GO_TEST_EXEC_TRAMPOLINE") {
		t.Logf("Running test function...")
		fn(t)
		return
	}

	// Env variables
	envv := os.Environ()
	for _, item := range envv {
		if strings.Contains(strings.ToUpper(item), "GO_TEST_EXEC_TRAMPOLINE") {
			t.Fatalf("env GO_TEST_EXEC_TRAMPOLINE is already defined")
		}
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
	for _, item := range flags {
		if strings.Contains(item, "GO_TEST_EXEC_TRAMPOLINE") {
			t.Fatalf("GO_TEST_EXEC_TRAMPOLINE cannot be passed as flag to systemd-run")
		}
	}
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
