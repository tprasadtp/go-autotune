// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import "log/slog"

// Option to apply for setting gomaxprocs.
type Option interface {
	apply(*config)
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
// Default is 10%.
func WithMaxReservePercent(percent uint64) Option {
	return &optionFunc{
		fn: func(c *config) {
			c.ReservePercent = percent
		},
	}
}

// WithMemLimitFunc can be used to replace default memory limit detection algorithm.
//
// This is advanced option intended to be used to support exotic configurations.
// Do not use this unless you are familiar with internals of this package.
func WithMemLimitFunc(fn func() (max uint64, high uint64, err error)) Option {
	if fn != nil {
		return &optionFunc{
			fn: func(c *config) {
				c.MemLimitFunc = fn
			},
		}
	}
	return nil
}
