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
