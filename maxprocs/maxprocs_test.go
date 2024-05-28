// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package maxprocs_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"strconv"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/trampoline"
	"github.com/tprasadtp/go-autotune/maxprocs"
)

func reset() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TestConfigure(t *testing.T) {
	tt := []struct {
		name   string
		ctx    context.Context
		opts   []maxprocs.Option
		env    string
		expect int
		ok     bool
	}{
		{
			name:   "Env/GOMAXPROCS=runtime.NumCPU()",
			env:    strconv.FormatInt(int64(runtime.NumCPU()), 10),
			expect: runtime.NumCPU(),
			ok:     true,
		},
		{
			name: "Env/GOMAXPROCS=0",
			env:  "0",
		},
		{
			name: "Env/GOMAXPROCS=InvalidFloat",
			env:  "1.5",
		},
		{
			name: "Env/GOMAXPROCS=InvalidString",
			env:  "foo-bar",
		},
		{
			name: "Env/GOMAXPROCS=Negative",
			env:  "-1",
		},
		{
			name:   "Env/GOMAXPROCS=1",
			env:    "1",
			expect: 1,
			ok:     true,
		},
		{
			name:   "Env/GOMAXPROCS=2",
			env:    "2",
			expect: 2,
			ok:     true,
		},
		{
			name:   "Unsupported",
			expect: runtime.NumCPU(),
			opts: []maxprocs.Option{
				maxprocs.WithCPUQuotaDetector(
					maxprocs.CPUQuotaDetectorFunc(
						func(context.Context) (float64, error) {
							return 0, fmt.Errorf("test: %w", errors.ErrUnsupported)
						},
					),
				),
			},
			ok: true,
		},
		{
			name: "UnknownError",
			opts: []maxprocs.Option{
				maxprocs.WithCPUQuotaDetector(
					maxprocs.CPUQuotaDetectorFunc(
						func(context.Context) (float64, error) {
							return 0, errors.New("test: unknown error")
						},
					),
				),
			},
		},
		{
			name:   "RoundFuncFloor/Quota>1",
			expect: 1,
			ok:     true,
			opts: []maxprocs.Option{
				maxprocs.WithRoundFunc(
					func(f float64) int {
						return int(math.Floor(f))
					}),
				maxprocs.WithCPUQuotaDetector(
					maxprocs.CPUQuotaDetectorFunc(
						func(context.Context) (float64, error) {
							return 1.5, nil
						},
					),
				),
			},
		},
		{
			name:   "RoundFuncFloor/Quota<1",
			expect: 1,
			ok:     true,
			opts: []maxprocs.Option{
				maxprocs.WithRoundFunc(
					func(f float64) int {
						return int(math.Floor(f))
					}),
				maxprocs.WithCPUQuotaDetector(
					maxprocs.CPUQuotaDetectorFunc(
						func(context.Context) (float64, error) {
							return 0.5, nil
						},
					),
				),
			},
		},
		{
			name:   "AlreadySet",
			ok:     true,
			expect: runtime.NumCPU(),
			opts: []maxprocs.Option{
				maxprocs.WithCPUQuotaDetector(
					maxprocs.CPUQuotaDetectorFunc(
						func(context.Context) (float64, error) {
							return float64(runtime.NumCPU()), nil
						},
					),
				),
			},
		},
		{
			name:   "IgnoreInvalid",
			ok:     true,
			expect: runtime.NumCPU(),
			opts: []maxprocs.Option{
				maxprocs.WithCPUQuotaDetector(
					maxprocs.CPUQuotaDetectorFunc(
						func(context.Context) (float64, error) {
							return -1, nil
						},
					),
				),
			},
		},
		{
			name: "InvalidRoundFunction",
			opts: []maxprocs.Option{
				maxprocs.WithRoundFunc(
					func(f float64) int {
						return int(math.Ceil(-1 * f))
					}),
				maxprocs.WithCPUQuotaDetector(
					maxprocs.CPUQuotaDetectorFunc(
						func(context.Context) (float64, error) {
							return 1, nil
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
			opts: []maxprocs.Option{
				maxprocs.WithCPUQuotaDetector(
					maxprocs.CPUQuotaDetectorFunc(
						func(context.Context) (float64, error) {
							return 1.5, nil
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
				t.Setenv("GOMAXPROCS", tc.env)
			}

			logger := slog.New(trampoline.NewTestingHandler(t))
			tc.opts = append(tc.opts, maxprocs.WithLogger(logger))
			err := maxprocs.Configure(tc.ctx, tc.opts...)
			if tc.ok {
				if err != nil {
					t.Errorf("expected no error, got %s", err)
				}
				if v := maxprocs.Current(); v != tc.expect {
					t.Errorf("GOMAXPROCS expected=%d, got=%d", tc.expect, v)
				}
			} else {
				if err == nil {
					t.Errorf("expected an error, got nil")
				}
				if v := maxprocs.Current(); v != runtime.NumCPU() {
					t.Errorf("GOMAXPROCS expected=%d, got=%d", runtime.NumCPU(), v)
				}
			}
		})
	}
}
