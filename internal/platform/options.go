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
