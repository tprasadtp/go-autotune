// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package autotune_test

import (
	"math"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func verify(t *testing.T, procs int, mem int64) {
	t.Helper()
	t.Run("GOMAXPROCS", func(t *testing.T) {
		v := maxprocs.Current()
		if v != procs {
			t.Errorf("expected=%d, got=%d", procs, v)
		}
	})

	t.Run("GOMEMLIMIT", func(t *testing.T) {
		v := memlimit.Current()
		if v != mem {
			t.Errorf("expected=%d, got=%d", mem, v)
		}
	})
}

func TestLinux(t *testing.T) {
	// Do not use table driven tests,
	// as test binary re-execs with systemd-run.
	t.Run("EnvGOMAXPROCSOne", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--setenv=GOMAXPROCS=1",
			"--property=CPUQuota=150%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, 1, math.MaxInt64)
		})
	})

	t.Run("EnvGOMAXPROCSMoreThanOne", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--setenv=GOMAXPROCS=2",
			"--property=CPUQuota=50%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, 2, math.MaxInt64)
		})
	})

	t.Run("EnvGOMAXPROCSInvalidNegative", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--setenv=GOMAXPROCS=-2",
			"--property=CPUQuota=200%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), math.MaxInt64)
		})
	})

	t.Run("EnvGOMAXPROCSInvalidZero", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--setenv=GOMAXPROCS=0",
			"--property=CPUQuota=50%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), math.MaxInt64)
		})
	})

	t.Run("EnvGOMAXPROCSInvalidFraction", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--setenv=GOMAXPROCS=0.5",
			"--property=CPUQuota=250%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), math.MaxInt64)
		})
	})

	t.Run("EnvGOMAXPROCSInvalid", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--setenv=GOMAXPROCS=foo",
			"--property=CPUQuota=50%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), math.MaxInt64)
		})
	})

	t.Run("NoLimits", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), math.MaxInt64)
		})
	})

	t.Run("CPULessThanOne", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=CPUQuota=50%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, 1, math.MaxInt64)
		})
	})

	t.Run("CPURoundToCeil", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=CPUQuota=150%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, 2, math.MaxInt64)
		})
	})

	t.Run("CPUMoreThanOne", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=CPUQuota=250%",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, 3, math.MaxInt64)
		})
	})

	t.Run("Env-GOMEMLIMIT-WithSuffix", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--setenv=GOMEMLIMIT=3GiB",
			"--property=MemoryHigh=2.5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 3*shared.GiByte)
		})
	})

	t.Run("Env-GOMEMLIMIT-Bytes", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--setenv=GOMEMLIMIT=3221225472",
			"--property=MemoryHigh=2.5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 3*shared.GiByte)
		})
	})

	t.Run("MemoryMaxSpecified", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=MemoryMax=2.5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 2.25*shared.GiByte)
		})
	})

	t.Run("MemoryMaxLarge", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=MemoryMax=5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 4.75*shared.GiByte)
		})
	})

	t.Run("MemoryHighSpecified", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=MemoryHigh=2.5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 2.5*shared.GiByte)
		})
	})

	t.Run("MemoryMaxLessThanMemoryHigh", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=MemoryMax=2.5G",
			"--property=MemoryHigh=3G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 2.25*shared.GiByte)
		})
	})
	t.Run("MemoryMaxEqualsMemoryHigh", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=MemoryMax=2.5G",
			"--property=MemoryHigh=2.5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 2.25*shared.GiByte)
		})
	})
	t.Run("MemoryMaxEqualsMemoryHighLargeValue", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=MemoryMax=5G",
			"--property=MemoryHigh=5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 4.75*shared.GiByte)
		})
	})
	t.Run("MemoryLargerThanMemoryHigh", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=debug",
			"--property=MemoryMax=3G",
			"--property=MemoryHigh=2.5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), 2.5*shared.GiByte)
		})
	})

	t.Run("Disable", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--setenv=GOAUTOTUNE=off",
			"--property=CPUQuota=150%",
			"--property=MemoryMax=3G",
			"--property=MemoryHigh=2.5G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			verify(t, runtime.NumCPU(), math.MaxInt64)
		})
	})
}
