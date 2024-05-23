// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package autotune

import (
	"context"
	"log/slog"

	"github.com/tprasadtp/go-autotune/internal/env"
	"github.com/tprasadtp/go-autotune/internal/quota"
	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func configure() {
	if env.IsFalse("GO_AUTOTUNE") || env.IsFalse("GOAUTOTUNE") {
		return
	}

	var logger *slog.Logger
	if env.IsDebug("GO_AUTOTUNE") || env.IsDebug("GOAUTOTUNE") {
		logger = slog.Default()
	}

	// To avoid parsing mountinfo and cgroup file twice,
	// get cgroup interface path for current process' cgroup
	// and re-use it.
	cgroupfs, err := quota.GetCgroupInterfacePath("")
	if err != nil {
		return
	}
	detector := quota.NewDetectorWithCgroupPath(cgroupfs)
	ctx := context.Background()

	_ = maxprocs.Configure(ctx,
		maxprocs.WithLogger(logger),
		maxprocs.WithCPUQuotaDetector(detector),
	)
	_ = memlimit.Configure(ctx,
		memlimit.WithLogger(logger),
		memlimit.WithMemoryQuotaDetector(detector),
	)
}
