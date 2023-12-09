// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cgroup

// Info is cgroup info for the current process.
type Info struct {
	// Cgroup mount point.
	Mount string

	// Name of the CGroup.
	Name string

	// From cpu.max
	CPUQuota float64

	// From memory.max
	MemoryMax int64

	// From memory.high
	MemoryHigh int64
}

// GetInfo returns cgroup info for the current process from given mountinfo file and cgroup file.
// For non linux systems, this always returns nil and [errors.ErrUnsupported].
func GetInfo(mountInfoFile string, cgroupFile string) (*Info, error) {
	return getInfo(mountInfoFile, cgroupFile)
}
