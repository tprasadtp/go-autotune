// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package platform

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
