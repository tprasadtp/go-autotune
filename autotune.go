// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package autotune automatically configures GOMAXPROCS and GOMEMLIMIT.
//
// Importing this package will automatically set GOMAXPROCS and GOMEMLIMIT
// via [runtime.GOMAXPROCS] and [runtime/debug.SetMemoryLimit] respectively,
// taking into consideration CPU quota, memory limits from assigned cgroup.
//
// # GOMAXPROCS
//
//   - If GOMAXPROCS environment variable is specified, it is always used, and
//     cgroup limits are ignored (even if GOMAXPROCS is invalid).
//   - CPU quota is automatically determined from cgroup [cpu.max] interface file
//     for Linux and via [QueryInformationJobObject] for Windows.
//   - Factional CPUs quotas are rounded off with [math.Ceil] by default. This
//     ensures maximum resource utilization. However, workload with a CPU quota
//     of 2.1 may encounter some CPU throttling. It is recommended to use
//     integer CPU quotas for workloads sensitive to CPU throttling.
//
// If you're using [Vertical Pod autoscaling] and do not wish to encounter CPU
// throttling, it is recommended that you use [CPU Management with static policy],
// to ensure CPU recommendation is an integer.
//
// For Windows containers with hyper-v isolation, hypervisor emulates specified
// CPUs cores thus default value of GOMAXPROCS is optimal and is not changed.
//
// # GOMEMLIMIT
//
// This package prefers using soft memory limit whenever possible.
// For Linux, cgroup memory limit [memory.max](referred from here onwards as max) is
// a hard memory limit and [memory.high](referred from here onwards as high) is a soft
// memory limit. For Windows [JOBOBJECT_EXTENDED_LIMIT_INFORMATION] is used to get
// max allowed memory. Currently Windows lacks support for soft memory limits.
// [JOBOBJECT_EXTENDED_LIMIT_INFORMATION] defines per process(ProcessMemoryLimit) and
// per job memory limits(JobMemoryLimit). ProcessMemoryLimit is always preferred over
// JobMemoryLimit. See [QueryInformationJobObject] and [JOBOBJECT_EXTENDED_LIMIT_INFORMATION]
// for more information.
//
//   - If GOMEMLIMIT environment variable is specified, it is always used, and
//     cgroup limits are ignored. If GOMEMLIMIT is invalid, no changes to GOMEMLIMIT
//     are made.
//   - A percentage of maximum available memory is set as reserved.
//     This helps to avoid OOMs when only max memory is specified.
//     Be default 10% is set as reserved for max < 5Gi and 5% for max >= 5Gi.
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
//     [MemoryHigh] is ignored as [MemoryMax] - [WithMaxReservePercent] is less than
//     [MemoryHigh].
//   - For a workload with [MemoryHigh]=250M but no [MemoryMax] specified,
//     GOMEMLIMIT is set to 250MiB = 262144000 bytes.
//   - For a workload with [MemoryMax]=250M but no [MemoryHigh] specified,
//     GOMEMLIMIT is set to 235929600 bytes. (250 - 250*(10/100)) = 225MiB = 235929600.
//
// # Use in library packages
//
// Libraries should avoid importing this package, and it should only be imported
// by the main package.
//
// # Logging
//
// By default, logging is disabled, to debug issues, use your own init function.
// See examples for more info.
//
// # Conflicting Modules
//
// This package MUST NOT be used with other packages which tweak GOMAXPROCS and
// GOMEMLIMIT. Some known incompatible packages include,
//
//   - [go.uber.org/autmaxprocs]
//   - [github.com/KimMachineGun/automemlimit]
//
// On non linux platforms this will only respect GOMAXPROCS and GOMEMLIMIT from
// environment variables.
//
//   - See [github.com/tprasadtp/go-autotune/maxprocs] for configuring GOMAXPROCS.
//   - See [github.com/tprasadtp/go-autotune/memlimit] for configuring GOMEMLIMIT.
//
// # Disable at Runtime
//
// To disable automatic configuration at runtime(for compiled binaries),
// Set "GO_AUTOTUNE" environment variable to "false" or "0".
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
