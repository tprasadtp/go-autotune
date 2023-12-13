// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit_test

import (
	"log/slog"
	"math"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func TestConfigure_Linux(t *testing.T) {
	// Do not use table driven tests,
	// as test binary re-execs with systemd-run.
	t.Run("NoLimits", func(t *testing.T) {
		args := []string{
			"--pipe",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			memlimit.Configure(memlimit.WithLogger(slog.Default()))
			var v = memlimit.Current()
			var expect int64 = math.MaxInt64
			if v != expect {
				t.Errorf("expected=%d, got=%d", expect, v)
			}
		})
	})
}
