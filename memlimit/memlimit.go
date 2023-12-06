// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/tprasadtp/go-autotune/internal/discard"
	"github.com/tprasadtp/go-autotune/internal/parse"
)

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

	// Check if GOMEMLIMIT env variable is set.
	goMemLimitEnv := os.Getenv("GOMEMLIMIT")
	if goMemLimitEnv != "" {
		memLimitFromEnv, err := parse.Size(goMemLimitEnv)
		if err == nil && memLimitFromEnv > 0 {
			cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
				"Setting GOMEMLIMIT from environment variable", slog.String("GOMEMLIMIT", goMemLimitEnv))
			debug.SetMemoryLimit(int64(memLimitFromEnv))
			return
		}

		// GOMEMLIMIT is invalid.
		cfg.Logger.LogAttrs(ctx, slog.LevelInfo,
			"Ignoring invalid GOMEMLIMIT environment variable", slog.String("GOMEMLIMIT", goMemLimitEnv))
	}
}
