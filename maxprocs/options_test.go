// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package maxprocs

import (
	"context"
	"log/slog"
	"math"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/trampoline"
)

func TestWithCPUQuotaFunc(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		opt := WithCPUQuotaDetector(nil)
		if opt != nil {
			t.Errorf("expected nil")
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		cfg := config{}
		detector := CPUQuotaDetectorFunc(
			func(_ context.Context) (float64, error) {
				return 0, nil
			},
		)
		opt := WithCPUQuotaDetector(detector)
		opt.apply(&cfg)
		if cfg.detector == nil {
			t.Errorf("expected non nil detector")
		}
	})
}

func TestWithRoundFunc(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		opt := WithRoundFunc(nil)
		if opt != nil {
			t.Errorf("expected nil")
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		cfg := config{}
		opt := WithRoundFunc(func(f float64) int { return int(math.Ceil(f)) })
		opt.apply(&cfg)
		if cfg.roundFunc == nil {
			t.Errorf("expected non nil roundFunc")
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
