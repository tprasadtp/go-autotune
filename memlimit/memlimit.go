// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package memlimit configures GOMEMLIMIT.
package memlimit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/tprasadtp/go-autotune/internal/discard"
	"github.com/tprasadtp/go-autotune/internal/shared"
)

type config struct {
	logger   *slog.Logger
	detector MemoryQuotaDetector
	reserve  int64
}

// Current returns current GOMEMLIMIT in bytes.
func Current() int64 {
	return debug.SetMemoryLimit(-1)
}

// Configure configures GOMEMLIMIT.
//
// If GOMEMLIMIT environment variable is specified, it is ALWAYS used, and limits are
// ignored. If GOMEMLIMIT environment variable is invalid, runtime MAY panic. Otherwise
// this package will attempt to detect defined memory limits using platform specific APIs.
//
// Memory limits can be soft limit, or hard limit. Hard memory limit cannot be breached
// by the process and typically leads to OOM killer being invoked for the process group/process
// when it is exceeded. For this reason, to let garbage collector free up memory early before
// OOM killer is involved, a small percentage of hard memory limit is set aside as reserved.
// This memory is fully accessible to the process and the runtime, but acts as a hint to the
// garbage collector. By default, 10% is set as reserved, hard memory limit is less than 5Gi
// and 5% otherwise.
//
// For Linux, cgroup v2 interface files are used to get memory limits.
// cgroup memory limit [memory.max] is hard memory limit and [memory.high] is
// soft memory limit. If using soft memory limits, an external process SHOULD monitor
// pressure stall information of the workload/cgroup AND alleviate the reclaim pressure.
// If your workload manager defines [memory.high], but you wish to only use [memory.max]
// use [WithIgnoreSoftLimit].
//
//   - If both [memory.max] and [memory.high] are specified, and ([memory.max] - reserved)
//     is less than [memory.high], GOMEMLIMIT is set to ([memory.max] - reserved).
//   - If both [memory.max] and [memory.high] limits are specified, and ([memory.max] - reserved)
//     is greater than [memory.high], GOMEMLIMIT is set to [memory.high].
//   - If only [memory.max] is specified, GOMEMLIMIT is set to ([memory.max] - reserved).
//   - If only [memory.high] limit is specified, GOMEMLIMIT is set to [memory.high].
//
// For Windows, [QueryInformationJobObject] API is used to get memory limits.
// [JOBOBJECT_EXTENDED_LIMIT_INFORMATION] defines per process(ProcessMemoryLimit)
// and per job memory limits(JobMemoryLimit). ProcessMemoryLimit is always preferred
// over JobMemoryLimit. Both are considered hard limits.
//
// [memory.max]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [memory.high]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
// [QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-queryinformationjobobject
// [JOBOBJECT_EXTENDED_LIMIT_INFORMATION]: https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_extended_limit_information
func Configure(ctx context.Context, opts ...Option) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return fmt.Errorf("memlimit: %w", ctx.Err())
	}

	cfg := &config{
		reserve: -1,
	}

	// Apply all options.
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
		cfg.detector = DefaultMemoryQuotaDetector()
	}

	var limit int64
	var err error

	// Get current value of memory limit.
	snapshot := debug.SetMemoryLimit(-1)

	// Check if GOMEMLIMIT env variable.
	env := os.Getenv("GOMEMLIMIT")
	if env != "" {
		// Value "off" does not appear to be documented but is part of implementation.
		// preserve the same behavior.
		//
		// https://go.googlesource.com/go/+/refs/tags/go1.22.3/src/runtime/mgcpacer.go#1323
		if env == "off" {
			limit = math.MaxInt64
		} else {
			limit, err = shared.ParseMemlimit(env)
			if err != nil {
				cfg.logger.LogAttrs(ctx, slog.LevelError,
					"GOMEMLIMIT environment variable is invalid",
					slog.String("GOMEMLIMIT", env),
				)
				return fmt.Errorf("GOMEMLIMIT environment variable(%q) is invalid", env)
			}
		}

		// Set GOMEMLIMIT from env variable.
		cfg.logger.LogAttrs(ctx, slog.LevelInfo,
			"Setting GOMEMLIMIT from environment variable",
			slog.String("GOMEMLIMIT", env))
		if snapshot != limit {
			cfg.logger.LogAttrs(ctx, slog.LevelInfo,
				"Setting GOMEMLIMIT from environment variable",
				slog.String("GOMEMLIMIT", env))
			debug.SetMemoryLimit(limit)
		} else {
			cfg.logger.LogAttrs(ctx, slog.LevelInfo,
				"GOMEMLIMIT is already set from environment variable",
				slog.String("GOMEMLIMIT", env))
		}
		return nil
	}

	// Get memory limits.
	var hard int64
	var soft int64
	hard, soft, err = cfg.detector.DetectMemoryQuota(ctx)
	if err != nil {
		// Ignore unsupported platform error and do nothing.
		if errors.Is(err, errors.ErrUnsupported) {
			return nil
		}

		cfg.logger.LogAttrs(ctx, slog.LevelError,
			"Failed to get memory limits",
			slog.Any("err", err))
		return fmt.Errorf("memlimit: %w", err)
	}

	// Set default ReservePercent value and ignore invalid values.
	// If MaxReservePercent is less than 0, use default values.
	switch {
	case cfg.reserve < 0:
		if hard >= 5*shared.GiByte {
			cfg.reserve = 5
		} else {
			cfg.reserve = 10
		}
	case cfg.reserve >= 100:
		cfg.logger.LogAttrs(ctx, slog.LevelError, "Invalid reserve percentage value",
			slog.Uint64("memory.reserve.percent", uint64(cfg.reserve)),
		)

		if hard >= 5*shared.GiByte {
			cfg.reserve = 5
		} else {
			cfg.reserve = 10
		}
	}
	reserve := int64(math.Ceil(float64(hard) * (float64(cfg.reserve) / 100)))
	if hard <= 0 && soft <= 0 {
		cfg.logger.LogAttrs(ctx, slog.LevelInfo, "Memory limits not specified")
		return nil
	}

	cfg.logger.LogAttrs(ctx, slog.LevelInfo, "Successfully obtained memory limits",
		slog.Int64("memlimit.hard", hard),
		slog.Int64("memlimit.soft", soft),
		slog.Int64("memlimit.reserve.bytes", reserve),
		slog.Uint64("memlimit.reserve.percent", uint64(cfg.reserve)),
	)

	switch {
	// Both hard and soft memory limits are defined.
	case hard > 0 && soft > 0:
		// Check if hard - reserve is lower than soft.
		if hard-reserve < soft {
			limit = hard - reserve
		} else {
			limit = soft
		}
	// Only hard memory limit is specified.
	case hard > 0:
		limit = hard - reserve
	// Only soft memory limit is specified.
	case soft > 0:
		limit = soft
	}

	// Set GOMEMLIMIT.
	if limit > 0 {
		if snapshot != limit {
			cfg.logger.LogAttrs(ctx, slog.LevelInfo,
				"Setting GOMEMLIMIT", slog.String("GOMEMLIMIT", strconv.FormatInt(limit, 10)))
			debug.SetMemoryLimit(limit)
		}
		return nil
	}

	cfg.logger.LogAttrs(ctx, slog.LevelInfo, "Memory limits are not defined")
	return nil
}
