// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package platform_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/internal/platform"
	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/internal/trampoline"
)

func TestGetQuotaDirect(t *testing.T) {
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

func TestGetQuotaTrampoline(t *testing.T) {
	tt := []trampoline.Scenario{
		{
			Name:   "Windows/NoQuota",
			Opts:   trampoline.Options{},
			Verify: VerifyQuotaFunc(0, 0, 0),
		},
		{
			Name: "Windows/CPUFraction",
			Opts: trampoline.Options{
				CPU: 0.5,
			},
			Verify: VerifyQuotaFunc(0.5, 0, 0),
		},
		{
			Name: "Windows/CPU=1",
			Opts: trampoline.Options{
				CPU: 1,
			},
			Verify: VerifyQuotaFunc(1, 0, 0),
		},
		{
			Name: "Windows/CPU=1.5",
			Opts: trampoline.Options{
				CPU: 1.5,
			},
			Verify: VerifyQuotaFunc(1.5, 0, 0),
		},
		{
			Name: "Windows/CPU=2.5",
			Opts: trampoline.Options{
				CPU: 2.5,
			},
			Verify: VerifyQuotaFunc(2.5, 0, 0),
		},
		{
			Name: "Windows/JobMemoryLimitOnly",
			Opts: trampoline.Options{
				M2: shared.MiByte * 250,
			},
			Verify: VerifyQuotaFunc(0, shared.MiByte*250, 0),
		},
		{
			Name: "Windows/ProcessMemoryLimitOnly",
			Opts: trampoline.Options{
				M1: shared.MiByte * 300,
			},
			Verify: VerifyQuotaFunc(0, shared.MiByte*300, 0),
		},
		{
			Name: "Windows/ProcessMemoryLimitSameAsJobMemoryLimit",
			Opts: trampoline.Options{
				M1: shared.MiByte * 250,
				M2: shared.MiByte * 250,
			},
			Verify: VerifyQuotaFunc(0, shared.MiByte*250, 0),
		},
		{
			Name: "Windows/ProcessMemoryLimitLessThanJobMemoryLimit",
			Opts: trampoline.Options{
				M1: shared.MiByte * 250,
				M2: shared.MiByte * 300,
			},
			Verify: VerifyQuotaFunc(0, shared.MiByte*250, 0),
		},
		// This is certainly a misconfiguration and unlikely to occur, but still test it.
		{
			Name: "Windows/ProcessMemoryLimitGreaterThanJobMemoryLimit",
			Opts: trampoline.Options{
				M1: shared.MiByte * 300,
				M2: shared.MiByte * 250,
			},
			Verify: VerifyQuotaFunc(0, shared.MiByte*250, 0),
		},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			trampoline.Trampoline(t, tc.Opts, tc.Verify, nil)
		})
	}
}
