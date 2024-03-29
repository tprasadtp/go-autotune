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

	"github.com/tprasadtp/go-autotune/internal/parse"
	"github.com/tprasadtp/go-autotune/internal/testutils"
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
			name:   "500MiB",
			env:    "500MiB",
			expect: 500 * parse.MiByte,
		},
		{
			name:   "5GiB",
			env:    "5GiB",
			expect: 5 * parse.GiByte,
		},
		{
			name:   "Invalid-500Mi",
			env:    "500Mi",
			expect: math.MaxInt64,
		},
		{
			name:   "Invalid-500MB",
			env:    "500MB",
			expect: math.MaxInt64,
		},
		{
			name:   "Invalid-Float-2.5GiB",
			env:    "2.5GiB",
			expect: math.MaxInt64,
		},
		{
			name:   "Invalid-Unit",
			env:    "250FooBar",
			expect: math.MaxInt64,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GOMEMLIMIT", tc.env)
			reset()
			memlimit.Configure(memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))))
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
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
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
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
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
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 0, 50 * parse.GiByte, nil
			}),
		)
		v := memlimit.Current()
		if v != 50*parse.GiByte {
			t.Errorf("expected=%d, got=%d", 500*parse.GiByte, v)
		}
	})

	t.Run("WithMaxReservePercent-50", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMaxReservePercent(50),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 10 * parse.GiByte, 0, nil
			}),
		)
		v := memlimit.Current()
		if v != 5*parse.GiByte {
			t.Errorf("expected=%d, got=%d", 5*parse.GiByte, v)
		}
	})

	t.Run("WithMaxReservePercent-0", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMaxReservePercent(0),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return 5 * parse.GiByte, 0, nil
			}),
		)
		v := memlimit.Current()
		if v != 5*parse.GiByte {
			t.Errorf("expected=%d, got=%d", 5*parse.GiByte, v)
		}
	})

	t.Run("AlreadySet", func(t *testing.T) {
		reset()
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
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
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
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
		var max int64 = 2.5 * parse.GiByte
		var high int64 = 3 * parse.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, high, nil
			}),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(2.25 * parse.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxEqualsHigh", func(t *testing.T) {
		reset()
		var max int64 = 2.5 * parse.GiByte
		var high int64 = 3 * parse.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, high, nil
			}),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(2.25 * parse.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxGreaterThanHigh", func(t *testing.T) {
		reset()
		var max int64 = 3 * parse.GiByte
		var high int64 = 2.5 * parse.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, high, nil
			}),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(2.5 * parse.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxGreaterThan5GiB", func(t *testing.T) {
		reset()
		var max int64 = 10 * parse.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, 0, nil
			}),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(9.5 * parse.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxInvalidReserverPercent", func(t *testing.T) {
		reset()
		var max int64 = 3 * parse.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, 0, nil
			}),
			memlimit.WithMaxReservePercent(125),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(2.7 * parse.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})

	t.Run("MaxInvalidReserverPercent5GiB", func(t *testing.T) {
		reset()
		var max int64 = 5 * parse.GiByte
		memlimit.Configure(
			memlimit.WithLogger(slog.New(testutils.NewTestingHandler(t))),
			memlimit.WithMemoryQuotaFunc(func() (int64, int64, error) {
				return max, 0, nil
			}),
			memlimit.WithMaxReservePercent(125),
		)
		v := memlimit.Current()
		// because we compute reserve with ceil, expect should be floor.
		expect := int64(math.Floor(4.75 * parse.GiByte))
		if v != expect {
			t.Errorf("expected=%d, got=%d", expect, v)
		}
	})
}
