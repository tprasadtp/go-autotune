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

// WithReservePercent configures percentage of available memory to reserve before setting
// GOMEMLIMIT. Invalid values are ignored. Default is 10%.
func WithReservePercent(percent uint64) Option {
	return &optionFunc{
		fn: func(c *config) {
			c.ReservePercent = percent
		},
	}
}
