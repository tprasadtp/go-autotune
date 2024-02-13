// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package platform

type optionFunc struct {
	fn func(*config)
}

func (opt *optionFunc) apply(f *config) {
	opt.fn(f)
}

// WithProcfsPath overrides path for /proc/self.
func WithProcfsPath(path string) Option {
	if path != "" {
		return &optionFunc{
			fn: func(c *config) {
				c.ProcfsPath = path
			},
		}
	}
	return nil
}

// WithCgroupInterfacePath overrides path for /sys/fs/cgroup/<cgroup>.
func WithCgroupInterfacePath(path string) Option {
	if path != "" {
		return &optionFunc{
			fn: func(c *config) {
				c.CgroupInterfacePath = path
			},
		}
	}
	return nil
}
