// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package maxprocs configures GOMAXPROCS.
package maxprocs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"runtime"
	"strconv"

	"github.com/tprasadtp/go-autotune/internal/discard"
)

type config struct {
	logger    *slog.Logger
	detector  CPUQuotaDetector
	roundFunc func(float64) int
}

// Current returns current GOMAXPROCS settings.
func Current() int {
	return runtime.GOMAXPROCS(-1)
}

// Configure configures GOMAXPROCS.
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
// For Windows containers with Hyper-V isolation, hypervisor emulates specified
// CPU cores, thus the default value of GOMAXPROCS is optimal and need not be changed.
//
// [cpu.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#core-interface-files
// [QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-queryinformationjobobject
// [CPU Management with static policy]: https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#using-cpu-management-with-static-policy
// [Vertical Pod autoscaling]: https://cloud.google.com/kubernetes-engine/docs/concepts/verticalpodautoscaler
func Configure(ctx context.Context, opts ...Option) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if ctx.Err() != nil {
		return fmt.Errorf("maxprocs: %w", ctx.Err())
	}

	// Apply all options.
	cfg := &config{}
	for i := range opts {
		if opts[i] != nil {
			opts[i].apply(cfg)
		}
	}

	// If logger is nil, use a logger backed by discard handler.
	if cfg.logger == nil {
		cfg.logger = slog.New(discard.NewHandler())
	}

	// If detector is nil, use default detector.
	if cfg.detector == nil {
		cfg.detector = DefaultCPUQuotaDetector()
	}

	// If rounding function is not specified, use math.Ceil
	if cfg.roundFunc == nil {
		cfg.roundFunc = func(f float64) int {
			return int(math.Ceil(f))
		}
	}

	snapshot := Current()

	// Check if GOMAXPROCS env variable is set.
	env := os.Getenv("GOMAXPROCS")
	if env != "" {
		maxProcsEnv, err := strconv.Atoi(env)
		if err == nil && maxProcsEnv > 0 {
			if snapshot != maxProcsEnv {
				cfg.logger.LogAttrs(ctx, slog.LevelInfo,
					"Setting GOMAXPROCS from environment variable",
					slog.String("GOMAXPROCS", env))
				runtime.GOMAXPROCS(maxProcsEnv)
			} else {
				cfg.logger.LogAttrs(ctx, slog.LevelInfo,
					"GOMAXPROCS is already set from environment variable",
					slog.String("GOMAXPROCS", env))
			}
			return nil
		}

		return fmt.Errorf("maxprocs: invalid GOMAXPROCS environment variable: %q", env)
	}

	// Get CPU quota.
	quota, err := cfg.detector.DetectCPUQuota(ctx)
	if err != nil {
		// Ignore unsupported platform error and do nothing.
		if errors.Is(err, errors.ErrUnsupported) {
			return nil
		}
		cfg.logger.LogAttrs(ctx, slog.LevelError, "Failed to obtain cpu quota",
			slog.Any("err", err),
		)
		return fmt.Errorf("maxprocs: %w", err)
	}

	if quota <= 0 {
		cfg.logger.LogAttrs(ctx, slog.LevelInfo, "CPU quota is not defined")
		return nil
	}

	cfg.logger.LogAttrs(ctx, slog.LevelInfo, "Successfully obtained cpu quota",
		slog.Float64("cpu.quota", quota),
	)

	// Round off fractional CPU using defined RoundFunc. Default is math.Ceil.
	procs := cfg.roundFunc(quota)

	if procs < 0 {
		return fmt.Errorf("maxprocs: RoundFunc returned negative value: %d", procs)
	}

	// GOMAXPROCS ensure at-least 1
	if procs < 1 {
		cfg.logger.LogAttrs(ctx, slog.LevelInfo, "Selecting minimum possible GOMAXPROCS value")
		procs = 1
	}

	if snapshot != procs {
		cfg.logger.LogAttrs(ctx, slog.LevelInfo, "Setting GOMAXPROCS",
			slog.String("GOMAXPROCS", strconv.FormatInt(int64(procs), 10)),
		)
		runtime.GOMAXPROCS(procs)
	} else {
		cfg.logger.LogAttrs(ctx, slog.LevelInfo, "GOMAXPROCS is already set",
			slog.String("GOMAXPROCS", strconv.FormatInt(int64(procs), 10)),
		)
	}
	return nil
}
