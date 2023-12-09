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
//   - CPU quota is automatically determined from [cpu.max].
//   - Factional CPUs quotas are rounded off with [math.Ceil] by default. This
//     ensures maximum resource utilization. However, workload with CPU quota
//     of say 2.1 may encounter some CPU throttling. It is recommended to use
//     integer CPU quotas for workloads sensitive to CPU throttling.
//
// If you're using [Vertical Pod autoscaling] and do not wish to encounter CPU
// throttling, it is recommended that you use [CPU Management with static policy],
// to ensure CPU recommendation is an integer.
//
// # GOMEMLIMIT
//
// This package prefers using soft memory limit whenever possible.
// cgroup memory limit [memory.max](referred from here onwards as max) is a hard
// memory limit and [memory.high](referred from here onwards as high) is a soft
// memory limit.
//
//   - If GOMEMLIMIT environment variable is specified, it is always used, and
//     cgroup limits are ignored. If GOMEMLIMIT is invalid, no changes to GOMEMLIMIT
//     are made.
//   - A percentage of maximum available memory(10%) is set as reserved.
//     This helps to avoid OOMs when only max memory is specified.
//   - If both max and high are positive and max - max*(reserved/100) is less than
//     high, GOMEMLIMIT is set to max - max*(reserved/100).
//   - If both max and high are non-zero and max - max*(reserved/100)
//     is greater than high, GOMEMLIMIT is set to high.
//   - If only max is positive, GOMEMLIMIT is set to max - max*(reserved/100).
//   - If only high is positive, GOMEMLIMIT is set to high.
//
// For example,
//   - For a workload with [MemoryMax]=250M and [MemoryHigh]=250M
//     GOMEMLIMIT is set to 235929600 bytes. (250 - 250*(10/100)) = 225M = 235929600.
//     [MemoryHigh] is ignored as [MemoryMax] - [WithMaxReservePercent] is less than
//     [MemoryHigh].
//   - For a workload with [MemoryHigh]=250M but no [MemoryMax] specified,
//     GOMEMLIMIT is set to 262144000 bytes.
//   - For a workload with [MemoryMax]=250M but no [MemoryHigh] specified,
//     GOMEMLIMIT is set to 235929600 bytes. (250 - 250*(10/100)) = 225M = 235929600.
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
// Set `GO_AUTOTUNE` environment variable to `false`.
//
// [CPU Management with static policy]: https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#using-cpu-management-with-static-policy
// [Vertical Pod autoscaling]: https://cloud.google.com/kubernetes-engine/docs/concepts/verticalpodautoscaler
// [memory.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [memory.high]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [cpu.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#core-interface-files
// [MemoryMax]: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryMax=bytes
// [MemoryHigh]: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryHigh=bytes
// [WithMaxReservePercent]: https://pkg.go.dev/github.com/tprasadtp/go-autotune/memlimit.WithMaxReservePercent
package autotune

import (
	"github.com/tprasadtp/go-autotune/internal/cgroup"
	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

//nolint:gochecknoinits // ignore
func init() {
	if !shared.IsFalse("GO_AUTOTUNE") && !shared.IsFalse("GOAUTOTUNE") {
		// To avoid parsing cgroup info twice.
		info, err := cgroup.GetInfo("", "")
		cpuQuotaFunc := func() (float64, error) {
			if err != nil {
				//nolint:wrapcheck // ignore
				return 0, err
			}
			return info.CPUQuota, nil
		}

		memlimitFunc := func() (int64, int64, error) {
			if err != nil {
				//nolint:wrapcheck // ignore
				return 0, 0, err
			}
			return info.MemoryMax, info.MemoryHigh, nil
		}

		maxprocs.Configure(maxprocs.WithCPUQuotaFunc(cpuQuotaFunc))
		memlimit.Configure(memlimit.WithMemoryQuotaFunc(memlimitFunc))
	}
}
