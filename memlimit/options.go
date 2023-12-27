// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import "log/slog"

// Option to apply for setting gomaxprocs.
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
				c.Logger = logger
			},
		}
	}
	return nil
}

// WithMaxReservePercent configures percentage of max memory
// to reserve before setting GOMEMLIMIT. Invalid values are ignored.
// If not specified, Default is 10% if max memory is less than 5GiB
// otherwise 5% is used.
func WithMaxReservePercent(percent uint8) Option {
	return &optionFunc{
		fn: func(c *config) {
			c.MaxReservePercent = int64(percent)
		},
	}
}

// WithMemoryQuotaFunc can be used to replace default memory limit detection algorithm.
//
// This is an advanced option intended to be used to support exotic configurations.
// Do not use this unless you are familiar with the internals of this package.
func WithMemoryQuotaFunc(fn func() (max int64, high int64, err error)) Option {
	if fn != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.QuotaFunc = fn
			},
		}
	}
	return nil
}
