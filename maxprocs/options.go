// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package maxprocs

import (
	"context"
	"log/slog"

	"github.com/tprasadtp/go-autotune/internal/quota"
)

var (
	_ CPUQuotaDetector = (*CPUQuotaDetectorFunc)(nil)
	_ CPUQuotaDetector = (*quota.Detector)(nil)
)

// CPUQuotaDetector detects cpu limits configured for the workload.
type CPUQuotaDetector interface {
	DetectCPUQuota(ctx context.Context) (float64, error)
}

// CPUQuotaDetectorFunc is an adapter to allow the use of ordinary functions as
// [CPUQuotaDetector]. If f is a function with the appropriate signature,
// CPUQuotaDetectorFunc(f) is a [CPUQuotaDetector] that calls f.
type CPUQuotaDetectorFunc func(context.Context) (float64, error)

// DetectCPUQuota Implements [CPUQuotaDetector] interface.
func (fn CPUQuotaDetectorFunc) DetectCPUQuota(ctx context.Context) (float64, error) {
	return fn(ctx)
}

// Option to apply while setting GOMAXPROCS.
type Option interface {
	apply(c *config)
}

type optionFunc struct {
	fn func(*config)
}

func (opt *optionFunc) apply(f *config) {
	opt.fn(f)
}

// WithLogger configures the logger used for setting GOMAXPROCS.
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

// WithRoundFunc can be used to replace default rounding function ([math.Ceil])
// which converts fractional CPU to integer values. This is typically not
// necessary for most apps unless you do not wish your application to encounter
// CPU throttling. Replacing this with custom function may result in underutilized
// or significantly throttled CPU.
//
// When using [Vertical Pod autoscaling], if fractional CPUs is not desired, it is
// recommended to set  [cpu-integer-post-processor-enabled], to ensure CPU recommendation
// is an integer.
//
// [Vertical Pod autoscaling]: https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler
// [cpu-integer-post-processor-enabled]: https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/FAQ.md#what-are-the-parameters-to-vpa-recommender
func WithRoundFunc(fn func(float64) int) Option {
	if fn != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.roundFunc = fn
			},
		}
	}
	return nil
}

// WithCPUQuotaDetector can be used to replace default CPU quota detection algorithm.
//
// This is an advanced option intended to be used to support custom configurations.
func WithCPUQuotaDetector(d CPUQuotaDetector) Option {
	if d != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.detector = d
			},
		}
	}
	return nil
}

// DefaultCPUQuotaDetector returns default [CPUQuotaDetector].
// This can be used to extend existing quota detection algorithm without
// re-implementing it.
func DefaultCPUQuotaDetector() CPUQuotaDetector {
	return &quota.Detector{}
}
