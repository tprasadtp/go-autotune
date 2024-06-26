// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import (
	"context"
	"log/slog"
	"math"

	"github.com/tprasadtp/go-autotune/internal/quota"
	"github.com/tprasadtp/go-autotune/internal/shared"
)

var (
	_ MemoryQuotaDetector = (*MemoryQuotaDetectorFunc)(nil)
	_ MemoryQuotaDetector = (*quota.Detector)(nil)
)

// MemoryQuotaDetector detects soft and hard memory limits configured for the
// workload. max is hard memory limit and high is soft memory limit.
type MemoryQuotaDetector interface {
	DetectMemoryQuota(ctx context.Context) (max, high int64, err error)
}

// MemoryQuotaDetectorFunc is an adapter to allow the use of ordinary functions as
// [MemoryQuotaDetector]. If f is a function with the appropriate signature,
// MemoryQuotaDetectorFunc(f) is a [MemoryQuotaDetector] that calls f.
type MemoryQuotaDetectorFunc func(context.Context) (max, high int64, err error)

// DetectMemoryQuota Implements [MemoryQuotaDetector] interface.
//
//nolint:nonamedreturns // for docs.
func (fn MemoryQuotaDetectorFunc) DetectMemoryQuota(ctx context.Context) (max, high int64, err error) {
	return fn(ctx)
}

// Option to apply when configuring GOMEMLIMIT.
type Option interface {
	apply(c *config)
}

type optionFunc struct {
	fn func(*config)
}

func (opt *optionFunc) apply(f *config) {
	opt.fn(f)
}

// WithLogger configures the logger used for setting GOMEMLIMIT.
func WithLogger(logger *slog.Logger) Option {
	if logger != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.logger = logger
			},
		}
	}
	return nil
}

// WithReserveFunc calculates amount  of hard memory limit to set as reserved
// before setting GOMEMLIMIT. This reserved memory is fully accessible to the
// process and runtime, but acts as a hint to garbage collector and may help
// avoid OOM killer being invoked when dealing with transient memory spikes.
//
// By default 10% of hard memory limit is set aside is reserved with a maximum
// of 100 MiB. If your application makes extensive use of in memory cache,
// or is memory intensive, tweaking this may be necessary to avoid OOMs when
// dealing with transient memory spikes. This has no effect on soft memory limits.
func WithReserveFunc(fn func(limit int64) (reserve int64)) Option {
	if fn != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.reserveFunc = fn
			},
		}
	}
	return nil
}

// WithMemoryQuotaDetector can be used to replace default memory quota detection algorithm.
//
// This is an advanced option intended to be used to detect memory quota from non-standard
// locations.
func WithMemoryQuotaDetector(d MemoryQuotaDetector) Option {
	if d != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.detector = d
			},
		}
	}
	return nil
}

// DefaultMemoryQuotaDetector returns default [MemoryQuotaDetector].
// This can be used to extend existing quota detection algorithm without
// re-implementing it.
func DefaultMemoryQuotaDetector() MemoryQuotaDetector {
	return &quota.Detector{}
}

// DefaultReserveFunc returns default [WithReserveFunc].
func DefaultReserveFunc() func(limit int64) (reserve int64) {
	return func(i int64) int64 {
		if i <= 0 {
			return 0
		}

		if v := math.Floor(float64(i) * 0.1); v <= shared.MiByte*100 {
			return int64(v)
		}
		return shared.MiByte * 100
	}
}
