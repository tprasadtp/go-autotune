// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune_test

import (
	"math"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/autotune"
)

func BenchmarkConfigure(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Run Configure.
		autotune.Configure()

		// Reset GOMAXPROCS and GOMEMLIMIT
		b.StopTimer()
		runtime.GOMAXPROCS(runtime.NumCPU())
		debug.SetMemoryLimit(math.MaxInt64)

		// Restart the timer.
		b.StartTimer()
	}
}
