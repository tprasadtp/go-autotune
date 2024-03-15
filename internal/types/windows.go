// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package types

// https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-updateprocthreadattribute
//
//nolint:revive,stylecheck // Keep consistent with Windows API
const (
	PROC_THREAD_ATTRIBUTE_JOB_LIST = 0x2000D
)

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
//
//nolint:revive,stylecheck // Keep consistent with Windows API
type JOBOBJECT_CPU_RATE_CONTROL_INFORMATION struct {
	ControlFlags uint32
	Value        uint32
}

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
//
//nolint:revive,stylecheck // Keep consistent with Windows API
const (
	JOB_OBJECT_CPU_RATE_CONTROL_ENABLE uint32 = 1 << iota
	JOB_OBJECT_CPU_RATE_CONTROL_WEIGHT_BASED
	JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP
	JOB_OBJECT_CPU_RATE_CONTROL_NOTIFY
	JOB_OBJECT_CPU_RATE_CONTROL_MIN_MAX_RATE
)
