// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package autotune_test

import (
	"math"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/parse"
	"github.com/tprasadtp/go-autotune/internal/testutils"
)

func TestLinux_EnvGOMAXPROCSOne(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOMAXPROCS=1",
		"--property=CPUQuota=150%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 1, math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSMoreThanOne(t *testing.T) {
	testutils.SkipIfCPUCountLessThan(t, 2)
	args := []string{
		"--pipe",
		"--setenv=GOMAXPROCS=2",
		"--property=CPUQuota=50%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 2, math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSInvalidNegative(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOMAXPROCS=-2",
		"--property=CPUQuota=200%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSInvalidZero(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOMAXPROCS=0",
		"--property=CPUQuota=50%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSInvalidFraction(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOMAXPROCS=0.5",
		"--property=CPUQuota=250%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_EnvGOMAXPROCSInvalid(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOMAXPROCS=foo",
		"--property=CPUQuota=50%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_NoLimits(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestLinux_CPUExactlyOne(t *testing.T) {
	testutils.SkipIfCPUControllerIsNotAvailable(t)
	args := []string{
		"--pipe",
		"--property=CPUQuota=100%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 1, math.MaxInt64)
	})
}

func TestLinux_CPULessThanOne(t *testing.T) {
	testutils.SkipIfCPUControllerIsNotAvailable(t)
	args := []string{
		"--pipe",
		"--property=CPUQuota=50%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 1, math.MaxInt64)
	})
}

func TestLinux_CPURoundToCeil(t *testing.T) {
	testutils.SkipIfCPUControllerIsNotAvailable(t)
	testutils.SkipIfCPUCountLessThan(t, 2)
	args := []string{
		"--pipe",
		"--property=CPUQuota=150%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 2, math.MaxInt64)
	})
}

func TestLinux_CPUMoreThanOne(t *testing.T) {
	testutils.SkipIfCPUControllerIsNotAvailable(t)
	testutils.SkipIfCPUCountLessThan(t, 3)
	args := []string{
		"--pipe",
		"--property=CPUQuota=250%",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, 3, math.MaxInt64)
	})
}

func TestLinux_Env_GOMEMLIMIT_WithSuffix(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOMEMLIMIT=3GiB",
		"--property=MemoryHigh=2.5G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 3*parse.GiByte)
	})
}

func TestLinux_Env_GOMEMLIMIT_Bytes(t *testing.T) {
	args := []string{
		"--pipe",
		"--setenv=GOMEMLIMIT=3221225472",
		"--property=MemoryHigh=2.5G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 3*parse.GiByte)
	})
}

func TestLinux_MemoryMaxSpecified(t *testing.T) {
	args := []string{
		"--pipe",
		"--property=MemoryMax=2.5G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.25*parse.GiByte)
	})
}

func TestLinux_MemoryMaxLarge(t *testing.T) {
	args := []string{
		"--pipe",
		"--property=MemoryMax=5G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 4.75*parse.GiByte)
	})
}

func TestLinux_MemoryHighSpecified(t *testing.T) {
	args := []string{
		"--pipe",
		"--property=MemoryHigh=2.5G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.5*parse.GiByte)
	})
}

func TestLinux_MemoryMaxLessThanMemoryHigh(t *testing.T) {
	args := []string{
		"--pipe",
		"--property=MemoryMax=2.5G",
		"--property=MemoryHigh=3G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.25*parse.GiByte)
	})
}

func TestLinux_MemoryMaxEqualsMemoryHigh(t *testing.T) {
	args := []string{
		"--pipe",
		"--property=MemoryMax=2.5G",
		"--property=MemoryHigh=2.5G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.25*parse.GiByte)
	})
}

func TestLinux_MemoryMaxEqualsMemoryHighLargeValue(t *testing.T) {
	args := []string{
		"--pipe",
		"--property=MemoryMax=5G",
		"--property=MemoryHigh=5G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 4.75*parse.GiByte)
	})
}

func TestLinux_MemoryLargerThanMemoryHigh(t *testing.T) {
	args := []string{
		"--pipe",
		"--property=MemoryMax=3G",
		"--property=MemoryHigh=2.5G",
		"--setenv=GOAUTOTUNE=debug",
	}
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.5*parse.GiByte)
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
	testutils.SystemdRun(t, args, func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}
