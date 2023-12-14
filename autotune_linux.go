// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune

import (
	"log/slog"

	"github.com/tprasadtp/go-autotune/internal/platform"
	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

//nolint:gochecknoinits // ignore
func init() {
	if shared.IsFalse("GO_AUTOTUNE") || shared.IsFalse("GOAUTOTUNE") {
		return
	}

	var logger *slog.Logger
	if shared.IsDebug("GO_AUTOTUNE") || shared.IsDebug("GOAUTOTUNE") {
		logger = slog.Default()
	}

	// To avoid parsing mountinfo and cgroup file twice,
	// get cgroup interface path for current process' cgroup
	// and re-use it.
	cgroupPath, err := platform.GetCgroupInterfacePath()
	cpuQuotaFunc := func() (float64, error) {
		if err != nil {
			//nolint:wrapcheck // ignore
			return 0, err
		}
		//nolint:wrapcheck // ignore
		return platform.GetCPUQuota(platform.WithCgroupInterfacePath(cgroupPath))
	}

	memQuotaFunc := func() (int64, int64, error) {
		if err != nil {
			//nolint:wrapcheck // ignore
			return 0, 0, err
		}
		//nolint:wrapcheck // ignore
		return platform.GetMemoryQuota(platform.WithCgroupInterfacePath(cgroupPath))
	}

	maxprocs.Configure(
		maxprocs.WithLogger(logger),
		maxprocs.WithCPUQuotaFunc(cpuQuotaFunc),
	)
	memlimit.Configure(
		memlimit.WithLogger(logger),
		memlimit.WithMemoryQuotaFunc(memQuotaFunc),
	)
}
