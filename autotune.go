// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package autotune automatically configures GOMAXPROCS and GOMEMLIMIT.
//
// Importing this package will automatically set GOMAXPROCS and GOMEMLIMIT
// via [runtime.GOMAXPROCS] and [runtime/debug.SetMemoryLimit] respectively,
// taking into consideration CPU quota, memory limits from assigned cgroup.
//
// This will always respect valid GOMAXPROCS and GOMEMLIMIT environment
// variables. To disable automatic configuration based on cgroup limits/quotas
// at runtime, set GO_AUTOTUNE environment variable to false.
//
// If you re using Vertical Pod Autoscaler, and do not wish to encounter CPU
// throttling it is recommended that you use [CPU Management with static policy],
// to ensure CPU recommendation is an integer.
//
// Libraries should avoid importing this package and it should only be imported by
// the main package.
//
// By default logging is disabled, to debug issues, use your own init function.
// See examples for more info.
//
// This package MUST NOT be used with other packages which tweak GOMAXPROCS and GOMEMLIMIT.
// Some known incompatible packages include,
//
//   - [go.uber.org/autmaxprocs]
//   - [github.com/KimMachineGun/automemlimit]
//
// On non linux platforms importing this package has no effect.
//
//   - See [github.com/tprasadtp/go-autotune/maxprocs] for configuring GOMAXPROCS.
//   - See [github.com/tprasadtp/go-autotune/memlimit] for configuring GOMEMLIMIT.
//
// [CPU Management with static policy]: https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#using-cpu-management-with-static-policy
package autotune
