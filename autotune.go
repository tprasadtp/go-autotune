// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package autotune automatically configures GOMAXPROCS and GOMEMLIMIT.
//
// Importing this package will automatically set GOMAXPROCS and GOMEMLIMIT
// via [runtime.GOMAXPROCS] and [runtime/debug.SetMemoryLimit] respectively,
// taking into consideration CPU quota and memory limits.
//
// # GOMAXPROCS
//
//   - If GOMAXPROCS environment variable is specified, it is always used, and
//     CPU quota is ignored.
//   - CPU quota is automatically determined from cgroup [cpu.max] interface file
//     for Linux and [QueryInformationJobObject] API for Windows.
//   - Factional CPUs quotas are rounded off with [math.Ceil] by default. This
//     ensures maximum resource utilization.
//   - If CPU quota is less than 1, GOMAXPROCS is set to 1.
//
// Workload with fractional CPU quota (for example, 2.1) may encounter some CPU
// throttling. If you're using [Vertical Pod autoscaling] and do not wish to encounter
// CPU throttling, it is recommended that you use [CPU Management with static policy],
// to ensure CPU recommendation is an integer.
//
// # GOMEMLIMIT
//
// Memory limits can be soft memory limit(high), or hard memory limits(max).
// This package prefers using soft memory limit(high) whenever possible.
//
// For Linux, cgroup memory limit [memory.max] is a hard memory limit and
// [memory.high] is a soft memory limit.
//
// For Windows, [QueryInformationJobObject] API is used to get memory limits.
// Windows lacks the support for soft memory limits. [JOBOBJECT_EXTENDED_LIMIT_INFORMATION]
// defines per process(ProcessMemoryLimit) and per job memory limits(JobMemoryLimit).
// ProcessMemoryLimit is always preferred over JobMemoryLimit. Both are considered hard limits.
//
//   - If GOMEMLIMIT environment variable is specified, it is always used, and
//     limits are ignored. If GOMEMLIMIT environment variable is invalid, runtime
//     may panic during initialization.
//   - A percentage of maximum available memory is set as reserved.
//     This helps to avoid OOMs when only max memory is specified.
//     By default, 10% is set as reserved for max < 5Gi and 5% for max >= 5Gi.
//   - If both max and high are positive and max - max*(reserved/100) is less than
//     high, GOMEMLIMIT is set to max - max*(reserved/100).
//   - If both max and high are positive and max - max*(reserved/100)
//     is greater than high, GOMEMLIMIT is set to high.
//   - If only max is positive, GOMEMLIMIT is set to max - max*(reserved/100).
//   - If only high is positive, GOMEMLIMIT is set to high.
//
// For example,
//   - For a workload with [MemoryMax]=250M and [MemoryHigh]=250M
//     GOMEMLIMIT is set to 235929600 bytes. (250M - 250*(10/100)) = 225MiB = 235929600.
//     [MemoryHigh] is ignored as [MemoryMax] - [WithMaxReservePercent] is less than
//     [MemoryHigh].
//   - For a workload with [MemoryMax]=10G, GOMEMLIMIT is set to 10200547328 bytes.
//     (10 - 10*(5/100)) = 9.5GiB = 10200547328.
//   - For a workload with [MemoryHigh]=250M but no [MemoryMax] specified,
//     GOMEMLIMIT is set to 250MiB = 262144000 bytes.
//   - For a workload with [MemoryMax]=250M but no [MemoryHigh] specified,
//     GOMEMLIMIT is set to 235929600 bytes. (250 - 250*(10/100)) = 225MiB = 235929600.
//
// # Use in library packages
//
// Libraries should avoid importing autotune package. It should only be imported
// by the main package. For using custom init function or configuring manually, use
//
//   - [github.com/tprasadtp/go-autotune/maxprocs] for configuring GOMAXPROCS.
//   - [github.com/tprasadtp/go-autotune/memlimit] for configuring GOMEMLIMIT.
//
// # Conflicting Modules
//
// This package MUST NOT be used with other packages which also tweak GOMAXPROCS
// or GOMEMLIMIT. Some known incompatible packages include,
//
//   - [go.uber.org/automaxprocs]
//   - [github.com/KimMachineGun/automemlimit]
//
// # Disable at Runtime
//
// To disable automatic configuration at runtime (for compiled binaries),
// Set "GOAUTOTUNE" environment variable to "false" or "0".
//
// [CPU Management with static policy]: https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#using-cpu-management-with-static-policy
// [Vertical Pod autoscaling]: https://cloud.google.com/kubernetes-engine/docs/concepts/verticalpodautoscaler
// [memory.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [memory.high]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [cpu.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#core-interface-files
// [MemoryMax]: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryMax=bytes
// [MemoryHigh]: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryHigh=bytes
// [WithMaxReservePercent]: https://pkg.go.dev/github.com/tprasadtp/go-autotune/memlimit.WithMaxReservePercent
// [QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-queryinformationjobobject
// [JOBOBJECT_EXTENDED_LIMIT_INFORMATION]: https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_extended_limit_information
package autotune
