// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/tprasadtp/go-autotune/internal/cgroup"
	"github.com/tprasadtp/go-autotune/internal/discard"
	"github.com/tprasadtp/go-autotune/internal/shared"
)

type config struct {
	Logger         *slog.Logger
	ReservePercent uint64
	MemLimitFunc   func() (int64, int64, error)
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

	// If MemLimitFunc is not specified use default based on CGroupV2.
	if cfg.MemLimitFunc == nil {
		//nolint:nonamedreturns // for docs.
		cfg.MemLimitFunc = func() (max int64, high int64, err error) {
			info, err := cgroup.GetQuota()
			if err != nil {
				return 0, 0, fmt.Errorf("memlimit: failed to get memory limits: %w", err)
			}
			return info.MemoryMax, info.MemoryHigh, nil
		}
	}

	// Check if GOMEMLIMIT env variable is set.
	goMemLimitEnv := os.Getenv("GOMEMLIMIT")
	if goMemLimitEnv != "" {
		memLimitFromEnv, err := shared.Size(goMemLimitEnv)
		if err == nil && memLimitFromEnv > 0 {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
				"Setting GOMEMLIMIT from environment variable", slog.String("GOMEMLIMIT", goMemLimitEnv))
			debug.SetMemoryLimit(int64(memLimitFromEnv))
		} else {
			cfg.Logger.LogAttrs(ctx, slog.LevelError,
				"GOMEMLIMIT environment variable is invalid", slog.String("GOMEMLIMIT", goMemLimitEnv))
		}
		return
	}

	// Get memory limits.
	max, high, err := cfg.MemLimitFunc()
	if err != nil {
		cfg.Logger.LogAttrs(ctx, slog.LevelError, "Failed to get memory limits",
			slog.Any("err", err))
		return
	}

	cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Successfully obtained memory limits",
		slog.Int64("memory.max", max),
		slog.Int64("memory.high", high),
	)

	// Set default ReservePercent value and ignore invalid values.
	switch {
	case cfg.ReservePercent == 0:
		cfg.ReservePercent = 10
		cfg.Logger.LogAttrs(ctx, slog.LevelDebug,
			"Using default ReservePercent value", slog.Uint64("ReservePercent", cfg.ReservePercent))
	case cfg.ReservePercent < 1 || cfg.ReservePercent > 99:
		cfg.Logger.LogAttrs(ctx, slog.LevelError,
			"Ignoring ReservePercent out of bounds value", slog.Uint64("ReservePercent", cfg.ReservePercent))
		cfg.ReservePercent = 10
	}

	var memLimitBytes int64
	switch {
	// Both max and high are defined.
	case max > 0 && high > 0:
		// Calculate max with reserve applied.
		maxWithReserve := int64(float64(max) - float64(max)*float64((cfg.ReservePercent/100)))

		// Check if max with reserve is lower than high.
		if maxWithReserve < high {
			cfg.Logger.LogAttrs(ctx, slog.LevelError,
				"Max allowed memory (with reserve specified) is less than allowed high memory, using lower of the two",
				slog.Uint64("GOMEMLIMIT", cfg.ReservePercent))
			memLimitBytes = maxWithReserve
		} else {
			memLimitBytes = high
		}
	// Only max is specified
	case max > 0 && high <= 0:
		memLimitBytes = int64(float64(max) - float64(max)*float64((cfg.ReservePercent/100)))
	// Only high is specified
	case high > 0 && max <= 0:
		memLimitBytes = high
	}

	// Set GOMEMLIMIT.
	if memLimitBytes > 0 {
		snapshot := debug.SetMemoryLimit(-1)
		if snapshot != memLimitBytes {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
				"Setting GOMEMLIMIT", slog.String("GOMEMLIMIT", strconv.FormatInt(max, 10)))
			defer func() {
				err := recover()
				if err != nil {
					cfg.Logger.LogAttrs(ctx, slog.LevelError,
						"panic while setting SetMemoryLimit(GOMEMLIMIT), reverting the change",
						slog.Any("err", err))
					debug.SetMemoryLimit(snapshot)
				}
			}()
			debug.SetMemoryLimit(int64(max))
		} else {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "GOMEMLIMIT is already set",
				slog.String("GOMEMLIMIT", strconv.FormatInt(memLimitBytes, 10)),
			)
		}
	} else {
		cfg.Logger.LogAttrs(ctx, slog.LevelDebug, "No memory limits specified")
	}
}
