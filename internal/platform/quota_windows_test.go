// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package platform_test

import (
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/parse"
	"github.com/tprasadtp/go-autotune/internal/platform"
	"github.com/tprasadtp/go-autotune/internal/shared"
)

func TestGetQuota_NoQuotaDirect(t *testing.T) {
	cpu, err := platform.GetCPUQuota()
	if err != nil {
		t.Errorf("expected no error, got=%s", err)
	}

	if cpu != 0 {
		t.Errorf("expected=0, got=%f", cpu)
	}

	max, high, err := platform.GetMemoryQuota()
	if err != nil {
		t.Errorf("expected no error, got=%s", err)
	}

	if max != 0 {
		t.Errorf("expected max=0, got=%d", max)
	}

	if high != 0 {
		t.Errorf("expected high=0, got=%d", high)
	}
}

func TestGetQuota_NoQuota(t *testing.T) {
	shared.WindowsRun(t, 0, 0, 0, "", func(t *testing.T) {
		cpu, err := platform.GetCPUQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if cpu != 0 {
			t.Errorf("expected=0, got=%f", cpu)
		}

		max, high, err := platform.GetMemoryQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if max != 0 {
			t.Errorf("expected max=0, got=%d", max)
		}

		if high != 0 {
			t.Errorf("expected high=0, got=%d", high)
		}
	})
}

func TestGetQuota_JobMemoryLimit(t *testing.T) {
	shared.WindowsRun(t, 0, 3*parse.GiByte, 0, "", func(t *testing.T) {
		cpu, err := platform.GetCPUQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if cpu != 0 {
			t.Errorf("expected=0, got=%f", cpu)
		}

		max, high, err := platform.GetMemoryQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if max != 3*parse.GiByte {
			t.Errorf("expected max=%d, got=%d", int64(3*parse.GiByte), max)
		}

		if high != 0 {
			t.Errorf("expected high=0, got=%d", high)
		}
	})
}

func TestGetQuota_ProcessMemoryLimit(t *testing.T) {
	shared.WindowsRun(t, 0, 0, 2.5*parse.GiByte, "", func(t *testing.T) {
		cpu, err := platform.GetCPUQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if cpu != 0 {
			t.Errorf("expected=0, got=%f", cpu)
		}

		max, high, err := platform.GetMemoryQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if max != 2.5*parse.GiByte {
			t.Errorf("expected max=%d, got=%d", int64(3*parse.GiByte), max)
		}

		if high != 0 {
			t.Errorf("expected high=0, got=%d", high)
		}
	})
}

func TestGetQuota_WithBothMemory(t *testing.T) {
	shared.WindowsRun(t, 0, 2.5*parse.GiByte, 3*parse.GiByte, "", func(t *testing.T) {
		cpu, err := platform.GetCPUQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if cpu != 0 {
			t.Errorf("expected=0, got=%f", cpu)
		}

		max, high, err := platform.GetMemoryQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if max != 3*parse.GiByte {
			t.Errorf("expected max=%d, got=%d", int64(3*parse.GiByte), max)
		}

		if high != 0 {
			t.Errorf("expected high=0, got=%d", high)
		}
	})
}

func TestGetQuota_CPULessThanOne(t *testing.T) {
	shared.WindowsRun(t, 0.5, 0, 0, "", func(t *testing.T) {
		cpu, err := platform.GetCPUQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if cpu != 0.5 {
			t.Errorf("expected=1, got=%f", cpu)
		}
	})
}

func TestGetQuota_CPUMoreThanOne(t *testing.T) {
	if runtime.NumCPU() == 1 {
		t.Skipf("Skipping CPU>1 tests on single core machine")
	}
	shared.WindowsRun(t, 1.5, 0, 0, "", func(t *testing.T) {
		cpu, err := platform.GetCPUQuota()
		if err != nil {
			t.Errorf("expected no error, got=%s", err)
		}

		if cpu != 1.5 {
			t.Errorf("expected=2, got=%f", cpu)
		}
	})
}
