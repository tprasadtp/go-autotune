// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit_test

import (
	"context"
	"log/slog"

	"github.com/tprasadtp/go-autotune/memlimit"
)

// This example wraps default memory quota detection algorithm, but ignores
// soft memory limits and only considers hard memory limit.
func ExampleDefaultMemoryQuotaDetector() {
	detector := memlimit.DefaultMemoryQuotaDetector()
	logger := slog.Default()
	ctx := context.Background()
	//nolint:nonamedreturns // for docs.
	wrapper := memlimit.MemoryQuotaDetectorFunc(func(ctx context.Context) (max, high int64, err error) {
		max, high, err = detector.DetectMemoryQuota(ctx)
		if err == nil {
			return max, 0, nil
		}
		return max, high, err
	})
	err := memlimit.Configure(ctx,
		memlimit.WithMemoryQuotaDetector(wrapper), memlimit.WithLogger(logger))
	if err != nil {
		panic(err)
	}
}
