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
// If GOMEMLIMIT environment variable is specified, it is ALWAYS used, and limits are
// ignored. If GOMEMLIMIT environment variable is invalid, runtime MAY panic. Otherwise
// this package will attempt to detect defined memory limits using platform specific APIs.
//
// Memory limit can be soft limit, or hard limit. Hard memory limit cannot be breached
// by the process and typically leads to OOM killer being invoked for the process group/process
// when it is exceeded. For this reason, to let garbage collector free up memory early before
// OOM killer is involved, a small percentage of hard memory limit is set aside as reserved.
// This memory is fully accessible to the process and the runtime, but acts as a hint to the
// garbage collector. By default, 10% is set as reserved, hard memory limit is less than 5Gi
// and 5% otherwise.
//
// For Linux, cgroup v2 interface files are used to get memory limits.
// cgroup memory limit [memory.max] is hard memory limit and [memory.high] is
// soft memory limit. If using soft memory limits, an external process SHOULD monitor
// pressure stall information of the workload/cgroup AND alleviate the reclaim pressure.
//
//   - If both [memory.max] and [memory.high] are specified, and ([memory.max] - reserved)
//     is less than [memory.high], GOMEMLIMIT is set to ([memory.max] - reserved).
//   - If both [memory.max] and [memory.high] limits are specified, and ([memory.max] - reserved)
//     is greater than [memory.high], GOMEMLIMIT is set to [memory.high].
//   - If only [memory.max] is specified, GOMEMLIMIT is set to ([memory.max] - reserved).
//   - If only [memory.high] limit is specified, GOMEMLIMIT is set to [memory.high].
//
// For a workload, running as a systemd unit with,
//
//   - [MemoryMax]=250M and [MemoryHigh]=250M GOMEMLIMIT is set to 235929600 bytes.
//     (250M - 250*(10%)) = 225MiB = 235929600. [MemoryHigh] is ignored as it is less than
//     [MemoryMax] - [MemoryMax]*[WithMaxReservePercent].
//   - [MemoryMax]=10G, GOMEMLIMIT is set to (10 - 10*(5%)) = 9.5GiB = 10200547328.
//   - [MemoryHigh]=250M but no [MemoryMax] specified, GOMEMLIMIT is set to 250MiB = 262144000 bytes.
//   - [MemoryMax]=250M but no [MemoryHigh] specified, GOMEMLIMIT is set to (250 - 250*(10%)) = 225MiB = 235929600.
//
// For Windows, [QueryInformationJobObject] API is used to get memory limits.
// [JOBOBJECT_EXTENDED_LIMIT_INFORMATION] defines per process(ProcessMemoryLimit)
// and per job memory limits(JobMemoryLimit). ProcessMemoryLimit is always preferred
// over JobMemoryLimit. Both are considered hard limits.
//
// For a windows container running with,
//
//   - Memory limit 250M, GOMEMLIMIT is set to 235929600 bytes.
//     (250M - 250*(10/100)) = 225MiB = 235929600.
//   - Memory limit 10G, GOMEMLIMIT is set to 10200547328 bytes.
//     (10 - 10*(5%)) = 9.5GiB = 10200547328.
//
// # Use in library packages
//
// Libraries SHOULD avoid importing autotune package. It should only be imported
// by the main package. For using a custom init function or configuring manually,
// use
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
