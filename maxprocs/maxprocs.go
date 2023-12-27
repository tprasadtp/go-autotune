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
	"github.com/tprasadtp/go-autotune/internal/platform"
)

type config struct {
	Logger       *slog.Logger
	RoundFunc    func(float64) int
	CPUQuotaFunc func() (float64, error)
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
// For Windows containers with hyper-v isolation, hypervisor emulates specified
// CPU cores, thus the default value of GOMAXPROCS is optimal and need not be changed.
//
// [cpu.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#core-interface-files
// [QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-queryinformationjobobject
func Configure(opts ...Option) {
	cfg := &config{}
	ctx := context.Background()

	// Apply all options.
	for i := range opts {
		if opts[i] != nil {
			opts[i].apply(cfg)
		}
	}

	// If logger is nil, use a discard logger.
	if cfg.Logger == nil {
		cfg.Logger = slog.New(discard.NewHandler())
	}

	// If CPUQuotaFunc is not specified, use default based on CGroupV2.
	if cfg.CPUQuotaFunc == nil {
		cfg.CPUQuotaFunc = func() (float64, error) {
			v, err := platform.GetCPUQuota()
			if err != nil {
				return -1, fmt.Errorf("maxprocs: failed to get cpu quota: %w", err)
			}
			return v, nil
		}
	}

	// If rounding function is not specified, use math.ceil
	if cfg.RoundFunc == nil {
		cfg.RoundFunc = func(f float64) int {
			if f < 0 {
				return 0
			}
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
				cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
					"Setting GOMAXPROCS from environment variable",
					slog.String("GOMAXPROCS", env))
				runtime.GOMAXPROCS(maxProcsEnv)
			}
		} else {
			cfg.Logger.LogAttrs(ctx, slog.LevelError,
				"Ignoring invalid GOMAXPROCS environment variable",
				slog.String("GOMAXPROCS", env))
		}
		return
	}

	// Get cgroup Info.
	quota, err := cfg.CPUQuotaFunc()
	if err != nil {
		// Log if error is not [errors.ErrUnsupported].
		//
		// This ensures non linux platforms do not log anything.
		if !errors.Is(err, errors.ErrUnsupported) {
			cfg.Logger.LogAttrs(ctx, slog.LevelError, "Failed to get cpu quota",
				slog.Any("err", err))
		}

		return
	}

	if quota > 0 {
		cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Successfully obtained cpu quota",
			slog.Float64("cpu.quota", quota),
		)

		// Round off fractional CPU using defined RoundFunc.
		// Default is math.Ceil.
		procs := cfg.RoundFunc(quota)

		// GOMAXPROCS ensure at-least 1
		if procs < 1 {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Selecting minimum possible GOMAXPROCS value")
			procs = 1
		}
		if snapshot != procs {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Setting GOMAXPROCS",
				slog.String("GOMAXPROCS", strconv.FormatInt(int64(procs), 10)),
			)
			runtime.GOMAXPROCS(procs)
		} else {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "GOMAXPROCS is already set",
				slog.String("GOMAXPROCS", strconv.FormatInt(int64(procs), 10)),
			)
		}
	} else {
		cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "cpu quota is not defined")
	}
}
