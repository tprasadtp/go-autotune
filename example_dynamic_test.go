// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune_test

import (
	"context"
	"errors"
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

// This example checks resource limits and sets GOMAXPROCS and GOMEMLIMIT
// at specified intervals. This may be useful for cases where resource limits
// are expected to change.
//
// See https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/
// for kubernetes in place resource resize documentation.
//
// If GOAUTOTUNE env variable set to false value, then GOMAXPROCS and GOMEMLIMIT
// are not modified.
func Example_inPlaceResourceResize() {
	// WaitGroup to wait on background goroutines.
	var wg sync.WaitGroup

	// Stop goroutines on signals.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Change this interval to what your application requires.
	// Avoid re-configuring too many times.
	interval := time.Second * 10

	switch strings.ToLower(os.Getenv("GOAUTOTUNE")) {
	case "0", "off", "disable", "disabled", "no", "false":
		slog.Info("Automatic resource limit configuration is disabled")
	default:
		err := maxprocs.Configure(
			ctx,
			maxprocs.WithLogger(slog.Default()),
		)
		slog.Error("Failed to configure GOMAXPROCS", slog.Any("err", err))
		err = memlimit.Configure(
			ctx,
			memlimit.WithLogger(slog.Default()),
		)
		slog.Error("Failed to configure GOMEMLIMIT", slog.Any("err", err))

		ticker := time.NewTicker(interval)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					ticker.Stop()
					slog.Info("Stopping background tasks...")
					return
				case <-ticker.C:
					err := maxprocs.Configure(
						ctx,
						maxprocs.WithLogger(slog.Default()),
					)
					if !errors.Is(err, context.Canceled) {
						slog.Error("Failed to configure GOMAXPROCS", slog.Any("err", err))
					}
					err = memlimit.Configure(
						ctx,
						memlimit.WithLogger(slog.Default()),
					)
					if !errors.Is(err, context.Canceled) {
						slog.Error("Failed to configure GOMEMLIMIT", slog.Any("err", err))
					}
				}
			}
		}()
	}

	// Wait for all background tasks to complete.
	slog.Info("Waiting for background tasks to complete...")
	wg.Wait()
}
