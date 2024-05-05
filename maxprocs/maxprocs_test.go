// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package maxprocs_test

import (
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

func TestConfigure_EnvVariable(t *testing.T) {
	tt := []struct {
		name   string
		env    string
		expect int
	}{
		{
			name:   "GOMAXPROCS=NumCPU",
			env:    strconv.FormatInt(int64(runtime.NumCPU()), 10),
			expect: runtime.NumCPU(),
		},
		{
			name:   "invalid-float",
			env:    "2.5",
			expect: runtime.NumCPU(),
		},
		{
			name:   "empty",
			env:    "2.5",
			expect: runtime.NumCPU(),
		},
		{
			name:   "zero",
			env:    "0",
			expect: runtime.NumCPU(),
		},
		{
			name:   "GOMAXPROCS=1",
			env:    "1",
			expect: 1,
		},
		{
			name:   "GOMAXPROCS=2",
			env:    "2",
			expect: 2,
		},
	}

	// GOMAXPROCS might have been modified by other tests,
	// Reset it before running any other tests.
	reset()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GOMAXPROCS", tc.env)
			maxprocs.Configure(maxprocs.WithLogger(slog.New(trampoline.NewTestingHandler(t))))
			c := maxprocs.Current()
			if tc.expect != c {
				t.Errorf("GOMAXPROCS expected=%d, got=%d", tc.expect, c)
			}
		})
	}
}

func TestConfigure_WithCPUQuotaFunc(t *testing.T) {
	t.Run("Unsupported", func(t *testing.T) {
		reset()
		maxprocs.Configure(
			maxprocs.WithLogger(slog.New(trampoline.NewTestingHandler(t))),
			maxprocs.WithCPUQuotaFunc(func() (float64, error) {
				return 0, fmt.Errorf("test: %w", errors.ErrUnsupported)
			}),
		)
		expected := runtime.NumCPU()
		v := maxprocs.Current()
		if v != expected {
			t.Errorf("expected=%d, got=%d", expected, v)
		}
	})
	t.Run("WithCPUQuotaFunc-Error", func(t *testing.T) {
		reset()
		maxprocs.Configure(
			maxprocs.WithLogger(slog.New(trampoline.NewTestingHandler(t))),
			maxprocs.WithCPUQuotaFunc(func() (float64, error) {
				return 0, fmt.Errorf("test: unknown error")
			}),
		)
		expected := runtime.NumCPU()
		v := maxprocs.Current()
		if v != expected {
			t.Errorf("expected=%d, got=%d", expected, v)
		}
	})
	t.Run("WithCPUQuotaFunc-FixedValue", func(t *testing.T) {
		reset()
		maxprocs.Configure(
			maxprocs.WithLogger(slog.New(trampoline.NewTestingHandler(t))),
			maxprocs.WithCPUQuotaFunc(func() (float64, error) {
				return 1, nil
			}),
		)
		v := maxprocs.Current()
		if v != 1 {
			t.Errorf("expected=1, got=%d", v)
		}
	})

	t.Run("WithRoundFunc-Floor", func(t *testing.T) {
		reset()
		maxprocs.Configure(
			maxprocs.WithLogger(slog.New(trampoline.NewTestingHandler(t))),
			maxprocs.WithRoundFunc(func(f float64) int {
				return int(math.Floor(f))
			}),
			maxprocs.WithCPUQuotaFunc(func() (float64, error) {
				return 0.5, nil
			}),
		)
		v := maxprocs.Current()
		if v != 1 {
			t.Errorf("expected=1, got=%d", v)
		}
	})

	t.Run("AlreadySet", func(t *testing.T) {
		reset()
		maxprocs.Configure(
			maxprocs.WithLogger(slog.New(trampoline.NewTestingHandler(t))),
			maxprocs.WithCPUQuotaFunc(func() (float64, error) {
				return float64(runtime.NumCPU()), nil
			}),
		)
		expected := runtime.NumCPU()
		v := maxprocs.Current()
		if v != expected {
			t.Errorf("expected=%d, got=%d", expected, v)
		}
	})

	t.Run("Undefined", func(t *testing.T) {
		reset()
		maxprocs.Configure(
			maxprocs.WithLogger(slog.New(trampoline.NewTestingHandler(t))),
			maxprocs.WithCPUQuotaFunc(func() (float64, error) {
				return 0, nil
			}),
		)
		expected := runtime.NumCPU()
		v := maxprocs.Current()
		if v != expected {
			t.Errorf("expected=%d, got=%d", expected, v)
		}
	})
}
