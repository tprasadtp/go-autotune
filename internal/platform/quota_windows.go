// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package platform

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"golang.org/x/sys/windows"
)

func isFlagSet(ref, value uint32) bool {
	return (ref & value) == ref
}

func getCPUQuota(_ ...Option) (float64, error) {
	cpuInfo := shared.JOBOBJECT_CPU_RATE_CONTROL_INFORMATION{}
	err := windows.QueryInformationJobObject(
		windows.Handle(0),
		windows.JobObjectCpuRateControlInformation,
		uintptr(unsafe.Pointer(&cpuInfo)),
		uint32(unsafe.Sizeof(cpuInfo)),
		nil,
	)
	if err != nil && !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return 0, fmt.Errorf("platform(windows): failed to get cpu quota: %w", err)
	}

	// Check if CPU quota is defined.
	// JOB_OBJECT_CPU_RATE_CONTROL_ENABLE is set if the job's CPU rate to be controlled
	// based on weight or hard cap.
	if isFlagSet(shared.JOB_OBJECT_CPU_RATE_CONTROL_ENABLE, cpuInfo.ControlFlags) {
		// The job's CPU rate is a hard limit. After the job reaches its CPU cycle
		// limit for the current scheduling interval, no threads associated with the
		// job will run until the next interval.
		if isFlagSet(shared.JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP, cpuInfo.ControlFlags) {
			// Portion of processor cycles that the threads in a job object can use
			// during each scheduling interval, as the number of cycles per
			// 10,000 cycles. Unlike Linux, this is specified for all cores on the system.
			return float64(cpuInfo.Value) / 10000 * float64(runtime.NumCPU()), nil
		}
	}
	return 0, nil
}

//nolint:nonamedreturns // for docs.
func getMemoryQuota(_ ...Option) (max, high int64, err error) {
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	err = windows.QueryInformationJobObject(
		windows.Handle(0),
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
		nil,
	)
	if err != nil && !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return 0, 0, fmt.Errorf("platform(windows): failed get to memory quota: %w", err)
	}

	// Memory can be limited by Job or process.
	// process limit is or higher priority than Joblimit.
	// Unlike Linux, this is a hard limit; there is no feature to add soft limit.
	switch {
	case info.ProcessMemoryLimit > 0:
		return int64(info.ProcessMemoryLimit), 0, nil
	case info.ProcessMemoryLimit == 0 && info.JobMemoryLimit > 0:
		return int64(info.JobMemoryLimit), 0, nil
	default:
		return 0, 0, nil
	}
}
