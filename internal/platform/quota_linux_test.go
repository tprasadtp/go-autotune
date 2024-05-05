// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package platform_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/internal/trampoline"
)

func TestGetQuotaTrampoline(t *testing.T) {
	tt := []trampoline.Scenario{
		{
			Name:   "Linux/NoQuota",
			Opts:   trampoline.Options{},
			Verify: VerifyQuotaFunc(0, 0, 0),
		},
		{
			Name: "Linux/CPUFraction",
			Opts: trampoline.Options{
				CPU: 0.5,
			},
			Verify: VerifyQuotaFunc(0.5, 0, 0),
		},
		{
			Name: "Linux/CPU=1",
			Opts: trampoline.Options{
				CPU: 1,
			},
			Verify: VerifyQuotaFunc(1, 0, 0),
		},
		{
			Name: "Linux/CPU=1.5",
			Opts: trampoline.Options{
				CPU: 1.5,
			},
			Verify: VerifyQuotaFunc(1.5, 0, 0),
		},
		{
			Name: "Linux/CPU=2.5",
			Opts: trampoline.Options{
				CPU: 2.5,
			},
			Verify: VerifyQuotaFunc(2.5, 0, 0),
		},
		{
			Name: "Linux/MemoryMaxOnly",
			Opts: trampoline.Options{
				M1: shared.MiByte * 250,
			},
			Verify: VerifyQuotaFunc(0, shared.MiByte*250, 0),
		},
		{
			Name: "Linux/MemoryHighOnly",
			Opts: trampoline.Options{
				M2: shared.MiByte * 300,
			},
			Verify: VerifyQuotaFunc(0, 0, shared.MiByte*300),
		},
		{
			Name: "Linux/MemoryMaxAndMemoryHigh",
			Opts: trampoline.Options{
				M1: shared.MiByte * 300,
				M2: shared.MiByte * 250,
			},
			Verify: VerifyQuotaFunc(0, shared.MiByte*300, shared.MiByte*250),
		},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			trampoline.Trampoline(t, tc.Opts, tc.Verify, nil)
		})
	}
}
