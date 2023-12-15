// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"bytes"
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

var hasCPUControllerCache bool
var hasCPUControllerOnce sync.Once

// SkipIfCPUControllerIsNotAvailable skips the test if CPU controller is not available.
// See https://github.com/systemd/systemd/pull/23887. This does not change test coverage
// much as unit test can use WithCPUQuotaFunc to emulate responses.
func SkipIfCPUControllerIsNotAvailable(t *testing.T) {
	// systemctl show user@$(id -u).service --property=DelegateControllers
	hasCPUControllerOnce.Do(func() {
		uid := os.Getuid()
		if uid == 0 {
			hasCPUControllerCache = true
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//nolint:gosec // input is trusted
		cmd := exec.CommandContext(ctx,
			"systemctl",
			"show",
			"--property=DelegateControllers",
			fmt.Sprintf("user@%d.service", uid),
		)
		buf := &bytes.Buffer{}
		cmd.Stderr = buf
		cmd.Stdout = buf

		t.Logf("Checking is CPU controllers are available via: %s", cmd)
		err := cmd.Run()
		if err != nil {
			t.Errorf("Failed to run cmd '%s': %s", cmd, err)
		}

		t.Logf("systemctl output: %s", buf.String())
		if strings.Contains(buf.String(), "cpu") {
			hasCPUControllerCache = true
		}
	})
	if !hasCPUControllerCache {
		t.Skipf("CPUController is not available. See https://github.com/systemd/systemd/pull/23887")
	}
}

// SystemdRun runs test function fn via systemd-run.
func SystemdRun(t *testing.T, flags []string, fn func(t *testing.T)) {
	t.Helper()
	if fn == nil {
		t.Fatalf("function is nil")
	}

	if !HasCommandSystemdRun() {
		t.Fatalf("systemd-run command is not available")
	}

	// If trampoline is true, run the given test function.
	if IsTrue("GO_TEST_EXEC_TRAMPOLINE") {
		fn(t)
		return
	}

	// Check Env variables
	envv := os.Environ()
	for _, item := range envv {
		if strings.Contains(strings.ToUpper(item), "GO_TEST_EXEC_TRAMPOLINE") {
			t.Fatalf("env GO_TEST_EXEC_TRAMPOLINE is already defined")
		}
	}

	userOrSystem := "--user"

	if syscall.Getuid() == 0 {
		userOrSystem = "--system"
	}

	// Build arguments to re-exec this test.
	args := []string{
		userOrSystem,
		"--no-ask-password",
		"--wait",
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	t.Logf("Running via : systemd-run %v", cmd.Args)
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to re-exec test: %s", err)
	}
}
