// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package maxprocs

import (
	"context"
	"github.com/tprasadtp/go-autotune/internal/discard"
	"log/slog"
	"os"
	"runtime"
	"strconv"
)

// Current returns current GOMAXPROCS settings.
func Current() int {
	return runtime.GOMAXPROCS(0)
}

// Configure configures GOMAXPROCS.
//
//   - If env variable GOMAXPROCS is set and is valid positive integer, it is always used.
//   - If platform is Linux and cgroups v2 is available, CPU quota from current PID is
//     determined automatically and used to determine GOMAXPROCS.
//   - On non linux platforms only GOMAXPROCS env variable is considered.
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

	// Check if GOMAXPROCS env variable is set.
	goMaxProcsEnv := os.Getenv("GOMAXPROCS")
	if goMaxProcsEnv != "" {
		maxProcsEnvParse, err := strconv.Atoi(goMaxProcsEnv)
		if err == nil && maxProcsEnvParse > 0 {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
				"Setting GOMAXPROCS from environment variable", slog.String("GOMAXPROCS", goMaxProcsEnv))
			runtime.GOMAXPROCS(maxProcsEnvParse)
			return
		}

		// GOMAXPROCS is invalid.
		cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
			"Ignoring invalid GOMAXPROCS environment variable", slog.String("GOMAXPROCS", goMaxProcsEnv))
	}
}
