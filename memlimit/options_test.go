// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import (
	"context"
	"log/slog"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/internal/trampoline"
)

func TestWithMemoryQuotaFunc(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		opt := WithMemoryQuotaDetector(nil)
		if opt != nil {
			t.Errorf("expected nil")
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		cfg := config{}
		opt := WithMemoryQuotaDetector(
			MemoryQuotaDetectorFunc(
				func(_ context.Context) (int64, int64, error) {
					return 0, 0, nil
				}))
		opt.apply(&cfg)
		if cfg.detector == nil {
			t.Errorf("expected non nil memoryQuotaFunc function")
		}
	})
}

func TestWithReservePercent(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		opt := WithReserveFunc(nil)
		if opt != nil {
			t.Errorf("expected nil")
		}
	})
	t.Run("Valid", func(t *testing.T) {
		cfg := config{}
		opt := WithReserveFunc(DefaultReserveFunc())
		opt.apply(&cfg)
		if cfg.reserveFunc == nil {
			t.Errorf("expected non nil value for cfg.reserveFunc")
		}
	})
}

func TestWithLogger(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		opt := WithLogger(nil)
		if opt != nil {
			t.Errorf("expected nil")
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		cfg := config{}
		opt := WithLogger(slog.New(trampoline.NewTestingHandler(t)))
		opt.apply(&cfg)
		if cfg.logger == nil {
			t.Errorf("expected non nil logger")
		}
	})
}

func TestDefaultReserveFunc(t *testing.T) {
	tt := []struct {
		name   string
		input  int64
		expect int64
	}{
		{
			name:  "zero",
			input: 0,
		},
		{
			name:  "negative",
			input: -1,
		},
		{
			name:   "250MiB",
			input:  250 * shared.MiByte,
			expect: 25 * shared.MiByte,
		},
		{
			name:   "500MiB",
			input:  500 * shared.MiByte,
			expect: 50 * shared.MiByte,
		},
		{
			name:   "1GiB",
			input:  1024 * shared.MiByte,
			expect: 100 * shared.MiByte,
		},
		{
			name:   "5GiB",
			input:  5 * shared.GiByte,
			expect: 100 * shared.MiByte,
		},
	}
	fn := DefaultReserveFunc()
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v := fn(tc.input)
			if v != tc.expect {
				t.Errorf("expected=%d, got=%d", tc.expect, v)
			}
		})
	}
}
