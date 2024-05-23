// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import (
	"context"
	"log/slog"
	"testing"

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
	t.Run("Valid", func(t *testing.T) {
		cfg := config{}
		opt := WithReservePercent(10)
		opt.apply(&cfg)
		if cfg.reserve != 10 {
			t.Errorf("expected maxReservePercent=10 got=%d", cfg.reserve)
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
