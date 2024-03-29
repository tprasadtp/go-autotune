// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package maxprocs

import "log/slog"

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
				c.Logger = logger
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
// Unless you are sure of your requirements, do not use this.
func WithRoundFunc(fn func(float64) int) Option {
	if fn != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.RoundFunc = fn
			},
		}
	}
	return nil
}

// WithCPUQuotaFunc can be used to replace default CPU quota detection algorithm.
//
// This is an advanced option intended to be used to support exotic configurations.
// Do not use this unless you are familiar with the internals of this package.
func WithCPUQuotaFunc(fn func() (float64, error)) Option {
	if fn != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.CPUQuotaFunc = fn
			},
		}
	}
	return nil
}
