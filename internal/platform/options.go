// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package platform

// Option to apply.
type Option interface {
	apply(c *config)
}

type config struct {
	CgroupInterfacePath string
	ProcfsPath          string
}

type optionFunc struct {
	fn func(*config)
}

func (opt *optionFunc) apply(f *config) {
	opt.fn(f)
}

// WithProcfsPath overrides path for /proc/self. Only used on Linux.
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

// WithCgroupInterfacePath overrides path for /sys/fs/cgroup/<cgroup>. Only used on Linux.
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
