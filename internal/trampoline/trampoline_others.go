// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux && !windows

package trampoline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func trampoline(tb testing.TB, opts Options, verify func(tb testing.TB), configure func()) {
	if verify == nil {
		tb.Fatalf("no verify function defined")
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

	// If CPUm M1 or M2 are defined, then error,
	// as they are only supported on windows and linux.
	if opts.CPU > 0 || opts.M1 > 0 || opts.M2 > 0 {
		tb.Errorf("Resource limits not supported: CPU=%f, M1=%d, M2=%d", opts.CPU, opts.M1, opts.M2)
	}

	// Check Env variables do not include GO_TEST_EXEC_TRAMPOLINE.
	envv := os.Environ()
	envv = append(envv, opts.Env...)
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

	// Set timeouts.
	//
	// Ideally we would set per set timeouts, but they are not available yet.
	// See https://github.com/golang/go/issues/48157 for more info.
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// Trampoline args.
	args := []string{
		// Only run a single test.
		fmt.Sprintf("-test.run=^%s$", tb.Name()),
		// Apply default timeout.
		fmt.Sprintf("-test.timeout=%s", opts.Timeout),
		// Always enable verbose logs. These are not necessarily printed
		// to stderr unless verbose logs are enabled.
		"-test.v=true",
	}

	// The return value will be empty if test coverage is not enabled.
	if v := CoverDir(tb); v != "" {
		args = append(args, fmt.Sprintf("-test.gocoverdir=%s", v))
	}

	// Add GO_TEST_EXEC_TRAMPOLINE=true to envs
	envv = append(envv, "GO_TEST_EXEC_TRAMPOLINE=true")

	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Env = envv
	cmd.Stdin = nil
	cmd.Stdout = NewWriter(tb, "trampoline")
	cmd.Stderr = NewWriter(tb, "trampoline")
	tb.Logf("Trampoline: %v", cmd.Args)
	err = cmd.Run()
	if err != nil {
		tb.Fatalf("Failed to re-exec test: %s", err)
	}
}
