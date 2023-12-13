// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package maxprocs_test

import (
	"log/slog"
	"math"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/maxprocs"
)

func TestConfigure_Linux(t *testing.T) {
	// Do not use table driven tests,
	// as test binary re-execs with systemd-run.
	t.Run("NoLimits", func(t *testing.T) {
		reset()
		args := []string{
			"--pipe",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			maxprocs.Configure(maxprocs.WithLogger(slog.Default()))
			v := maxprocs.Current()
			expect := runtime.NumCPU()
			if v != expect {
				t.Errorf("expected=%d, got=%d", expect, v)
			}
		})
	})
	t.Run("LessThanOne", func(t *testing.T) {
		reset()
		args := []string{
			"--pipe",
			"--property=CPUQuota=50%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			maxprocs.Configure(maxprocs.WithLogger(slog.Default()))
			v := maxprocs.Current()
			if v != 1 {
				t.Errorf("expected=1, got=%d", v)
			}
		})
	})
	t.Run("RoundFuncFloor", func(t *testing.T) {
		reset()
		args := []string{
			"--pipe",
			"--property=CPUQuota=150%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			maxprocs.Configure(
				maxprocs.WithLogger(slog.Default()),
				maxprocs.WithRoundFunc(func(f float64) int {
					return int(math.Floor(f))
				}),
			)
			v := maxprocs.Current()
			if v != 1 {
				t.Errorf("expected=2, got=%d", v)
			}
		})
	})
	t.Run("RoundFuncDefault", func(t *testing.T) {
		reset()
		args := []string{
			"--pipe",
			"--property=CPUQuota=150%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			maxprocs.Configure(
				maxprocs.WithLogger(slog.Default()),
			)
			v := maxprocs.Current()
			if v != 2 {
				t.Errorf("expected=2, got=%d", v)
			}
		})
	})

	t.Run("Integer", func(t *testing.T) {
		reset()
		args := []string{
			"--pipe",
			"--property=CPUQuota=100%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			maxprocs.Configure(
				maxprocs.WithLogger(slog.Default()),
			)
			v := maxprocs.Current()
			if v != 1 {
				t.Errorf("expected=1, got=%d", v)
			}
		})
	})
}
