// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package maxprocs_test

import (
	"math"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/maxprocs"
)

func TestConfigure_SystemdRun(t *testing.T) {
	t.Run("NoLimits", func(t *testing.T) {
		reset()
		args := []string{
			"--pipe",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			maxprocs.Configure(maxprocs.WithLogger(logger()))
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
			maxprocs.Configure(maxprocs.WithLogger(logger()))
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
			"--property=CPUQuota=250%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			maxprocs.Configure(
				maxprocs.WithLogger(logger()),
				maxprocs.WithRoundFunc(func(f float64) int {
					return int(math.Floor(f))
				}),
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
				maxprocs.WithLogger(logger()),
			)
			v := maxprocs.Current()
			if v != 1 {
				t.Errorf("expected=1, got=%d", v)
			}
		})
	})
}
