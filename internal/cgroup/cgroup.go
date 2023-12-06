// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cgroup

import (
	"fmt"
	"runtime"
)

// ErrUnsupportedPlatform is returned when cgroups is not supported on the platform.
var ErrUnsupportedPlatform = fmt.Errorf("cgroup: unsupported platform, %s", runtime.GOOS)

// Info is cgroup info for current process.
type Info struct {
	// Cgroup mount point.
	Mount string

	// Name of the CGroup.
	Name string

	// From cpu.max
	CPUQuota        float64
	CPUQuotaDefined bool

	// From memory.max
	MemoryMax        uint64
	MemoryMaxDefined bool
}

// Returns cgroup info for current process from given mountinfo file and cgroup file.
// For non linux systems this always returns nil and [ErrUnsupportedPlatform].
func GetInfo(mountInfoFile string, cgroupFile string) (*Info, error) {
	return getInfo(mountInfoFile, cgroupFile)
}
