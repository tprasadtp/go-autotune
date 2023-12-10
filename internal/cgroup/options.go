// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cgroup

type optionFunc struct {
	fn func(*config)
}

func (opt *optionFunc) apply(f *config) {
	opt.fn(f)
}

// Option to apply.
type Option interface {
	apply(*config)
}

// WithProcFSPath overrides path for /proc/self.
func WithProcFSPath(path string) Option {
	if path != "" {
		return &optionFunc{
			fn: func(c *config) {
				c.ProFSPath = path
			},
		}
	}
	return nil
}

// WithInterfacePath overrides path for /sys/fs/cgroup/<cgroup>.
func WithInterfacePath(path string) Option {
	if path != "" {
		return &optionFunc{
			fn: func(c *config) {
				c.InterfacePath = path
			},
		}
	}
	return nil
}
