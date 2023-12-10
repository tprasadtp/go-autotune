// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

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

	"github.com/tprasadtp/go-autotune/internal/cgroup"
	"github.com/tprasadtp/go-autotune/internal/discard"
)

type config struct {
	Logger       *slog.Logger
	RoundFunc    func(float64) uint64
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
		cfg.Logger = slog.New(discard.NewDiscardHandler())
	}

	// If CPUQuotaFunc is not specified, use default based on CGroupV2.
	if cfg.CPUQuotaFunc == nil {
		cfg.CPUQuotaFunc = func() (float64, error) {
			info, err := cgroup.GetQuota()
			if err != nil {
				return -1, fmt.Errorf("maxprocs: failed to get cpu quota: %w", err)
			}
			return info.CPU, nil
		}
	}

	// If rounding function is not specified, use math.ceil
	if cfg.RoundFunc == nil {
		cfg.RoundFunc = func(f float64) uint64 {
			if f < 0 {
				return 0
			}
			return uint64(math.Ceil(f))
		}
	}

	snapshot := Current()

	// Check if GOMAXPROCS env variable is set.
	goMaxProcsEnv := os.Getenv("GOMAXPROCS")
	if goMaxProcsEnv != "" {
		maxProcsEnvParse, err := strconv.Atoi(goMaxProcsEnv)
		if err == nil && maxProcsEnvParse > 0 {
			if snapshot != maxProcsEnvParse {
				cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
					"Setting GOMAXPROCS from environment variable",
					slog.String("GOMAXPROCS", goMaxProcsEnv))
				runtime.GOMAXPROCS(maxProcsEnvParse)
			} else {
				cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "GOMAXPROCS is already set",
					slog.String("GOMAXPROCS", goMaxProcsEnv),
				)
			}
		} else {
			cfg.Logger.LogAttrs(ctx, slog.LevelError,
				"Ignoring invalid GOMAXPROCS environment variable",
				slog.String("GOMAXPROCS", goMaxProcsEnv))
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
			cfg.Logger.LogAttrs(ctx, slog.LevelError, "Failed to get CPU quota info",
				slog.Any("err", err))
		}

		return
	}

	if quota > 0 {
		cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Successfully obtained CPU Quota",
			slog.Float64("CPUQuota", quota),
		)
		// Round off fractional CPU using defined RoundFunc.
		// Default is math.Ceil.
		procs := cfg.RoundFunc(quota)

		// GOMAXPROCS ensure at-least 1
		if procs < 1 {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Selecting minimum possible GOMAXPROCS value")
			procs = 1
		}
		if snapshot != int(procs) {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Setting GOMAXPROCS",
				slog.String("GOMAXPROCS", strconv.FormatUint(procs, 10)),
			)
			runtime.GOMAXPROCS(int(procs))
		} else {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "GOMAXPROCS is already set",
				slog.String("GOMAXPROCS", strconv.FormatUint(procs, 10)),
			)
		}
	} else {
		cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "CPUQuota is not defined/unlimited")
	}
}
