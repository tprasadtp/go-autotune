// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package memlimit configures GOMEMLIMIT.
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
	"github.com/tprasadtp/go-autotune/internal/parse"
	"github.com/tprasadtp/go-autotune/internal/platform"
)

type config struct {
	MaxReservePercent int64
	Logger            *slog.Logger
	QuotaFunc         func() (int64, int64, error)
}

// Current returns current GOMEMLIMIT in bytes.
func Current() int64 {
	return debug.SetMemoryLimit(-1)
}

// Configure configures GOMEMLIMIT.
//
// Memory limits can be soft memory limit, or hard memory limit.
//
// For Linux, cgroup v2 interface files are used to get memory limits.
// cgroup memory limit [memory.max] is hard memory limit and [memory.high] is
// soft memory limit. If using soft memory limits, an external process SHOULD monitor
// pressure stall information of the workload/cgroup AND alleviate the reclaim pressure.
//
// For Windows, [QueryInformationJobObject] API is used to get memory limits.
// [JOBOBJECT_EXTENDED_LIMIT_INFORMATION] defines per process(ProcessMemoryLimit)
// and per job memory limits(JobMemoryLimit). ProcessMemoryLimit is always preferred
// over JobMemoryLimit. Both are considered hard limits.
//
//   - If GOMEMLIMIT environment variable is specified, it is ALWAYS used, and limits are
//     ignored. If GOMEMLIMIT environment variable is invalid, runtime MAY panic.
//   - A percentage of hard memory limit is set as reserved. This helps to avoid OOMs.
//     By default, 10% is set as reserved, if limit is less than 5Gi and 5% otherwise.
//   - If both hard and soft memory limits are specified, and (hard memory limit - reserved)
//     is less than soft memory limit, GOMEMLIMIT is set to (hard memory limit - reserved).
//   - If both hard and soft memory limits are specified, and (hard memory limit - reserved)
//     is greater than soft memory limit, GOMEMLIMIT is set to soft memory limit.
//   - If only hard memory limit is specified, GOMEMLIMIT is set to (hard memory limit - reserved).
//   - If only soft memory limit is specified, GOMEMLIMIT is set to soft memory limit.
//
// [memory.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [memory.high]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-queryinformationjobobject
// [JOBOBJECT_EXTENDED_LIMIT_INFORMATION]: https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_extended_limit_information
func Configure(opts ...Option) {
	cfg := &config{
		MaxReservePercent: -1,
	}
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
		limit, err := parse.Memlimit(env)
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
	// If MaxReservePercent is less than 0, use default values.
	switch {
	case cfg.MaxReservePercent < 0:
		if max >= 5*parse.GiByte {
			cfg.MaxReservePercent = 5
		} else {
			cfg.MaxReservePercent = 10
		}
	case cfg.MaxReservePercent > 99:
		cfg.Logger.LogAttrs(ctx, slog.LevelError, "Invalid reserve percentage value",
			slog.Uint64("memory.reserve.percent", uint64(cfg.MaxReservePercent)),
		)

		if max >= 5*parse.GiByte {
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
