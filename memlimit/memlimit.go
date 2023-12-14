// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/tprasadtp/go-autotune/internal/discard"
	"github.com/tprasadtp/go-autotune/internal/platform"
	"github.com/tprasadtp/go-autotune/internal/shared"
)

type config struct {
	Logger            *slog.Logger
	MaxReservePercent uint8
	QuotaFunc         func() (int64, int64, error)
}

// Current returns current memelimit in bytes.
func Current() int64 {
	return debug.SetMemoryLimit(-1)
}

// Configure configures GOMEMLIMIT.
//
//   - If env variable GOMEMLIMIT is set, it is always used. Invalid value will be ignored.
//   - If running on Linux and cgroups v2 is available, memory quota for the current
//     process is determined automatically and used to determine GOMEMLIMIT.
//   - On non linux platforms only GOMEMLIMIT env variable is considered.
//
// This function prefers using soft memory limit whenever possible.
// cgroup memory limit [memory.max](referred from here onwards as max) is a hard
// memory limit and [memory.high](referred from here onwards as high) is a soft
// memory limit.
//
//   - A percentage of maximum available memory is set as reserved.
//     This helps to avoid OOMs when only max memory is specified.
//     Be default 10% is set as reserved for max < 5Gi and 5% for max >= 5Gi.
//   - If both max and high are positive and max - max*(reserved/100) is less than
//     high, GOMEMLIMIT is set to max - max*(reserved/100).
//   - If both max and high are positive and max - max*(reserved/100)
//     is greater than high, GOMEMLIMIT is set to high.
//   - If only max is positive, GOMEMLIMIT is set to max - max*(reserved/100).
//   - If only high is positive, GOMEMLIMIT is set to high.
//
// [memory.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [memory.high]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
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

	// If MemLimitFunc is not specified use default.
	if cfg.QuotaFunc == nil {
		//nolint:nonamedreturns // for docs.
		cfg.QuotaFunc = func() (max int64, high int64, err error) {
			max, high, err = platform.GetMemoryQuota()
			if err != nil {
				return 0, 0, fmt.Errorf("memlimit: failed to get memory limits: %w", err)
			}
			return max, high, nil
		}
	}

	// Check if GOMEMLIMIT env variable is set.
	env := os.Getenv("GOMEMLIMIT")
	if env != "" {
		limit, err := shared.ParseSize(env)
		if err == nil && limit > 0 {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
				"Setting GOMEMLIMIT from environment variable",
				slog.String("GOMEMLIMIT", env))
			snapshot := debug.SetMemoryLimit(-1)
			if snapshot != limit {
				cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
					"Setting GOMEMLIMIT",
					slog.String("GOMEMLIMIT", strconv.FormatInt(limit, 10)))
				defer func() {
					err := recover()
					if err != nil {
						cfg.Logger.LogAttrs(ctx, slog.LevelError,
							"Panic while setting GOMEMLIMIT, reverting the change",
							slog.Any("err", err))
						debug.SetMemoryLimit(snapshot)
					}
				}()
				debug.SetMemoryLimit(limit)
			}
		} else {
			cfg.Logger.LogAttrs(ctx, slog.LevelError,
				"GOMEMLIMIT environment variable is invalid", slog.String("GOMEMLIMIT", env))
		}
		return
	}

	// Get memory limits.
	max, high, err := cfg.QuotaFunc()
	if err != nil {
		cfg.Logger.LogAttrs(ctx, slog.LevelError, "Failed to get memory limits",
			slog.Any("err", err))
		return
	}

	// Set default ReservePercent value and ignore invalid values.
	switch {
	case cfg.MaxReservePercent == 0:
		if max >= 5*shared.GiByte {
			cfg.MaxReservePercent = 5
		} else {
			cfg.MaxReservePercent = 10
		}
	case cfg.MaxReservePercent > 99:
		cfg.Logger.LogAttrs(ctx, slog.LevelError, "Ignoring invalid reserve percentage value",
			slog.Uint64("memory.reserve.percent", uint64(cfg.MaxReservePercent)),
		)

		if max >= 5*shared.GiByte {
			cfg.MaxReservePercent = 5
		} else {
			cfg.MaxReservePercent = 10
		}
	}
	reserve := int64(math.Ceil(float64(max) * (float64(cfg.MaxReservePercent) / 100)))
	if max <= 0 && high <= 0 {
		cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Memory limits not specified")
		return
	}

	cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Successfully obtained memory limits",
		slog.Int64("memory.max", max),
		slog.Int64("memory.high", high),
		slog.Int64("memory.reserve.bytes", reserve),
		slog.Uint64("memory.reserve.percent", uint64(cfg.MaxReservePercent)),
	)

	var limit int64
	switch {
	// Both max and high are defined.
	case max > 0 && high > 0:
		// Check if max - reserve is lower than high.
		if max-reserve < high {
			limit = max - reserve
		} else {
			limit = high
		}
	// Only max is specified
	case max > 0 && high <= 0:
		limit = max - reserve
	// Only high is specified
	case high > 0 && max <= 0:
		limit = high
	}

	// Set GOMEMLIMIT.
	if limit > 0 {
		snapshot := debug.SetMemoryLimit(-1)
		if snapshot != limit {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
				"Setting GOMEMLIMIT", slog.String("GOMEMLIMIT", strconv.FormatInt(limit, 10)))
			defer func() {
				err := recover()
				if err != nil {
					cfg.Logger.LogAttrs(ctx, slog.LevelError,
						"Panic while setting GOMEMLIMIT, reverting the change",
						slog.Any("err", err))
					debug.SetMemoryLimit(snapshot)
				}
			}()
			debug.SetMemoryLimit(limit)
		}
	}
}
