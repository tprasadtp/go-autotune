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
	"strings"
	"sync"
	"testing"
	"time"
)

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

var (
	hasCPUControllerCache bool
	hasCPUControllerOnce  sync.Once
)

// SkipIfCPUControllerNotAvailable skips the test if CPU controller is not available.
// See https://github.com/systemd/systemd/pull/23887. This does not change test coverage
// much as unit test can use WithCPUQuotaFunc to emulate responses.
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
