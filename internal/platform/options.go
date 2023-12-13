// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package platform

type config struct {
	CgroupInterfacePath string
	ProcfsPath          string
}

type optionFunc struct {
	fn func(*config)
}

// Option to apply.
type Option interface {
	apply(c *config)
}

func (opt *optionFunc) apply(f *config) {
	opt.fn(f)
}
