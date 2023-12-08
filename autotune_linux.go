// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

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
		// To avoid computing limits and parsing cgroup info twice.
		info, err := cgroup.GetInfo("", "")
		cpuQuotaFunc := func() (float64, error) {
			if err != nil {
				//nolint:wrapcheck // ignore
				return 0, err
			}
			return info.CPUQuota, nil
		}

		memlimitFunc := func() (uint64, uint64, error) {
			if err != nil {
				//nolint:wrapcheck // ignore
				return 0, 0, err
			}
			return info.MemoryMax, info.MemoryHigh, nil
		}

		maxprocs.Configure(maxprocs.WithCPUQuotaFunc(cpuQuotaFunc))
		memlimit.Configure(memlimit.WithMemLimitFunc(memlimitFunc))
	}
}
