// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cgroup

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
)

// Quota is cgroup info for the current process.
type Quota struct {
	// From cpu.max
	CPU float64

	// From memory.max
	MemoryMax int64

	// From memory.high
	MemoryHigh int64
}

type config struct {
	InterfacePath string
	ProFSPath     string
}

// GetInterfacePath returns base path for current process' cgroup interface files.
// For non linux systems, this always returns nil and [errors.ErrUnsupported].
func GetInterfacePath(options ...Option) (string, error) {
	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("cgroup: unsupported platform(%s): %w", runtime.GOOS, errors.ErrUnsupported)
	}

	cfg := &config{}

	for i := range options {
		if options[i] != nil {
			options[i].apply(cfg)
		}
	}

	if cfg.ProFSPath == "" {
		cfg.ProFSPath = "/proc/self"
	}

	mount, err := mountPointFromFile(filepath.Join(cfg.ProFSPath, "mountinfo"))
	if err != nil {
		return "", fmt.Errorf("cgroup: failed to get cgroup2 mountpoint: %w", err)
	}

	name, err := nameFromFile(filepath.Join(cfg.ProFSPath, "cgroup"))
	if err != nil {
		return "", fmt.Errorf("cgroup: failed to get cgroup name: %w", err)
	}

	return filepath.Join(mount, name), nil
}

// GetQuota returns cgroup quotas for the current process.
// For non linux systems, this always returns nil and [errors.ErrUnsupported].
func GetQuota(options ...Option) (*Quota, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("cgroup: unsupported platform(%s): %w", runtime.GOOS, errors.ErrUnsupported)
	}

	cfg := &config{}

	for i := range options {
		if options[i] != nil {
			options[i].apply(cfg)
		}
	}

	if cfg.InterfacePath == "" {
		path, err := GetInterfacePath(options...)
		if err != nil {
			return nil, err
		}
		cfg.InterfacePath = path
	}

	var v Quota
	var err error

	// Read cpu.max
	v.CPU, err = getCPUQuotaFromFile(filepath.Join(cfg.InterfacePath, "cpu.max"))
	if err != nil {
		return nil, fmt.Errorf("cgroup: failed to get cpu quota: %w", err)
	}

	// Read memory.max
	v.MemoryMax, err = memLimitFromFile(filepath.Join(cfg.InterfacePath, "memory.max"))
	if err != nil {
		return nil, fmt.Errorf("cgroup: failed to get memory max: %w", err)
	}

	// Read memory.high
	v.MemoryHigh, err = memLimitFromFile(filepath.Join(cfg.InterfacePath, "memory.high"))
	if err != nil {
		return nil, fmt.Errorf("cgroup: failed to get memory high: %w", err)
	}

	return &v, nil
}
