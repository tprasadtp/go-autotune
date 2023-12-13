// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package platform

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

func isFlagSet(ref, value uint32) bool {
	return (ref & value) == ref
}

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
type JOBOBJECT_CPU_RATE_CONTROL_INFORMATION struct {
	ControlFlags uint32
	Value        uint32
}

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
const (
	JOB_OBJECT_CPU_RATE_CONTROL_ENABLE uint32 = 1 << iota
	JOB_OBJECT_CPU_RATE_CONTROL_WEIGHT_BASED
	JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP
	JOB_OBJECT_CPU_RATE_CONTROL_NOTIFY
	JOB_OBJECT_CPU_RATE_CONTROL_MIN_MAX_RATE
)

func getCPUQuota(options ...Option) (float64, error) {
	// JOBOBJECT_CPU_RATE_CONTROL_INFORMATION is not defined by golang.org/x/sys/windows.
	cpuInfo := JOBOBJECT_CPU_RATE_CONTROL_INFORMATION{}
	err := windows.QueryInformationJobObject(
		windows.Handle(0),
		windows.JobObjectCpuRateControlInformation,
		uintptr(unsafe.Pointer(&cpuInfo)),
		uint32(unsafe.Sizeof(cpuInfo)),
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("platform(windows): failed to get cpu quota: %w", err)
	}

	// Check if CPU quota is defined.
	// JOB_OBJECT_CPU_RATE_CONTROL_ENABLE is set if the job's CPU rate to be controlled
	// based on weight or hard cap.
	if isFlagSet(JOB_OBJECT_CPU_RATE_CONTROL_ENABLE, cpuInfo.ControlFlags) {
		// The job's CPU rate is a hard limit. After the job reaches its CPU cycle
		// limit for the current scheduling interval, no threads associated with the
		// job will run until the next interval.
		if isFlagSet(JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP, cpuInfo.ControlFlags) {
			// Portion of processor cycles that the threads in a job object can use
			// during each scheduling interval, as the number of cycles per
			// 10,000 cycles. Unlike linux this is specified for all cores on the system.
			return float64(cpuInfo.Value) / 10000 * float64(runtime.NumCPU()), nil
		}
	}
	return 0, nil
}

//nolint:nonamedreturns // for docs.
func getMemoryQuota(options ...Option) (max, high int64, err error) {
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	err = windows.QueryInformationJobObject(
		windows.Handle(0),
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
		nil,
	)

	if err != nil {
		return 0, 0, fmt.Errorf("platform(windows): failed get to memory quota: %w", err)
	}

	// Memory can be limited by Job or process.
	// process limit is or higher priority than Joblimit.
	// Unlike Linux this is a hard limit, there is no feature to add soft limit.
	switch {
	case info.ProcessMemoryLimit > 0:
		return int64(info.ProcessMemoryLimit), 0, nil
	case info.ProcessMemoryLimit == 0 && info.JobMemoryLimit > 0:
		return int64(info.JobMemoryLimit), 0, nil
	default:
		return 0, 0, nil
	}
}
