// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package autotune_test

import (
	"math"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
)

func TestLinux_EnvGOMAXPROCSOne(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--setenv=GOMAXPROCS=1",
		"--property=CPUQuota=150%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 1, math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSMoreThanOne(t *testing.T) {
	if runtime.NumCPU() == 1 {
		t.Skipf("Skipping CPU>1 tests on single core machine")
	}
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--setenv=GOMAXPROCS=2",
		"--property=CPUQuota=50%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 2, math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSInvalidNegative(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--setenv=GOMAXPROCS=-2",
		"--property=CPUQuota=200%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSInvalidZero(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--setenv=GOMAXPROCS=0",
		"--property=CPUQuota=50%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSInvalidFraction(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--setenv=GOMAXPROCS=0.5",
		"--property=CPUQuota=250%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSInvalid(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--setenv=GOMAXPROCS=foo",
		"--property=CPUQuota=50%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_NoLimits(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_CPUExactlyOne(t *testing.T) {
	shared.SkipIfCPUControllerIsNotAvailable(t)
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=CPUQuota=100%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 1, math.MaxInt64)
	})
}

func TestLinux_CPULessThanOne(t *testing.T) {
	shared.SkipIfCPUControllerIsNotAvailable(t)
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=CPUQuota=50%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 1, math.MaxInt64)
	})
}

func TestLinux_CPURoundToCeil(t *testing.T) {
	shared.SkipIfCPUControllerIsNotAvailable(t)
	if runtime.NumCPU() == 1 {
		t.Skipf("Skipping CPU>1 tests on single core machine")
	}
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=CPUQuota=150%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 2, math.MaxInt64)
	})
}

func TestLinux_CPUMoreThanOne(t *testing.T) {
	shared.SkipIfCPUControllerIsNotAvailable(t)
	if runtime.NumCPU() < 3 {
		t.Skipf("Skipping CPU>2 tests on dual core machine")
	}
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=CPUQuota=250%",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 3, math.MaxInt64)
	})
}

func TestLinux_Env_GOMEMLIMIT_WithSuffix(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--setenv=GOMEMLIMIT=3GiB",
		"--property=MemoryHigh=2.5G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 3*shared.GiByte)
	})
}

func TestLinux_Env_GOMEMLIMIT_Bytes(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--setenv=GOMEMLIMIT=3221225472",
		"--property=MemoryHigh=2.5G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 3*shared.GiByte)
	})
}

func TestLinux_MemoryMaxSpecified(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=MemoryMax=2.5G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.25*shared.GiByte)
	})
}

func TestLinux_MemoryMaxLarge(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=MemoryMax=5G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 4.75*shared.GiByte)
	})
}

func TestLinux_MemoryHighSpecified(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=MemoryHigh=2.5G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.5*shared.GiByte)
	})
}

func TestLinux_MemoryMaxLessThanMemoryHigh(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=MemoryMax=2.5G",
		"--property=MemoryHigh=3G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.25*shared.GiByte)
	})
}
func TestLinux_MemoryMaxEqualsMemoryHigh(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=MemoryMax=2.5G",
		"--property=MemoryHigh=2.5G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.25*shared.GiByte)
	})
}
func TestLinux_MemoryMaxEqualsMemoryHighLargeValue(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=MemoryMax=5G",
		"--property=MemoryHigh=5G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 4.75*shared.GiByte)
	})
}
func TestLinux_MemoryLargerThanMemoryHigh(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
		"--property=MemoryMax=3G",
		"--property=MemoryHigh=2.5G",
	}
	shared.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.5*shared.GiByte)
	})
}

func TestLinux_Disable(t *testing.T) {
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
}
