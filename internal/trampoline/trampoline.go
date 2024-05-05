// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package trampoline provides utilities to re-exec test functions with resource limits.
package trampoline

import (
	"testing"
	"time"
)

// Options configured on the trampoline.
type Options struct {
	// Timeout for the trampoline exe.
	Timeout time.Duration

	// CPU limits.
	CPU float64

	// memory.max on Linux and process memory limit on windows.
	M1 int

	// memory.high on Linux and JobObject memory limit on windows.
	M2 int

	// Additional environment variables in KEY=VALUE format.
	// MUST NOT contain string GO_TEST_EXEC_TRAMPOLINE(case insensitive).
	Env []string
}

// Trampoline Test scenario.
type Scenario struct {
	// Name of the test scenario. Used as name of subtest.
	Name string

	// Trampoline options.
	Opts Options

	// This is the actual test function.
	Verify func(tb testing.TB)
}

// Trampoline re-runs the current test function via systemd run on linux and
// [golang.org/x/sys/windows.CreateProcess] with appropriate resource limits.
// verify is the test function which should be checked. configure is a hook
// to run any setup tasks before running verify. Though configure can be nil
// verify must be a non nil test function.
func Trampoline(tb testing.TB, opts Options, verify func(tb testing.TB), configure func()) {
	trampoline(tb, opts, verify, configure)
}
