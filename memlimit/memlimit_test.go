// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"runtime/debug"
	"strconv"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/trampoline"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func reset() {
	debug.SetMemoryLimit(math.MaxInt64)
}

func TestConfigure(t *testing.T) {
	tt := []struct {
		name     string
		ctx      context.Context
		opts     []memlimit.Option
		res      uint8
		detector memlimit.MemoryQuotaDetector
		env      string
		expect   int64
		ok       bool
	}{
		{
			name: "Env/GOMEMLIMIT=0",
			env:  "0",
			ok:   true,
		},
		{
			name: "Env/GOMEMLIMIT=InvalidFloat",
			env:  "1.5.1Mi",
		},
		{
			name: "Env/GOMEMLIMIT=InvalidString",
			env:  "foo-bar",
		},
		{
			name: "Env/GOMEMLIMIT=InvalidUnit",
			env:  "5Gi",
		},
		{
			name:   "Env/GOMEMLIMIT=5000MiB",
			env:    "5000MiB",
			expect: shared.MiByte * 5000,
			ok:     true,
		},
		{
			name:   "Env/GOMEMLIMIT=500000KiB",
			env:    "500000KiB",
			expect: shared.KiByte * 500000,
			ok:     true,
		},
		{
			name:   "Env/GOMEMLIMIT=math.MaxInt64",
			env:    strconv.FormatInt(int64(math.MaxInt64), 10),
			expect: math.MaxInt64,
			ok:     true,
		},
		{
			name:   "Env/GOMEMLIMIT=off",
			env:    "off",
			expect: math.MaxInt64,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return shared.GiByte * 5, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "NotSpecified",
			expect: math.MaxInt64,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 0, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "Unsupported",
			expect: math.MaxInt64,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 0, 0, errors.ErrUnsupported
						},
					),
				),
			},
		},
		{
			name: "ContextCancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
		},
		{
			name: "MemoryQuotaFuncErrors",
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 0, 0, fmt.Errorf("fake: test error")
						},
					),
				),
			},
		},
		{
			name:   "WithReservePercent/Default/HardLimit>=5GiB",
			expect: 4.75 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return shared.GiByte * 5, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "WithReservePercent/Default/HardLimit<5GiB",
			expect: 2.25 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 2.5 * shared.GiByte, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "WithReservePercent/50",
			expect: 2.5 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithReservePercent(50),
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return shared.GiByte * 5, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "WithReservePercent/0",
			expect: 5 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithReservePercent(0),
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return shared.GiByte * 5, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "WithReservePercent/Invalid/HardLimit>=5Gi",
			expect: 4.75 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithReservePercent(250),
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return shared.GiByte * 5, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "WithReservePercent/Invalid/HardLimit<5Gi",
			expect: 2.25 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithReservePercent(250),
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 2.5 * shared.GiByte, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "Undefined",
			expect: math.MaxInt64,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 0, 0, nil
						},
					),
				),
			},
		},
		// With Limits
		{
			name:   "HardLimitOnly/LessThan5GiB",
			expect: 2.25 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 2.5 * shared.GiByte, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "HardLimitOnly/5GiB",
			expect: 4.75 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 5 * shared.GiByte, 0, nil
						},
					),
				),
			},
		},
		{
			name:   "SoftLimitOnly",
			expect: 2.5 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 0, 2.5 * shared.GiByte, nil
						},
					),
				),
			},
		},
		{
			name:   "HardLimit=SoftLimit",
			expect: 2.25 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 2.5 * shared.GiByte, 2.5 * shared.GiByte, nil
						},
					),
				),
			},
		},
		{
			name:   "HardLimit>SoftLimit",
			expect: 2.5 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 3 * shared.GiByte, 2.5 * shared.GiByte, nil
						},
					),
				),
			},
		},
		{
			name:   "HardLimit<SoftLimit",
			expect: 2.25 * shared.GiByte,
			ok:     true,
			opts: []memlimit.Option{
				memlimit.WithMemoryQuotaDetector(
					memlimit.MemoryQuotaDetectorFunc(
						func(_ context.Context) (int64, int64, error) {
							return 2.5 * shared.GiByte, 3 * shared.GiByte, nil
						},
					),
				),
			},
		},
	}
	t.Cleanup(reset) // avoid side effects in other tests.

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(reset)

			if tc.env != "" {
				t.Setenv("GOMEMLIMIT", tc.env)
			}

			logger := slog.New(trampoline.NewTestingHandler(t))
			tc.opts = append(tc.opts, memlimit.WithLogger(logger))
			err := memlimit.Configure(tc.ctx, tc.opts...)
			if tc.ok {
				if err != nil {
					t.Errorf("expected no error, got %s", err)
				}
				if v := memlimit.Current(); v != tc.expect {
					t.Errorf("GOMEMLIMIT expected=%d, got=%d", tc.expect, v)
				}
			} else {
				if err == nil {
					t.Errorf("expected an error, got nil")
				}
				if v := memlimit.Current(); v != math.MaxInt64 {
					t.Errorf("GOMEMLIMIT expected=math.MaxInt64, got=%d", v)
				}
			}
		})
	}
}
