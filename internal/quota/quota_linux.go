// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package quota

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Detector struct {
	cgroupfs string
}

// NewDetectorWithCgroupPath returns a [Detector] with custom path to cgroup interface files.
// This is typically useful in tests or to avoid parsing mountinfo file more than once and re-use
// detected cgroup interface paths. Use [GetCgroupInterfacePath] for computing path to cgroup
// interface files.
func NewDetectorWithCgroupPath(path string) *Detector {
	return &Detector{
		cgroupfs: path,
	}
}

func (d *Detector) DetectCPUQuota(_ context.Context) (float64, error) {
	var err error
	if d.cgroupfs == "" {
		d.cgroupfs, err = GetCgroupInterfacePath("")
	}

	if err != nil {
		return 0, err
	}

	file, err := os.Open(filepath.Join(d.cgroupfs, "cpu.max"))
	if err != nil {
		// If file is missing then cpu controller is not enabled
		// or cpu limits are not defined.
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
	}

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || len(fields) > 2 {
			return 0, fmt.Errorf("quota(cgroup): invalid format cpu.max")
		}

		// No CPU limits.
		if fields[0] == "max" {
			return 0, nil
		}

		// Get Maximum CPU quota
		max, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil || max == 0 {
			return 0, fmt.Errorf("quota(cgroup): invalid format cpu.max")
		}

		// Check if period is defined.
		var period uint64
		if len(fields) == 2 {
			period, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("quota(cgroup): invalid format cpu.max: %w", err)
			}
		} else {
			// Default CPU period value.
			period = 100000
		}

		return float64(max) / float64(period), nil
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("quota(cgroup): failed to scan cpu.max: %w", err)
	}

	return 0, io.ErrUnexpectedEOF
}

//nolint:nonamedreturns // for docs.
func (d *Detector) DetectMemoryQuota(_ context.Context) (hard, soft int64, err error) {
	if d.cgroupfs == "" {
		d.cgroupfs, err = GetCgroupInterfacePath("")
	}

	if err != nil {
		return 0, 0, err
	}

	// Read memory.max
	hard, err = memLimitFromFile(filepath.Join(d.cgroupfs, "memory.max"))
	if err != nil {
		return 0, 0, fmt.Errorf("quota(linux): failed to get memory max: %w", err)
	}

	// Read memory.high
	soft, err = memLimitFromFile(filepath.Join(d.cgroupfs, "memory.high"))
	if err != nil {
		return 0, 0, fmt.Errorf("quota(linux): failed to get memory high: %w", err)
	}

	return hard, soft, nil
}
