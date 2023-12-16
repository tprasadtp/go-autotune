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
//   - If env variable GOMAXPROCS is set and is valid positive integer, it is always used.
//   - If running on Linux and cgroups v2 is available, CPU quota from current PID is
//     determined automatically and used to determine GOMAXPROCS.
//   - On non linux platforms only GOMAXPROCS env variable is considered.
//   - Fractional CPUs quotas are rounded off with [math.Ceil] by default,
//     unless overridden with [WithRoundFunc].
//   - If CPU quota is less than 1, GOMAXPROCS is set to 1.
//   - This function by default logs nothing. Custom logger can be specified via
//     [WithLogger].
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
