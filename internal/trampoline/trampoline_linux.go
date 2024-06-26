// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package trampoline

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

//nolint:gochecknoglobals
var (
	hasCommandSystemdRunCache bool
	hasCommandSystemdRunOnce  sync.Once
)

func HasCommandSystemdRun() bool {
	hasCommandSystemdRunOnce.Do(func() {
		if _, err := exec.LookPath("systemd-run"); err == nil {
			hasCommandSystemdRunCache = true
		}
	})
	return hasCommandSystemdRunCache
}

//nolint:gochecknoglobals
var (
	hasCPUControllerCache bool
	hasCPUControllerOnce  sync.Once
)

// SkipIfCPUControllerNotAvailable skips the test if CPU controller is not available.
// See https://github.com/systemd/systemd/pull/23887. This does not change test coverage
// much as unit test can use interfaces to emulate responses.
func SkipIfCPUControllerNotAvailable(tb testing.TB) {
	// systemctl show user@$(id -u).service --property=DelegateControllers
	hasCPUControllerOnce.Do(func() {
		uid := os.Getuid()
		// Assume root always has access to CPU controller.
		// Tests do not support running in a systemd unit with already applied
		// resource limits or cgroup sandbox options.
		if uid == 0 {
			hasCPUControllerCache = true
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//nolint:gosec // input is from trusted source.
		cmd := exec.CommandContext(ctx,
			"systemctl",
			"show",
			"--property=DelegateControllers",
			fmt.Sprintf("user@%d.service", uid),
		)
		buf := &bytes.Buffer{}
		cmd.Stderr = buf
		cmd.Stdout = buf

		tb.Log("Checking is CPU controllers are available")
		err := cmd.Run()
		if err != nil {
			tb.Errorf("Failed to run cmd '%s': %s", cmd, err)
		}

		tb.Logf("systemctl output: %s", buf.String())
		if strings.Contains(buf.String(), "cpu") {
			hasCPUControllerCache = true
		}
	})
	if !hasCPUControllerCache {
		tb.Skipf("CPUController is not available. See https://github.com/systemd/systemd/pull/23887")
	}
}

func trampoline(tb testing.TB, opts Options, verify func(tb testing.TB), configure func()) {
	if verify == nil {
		tb.Fatalf("no verify function defined")
	}

	if !HasCommandSystemdRun() {
		tb.Fatalf("systemd-run command is not available")
	}

	// Options default overrides.
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second * 30
	}

	// If trampoline is defined, run the given test function.
	if _, ok := os.LookupEnv("GO_TEST_EXEC_TRAMPOLINE"); ok {
		// If fn hook is specified, then, run it.
		// This is typically the function which sets GOMAXPROCS and GOMEMLIMIT.
		// This can be nil, if its already set via import side effects.
		if configure != nil {
			configure()
		}

		// verify is a test assertion function.
		verify(tb)
		return
	}

	// Skip if CPU controller is not available.
	if opts.CPU > 0 {
		SkipIfCPUControllerNotAvailable(tb)
	}

	// Skip if available CPUs < configured CPUs. Though systemd handles
	// this fine It is not supported on Windows.
	if opts.CPU > float64(runtime.NumCPU()) {
		tb.Skipf("CPU=%f > runtime.NumCPU(%d)", opts.CPU, runtime.NumCPU())
	}

	// User or system systemd instance to use.
	userOrSystem := "--user"
	if unix.Geteuid() == 0 {
		userOrSystem = "--system"
	}

	// Build arguments to re-exec this test.
	args := []string{
		userOrSystem,
		"--no-ask-password",
		"--wait",     // wait till trampoline exits
		"--same-dir", // run in the same directory as the package being tested.
		"--collect",  // unload the transient unit after it completed, even if it failed.
		"--pipe",     // do not log to journald, instead stream to pipe
	}

	// Check Env variables do not include GO_TEST_EXEC_TRAMPOLINE.
	// and build --setenv arguments.
	for _, item := range opts.Env {
		if item != "" {
			if strings.Contains(strings.ToUpper(item), "GO_TEST_EXEC_TRAMPOLINE") {
				tb.Fatalf("env GO_TEST_EXEC_TRAMPOLINE is already defined")
			} else {
				args = append(
					args,
					fmt.Sprintf("--setenv=%s", item),
				)
			}
		}
	}

	// Breaker.
	envv := os.Environ()
	for _, item := range envv {
		if strings.Contains(strings.ToUpper(item), "GO_TEST_EXEC_TRAMPOLINE") {
			tb.Fatalf("env GO_TEST_EXEC_TRAMPOLINE is already defined")
		}
	}

	// Get Current Executable.
	exe, err := os.Executable()
	if err != nil {
		tb.Fatalf("failed to get executable: %s", err)
	}

	// CPU limit flags.
	if opts.CPU > 0 {
		args = append(
			args,
			fmt.Sprintf("--property=CPUQuota=%d%%", int(opts.CPU*100)),
		)
	}

	// M1 corresponds to memory.max
	if opts.M1 > 0 {
		args = append(args, fmt.Sprintf("--property=MemoryMax=%d", opts.M1))
	}

	// M2 corresponds to memory.high
	if opts.M2 > 0 {
		args = append(args, fmt.Sprintf("--property=MemoryHigh=%d", opts.M2))
	}

	// Set timeouts.
	//
	// Ideally we would set per set timeouts, but they are not available yet.
	// See https://github.com/golang/go/issues/48157 for more info.
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// Trampoline args.
	args = append(args,
		// Always override GO_TEST_EXEC_TRAMPOLINE env set by args.
		"--setenv=GO_TEST_EXEC_TRAMPOLINE=true",
		// Pass other arguments to test binary.
		"--",
		exe,
		// Only run a single test.
		fmt.Sprintf("-test.run=^%s$", tb.Name()),
		// Apply default timeout.
		fmt.Sprintf("-test.timeout=%s", opts.Timeout),
		// Always enable verbose logs. These are not necessarily printed
		// to stderr unless verbose logs are enabled.
		"-test.v=true",
	)

	// The return value will be empty if test coverage is not enabled.
	if v := CoverDir(tb); v != "" {
		args = append(args, fmt.Sprintf("-test.gocoverdir=%s", v))
	}

	cmd := exec.CommandContext(ctx, "systemd-run", args...)
	cmd.Stdin = nil
	cmd.Stdout = NewWriter(tb, "trampoline")
	cmd.Stderr = NewWriter(tb, "trampoline")
	tb.Logf("Running via : %v", cmd.Args)
	err = cmd.Run()
	if err != nil {
		tb.Fatalf("Failed to re-exec test: %s", err)
	}
}
