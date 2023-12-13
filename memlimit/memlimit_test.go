// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit_test

import (
	"fmt"
	"log/slog"
	"math"
	"runtime/debug"
	"strconv"
	"syscall"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func reset() {
	debug.SetMemoryLimit(math.MaxInt64)
}

func TestConfigure_EnvVariable(t *testing.T) {
	tt := []struct {
		name   string
		env    string
		expect int64
	}{
		{
			name:   "GOMEMLIMIT=MaxInt64",
			env:    strconv.FormatInt(math.MaxInt64, 10),
			expect: math.MaxInt64,
		},
		{
			name:   "invalid-unit",
			env:    "250FooBar",
			expect: math.MaxInt64,
		},
		{
			name:   "empty",
			env:    "",
			expect: math.MaxInt64,
		},
		{
			name:   "zero",
			env:    "0",
			expect: math.MaxInt64,
		},
		{
			name:   "GOMEMLIMIT=500Mi",
			env:    "500Mi",
			expect: 500 * shared.MiByte,
		},
		{
			name:   "GOMEMLIMIT=500MiB",
			env:    "500MiB",
			expect: 500 * shared.MiByte,
		},
		{
			name:   "GOMEMLIMIT=500MB",
			env:    "500MB",
			expect: 500 * shared.MByte,
		},
		{
			name:   "GOMEMLIMIT=5GiB",
			env:    "5GiB",
			expect: 5 * shared.GiByte,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GOMEMLIMIT", tc.env)
			reset()
			memlimit.Configure(memlimit.WithLogger(slog.Default()))
			c := memlimit.Current()
			if tc.expect != c {
				t.Errorf("GOMEMLIMIT expected=%d, got=%d", tc.expect, c)
			}
		})
	}
}

func TestConfigure_WithOptions(t *testing.T) {
	t.Run("Unsupported", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 0, 0, fmt.Errorf("test: %w", syscall.ENOTSUP)
			}),
		)
		v := memlimit.Current()
		if v != math.MaxInt64 {
			t.Errorf("expected=%d, got=%d", math.MaxInt64, v)
		}
	})
	t.Run("WithMemoryQuotaFunc-Errors", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 0, 0, fmt.Errorf("test: unknown error")
			}),
		)
		v := memlimit.Current()
		if v != math.MaxInt64 {
			t.Errorf("expected=%d, got=%d", math.MaxInt64, v)
		}
	})

	t.Run("FixedValue", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 0, 50 * shared.GiByte, nil
			}),
		)
		v := memlimit.Current()
		if v != 50*shared.GiByte {
			t.Errorf("expected=%d, got=%d", 500*shared.GiByte, v)
		}
	})

	t.Run("WithMaxReservePercent", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMaxReservePercent(50),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 50 * shared.GiByte, 0, nil
			}),
		)
		v := memlimit.Current()
		if v != 25*shared.GiByte {
			t.Errorf("expected=%d, got=%d", 25*shared.GiByte, v)
		}
	})

	t.Run("AlreadySet", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 0, math.MaxInt64, nil
			}),
		)
		v := memlimit.Current()
		if v != math.MaxInt64 {
			t.Errorf("expected=%d, got=%d", math.MaxInt64, v)
		}
	})

	t.Run("Undefined", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 0, 0, nil
			}),
		)
		v := memlimit.Current()
		if v != math.MaxInt64 {
			t.Errorf("expected=%d, got=%d", math.MaxInt64, v)
		}
	})

	t.Run("MaxLessThanHigh", func(t *testing.T) {
		reset()
		var max int64 = 2.5 * shared.GiByte
		var high int64 = 3 * shared.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, high, nil
			}),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(2.25 * shared.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxEqualsHigh", func(t *testing.T) {
		reset()
		var max int64 = 2.5 * shared.GiByte
		var high int64 = 3 * shared.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, high, nil
			}),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(2.25 * shared.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxGreaterThanHigh", func(t *testing.T) {
		reset()
		var max int64 = 3 * shared.GiByte
		var high int64 = 2.5 * shared.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, high, nil
			}),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(2.5 * shared.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxGreaterThan5GiB", func(t *testing.T) {
		reset()
		var max int64 = 10 * shared.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, 0, nil
			}),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(9.5 * shared.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxInvalidReserverPercent", func(t *testing.T) {
		reset()
		var max int64 = 3 * shared.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, 0, nil
			}),
			memlimit.WithMaxReservePercent(125),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(2.7 * shared.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxInvalidReserverPercent5GiB", func(t *testing.T) {
		reset()
		var max int64 = 5 * shared.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, 0, nil
			}),
			memlimit.WithMaxReservePercent(125),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(4.75 * shared.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})
}
