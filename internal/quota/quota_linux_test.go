// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package quota_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/quota"
	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/internal/trampoline"
)

func TestDetectCPUQuota(t *testing.T) {
	tt := []struct {
		name   string
		path   string
		expect float64
		err    bool
	}{
		{
			name: "no-limits",
			path: "no-limits",
		},
		{
			name:   "cpu-50",
			path:   "cpu-50",
			expect: 0.5,
		},
		{
			name:   "cpu-250",
			path:   "cpu-250",
			expect: 2.5,
		},
		{
			name:   "cpu-250-10ms",
			path:   "cpu-250-10ms",
			expect: 2.5,
		},
		{
			name:   "cpu-300",
			path:   "cpu-300",
			expect: 3,
		},
		{
			name: "cpu-invalid",
			path: "cpu-invalid",
			err:  true,
		},
		{
			name: "cpu-negative",
			path: "cpu-negative",
			err:  true,
		},
		{
			name: "cpu-negative-interval",
			path: "cpu-negative-interval",
			err:  true,
		},
		{
			name: "no-limits-no-files",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			d := quota.NewDetectorWithCgroupPath(
				filepath.Join("testdata", "cgroup", tc.path),
			)

			ctx := context.Background()
			v, err := d.DetectCPUQuota(ctx)

			if tc.err {
				if err == nil {
					t.Errorf("expected an error, got nil")
				}

				if v != 0 {
					t.Errorf("must return 0 when error is expected, got=%f", v)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %s", err)
				}
				if v != tc.expect {
					t.Errorf("expected=%f, got=%f", tc.expect, v)
				}
			}
		})
	}
}

func TestDetectMemoryQuota(t *testing.T) {
	tt := []struct {
		name string
		path string
		max  int64
		high int64
		err  bool
	}{
		{
			name: "no-limits",
			path: "no-limits",
		},
		{
			name: "mem-high-250",
			path: "mem-high-250",
			high: 250 * shared.MiByte,
		},
		{
			name: "mem-max-250",
			path: "mem-max-250",
			max:  250 * shared.MiByte,
		},
		{
			name: "mem-max-250-high-200",
			path: "mem-max-250-high-200",
			max:  250 * shared.MiByte,
			high: 200 * shared.MiByte,
		},
		{
			name: "mem-max-250-high-250",
			path: "mem-max-250-high-250",
			max:  250 * shared.MiByte,
			high: 250 * shared.MiByte,
		},
		{
			name: "mem-max-300-high-500",
			path: "mem-max-300-high-500",
			max:  300 * shared.MiByte,
			high: 500 * shared.MiByte,
		},
		{
			name: "mem-high-invalid",
			path: "mem-high-invalid",
			err:  true,
		},
		{
			name: "mem-high-negative",
			path: "mem-high-negative",
			err:  true,
		},
		{
			name: "mem-max-invalid",
			path: "mem-max-invalid",
			err:  true,
		},
		{
			name: "mem-max-negative",
			path: "mem-max-negative",
			err:  true,
		},
		{
			name: "no-limits-no-files",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			d := quota.NewDetectorWithCgroupPath(
				filepath.Join("testdata", "cgroup", tc.path),
			)

			ctx := context.Background()
			max, high, err := d.DetectMemoryQuota(ctx)

			if tc.err {
				if err == nil {
					t.Errorf("expected an error, got nil")
				}

				if max != 0 {
					t.Errorf("max=0 when error is expected")
				}

				if high != 0 {
					t.Errorf("high=0 when error is expected")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %s", err)
				}
				if max != tc.max {
					t.Errorf("max=%d expected=%d", max, tc.max)
				}

				if high != tc.high {
					t.Errorf("high=%d expected=%d", max, tc.high)
				}
			}
		})
	}
}

func TestTrampolineLinux(t *testing.T) {
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
