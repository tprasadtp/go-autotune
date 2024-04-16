// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune_test

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

// AutotuneInBackground checks resource limits and sets GOMAXPROCS and GOMEMLIMIT
// at specified intervals. This may be useful for in place resource resize
// See https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/
// for kubernetes docs.
//
// If GOAUTOTUNE env variable set to false value, disable it.
func AutotuneInBackground(ctx context.Context, wg *sync.WaitGroup, logger *slog.Logger, interval time.Duration) {
	switch strings.ToLower(os.Getenv("GOAUTOTUNE")) {
	case "0", "off", "disable", "disabled", "no", "false":
		logger.LogAttrs(ctx, slog.LevelInfo, "Automatic resource limit configuration is disabled")
		return
	}

	// Initial configuration.
	maxprocs.Configure(
		maxprocs.WithLogger(logger),
	)
	memlimit.Configure(
		memlimit.WithLogger(logger),
	)

	if interval >= time.Second {
		ticker := time.NewTicker(interval)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					logger.Info("Stopping background autotune task")
					ticker.Stop()
					return
				case <-ticker.C:
					maxprocs.Configure(
						maxprocs.WithLogger(logger),
					)
					memlimit.Configure(
						memlimit.WithLogger(logger),
					)
				}
			}
		}()
	}
}

func Example_inPlaceResourceResize() {
	var wg sync.WaitGroup

	// Stop goroutines on signals.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Change this interval to what your application requires.
	// Avoid re-configuring too many times.
	interval := time.Second * 30
	AutotuneInBackground(ctx, &wg, slog.Default(), interval)

	// Wait for all background tasks to complete.
	slog.Info("Waiting for background tasks to complete...")
	wg.Wait()
}
