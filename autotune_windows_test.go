// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package autotune_test

import (
	"math"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/parse"
	"github.com/tprasadtp/go-autotune/internal/testutils"
)

func TestWindows_NoLimits(t *testing.T) {
	testutils.WindowsRun(t, 0, 0, 0, "debug", func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}

func TestWindows_CPULessThanOne(t *testing.T) {
	testutils.WindowsRun(t, 0.5, 0, 0, "debug", func(t *testing.T) {
		verify(t, 1, math.MaxInt64)
	})
}

func TestWindows_CPURoundToCeil(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skipf("Skipping CPU>1 tests on single core machine")
	}
	testutils.WindowsRun(t, 1.25, 0, 0, "debug", func(t *testing.T) {
		verify(t, 2, math.MaxInt64)
	})
}

func TestWindows_CPUMoreThanOne(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skipf("Skipping CPU>1 tests on single core machine")
	}
	testutils.WindowsRun(t, 2.0, 0, 0, "debug", func(t *testing.T) {
		verify(t, 2, math.MaxInt64)
	})
}

func TestWindows_CPUMoreThanOneFraction(t *testing.T) {
	if runtime.NumCPU() < 3 {
		t.Skipf("Skipping CPU>2 tests on dual core machine")
	}
	testutils.WindowsRun(t, 2.5, 0, 0, "debug", func(t *testing.T) {
		verify(t, 3, math.MaxInt64)
	})
}

func TestWindows_JobMemoryLimit(t *testing.T) {
	testutils.WindowsRun(t, 0, 2.5*parse.GiByte, 0, "debug", func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.25*parse.GiByte)
	})
}

func TestWindows_JobMemoryLimitHigh(t *testing.T) {
	testutils.WindowsRun(t, 0, 5*parse.GiByte, 0, "debug", func(t *testing.T) {
		verify(t, runtime.NumCPU(), 4.75*parse.GiByte)
	})
}

func TestWindows_ProcessMemoryLimit(t *testing.T) {
	testutils.WindowsRun(t, 0, 0, 2.5*parse.GiByte, "debug", func(t *testing.T) {
		verify(t, runtime.NumCPU(), 2.25*parse.GiByte)
	})
}

func TestWindows_ProcessMemoryLimitHigh(t *testing.T) {
	testutils.WindowsRun(t, 0, 0, 5*parse.GiByte, "debug", func(t *testing.T) {
		verify(t, runtime.NumCPU(), 4.75*parse.GiByte)
	})
}

func TestWindows_Disable(t *testing.T) {
	testutils.WindowsRun(t, 0.5, 2.5*parse.GiByte, 0, "false", func(t *testing.T) {
		verify(t, runtime.NumCPU(), math.MaxInt64)
	})
}
