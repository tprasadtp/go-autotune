// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package scenarios

import (
	"math"
	"runtime"
	"runtime/debug"
	"testing"
)

// VerifyFunc returns a function which can be used by [Trampoline].
func VerifyFunc(cpus int, memory int64) func(tb testing.TB) {
	if cpus <= 0 {
		cpus = runtime.NumCPU()
	}

	if memory <= 0 {
		memory = math.MaxInt64
	}

	return func(tb testing.TB) {
		tb.Helper()
		v := runtime.GOMAXPROCS(-1)
		if v != cpus {
			tb.Errorf("GOMAXPROCS expected=%d, got=%d", cpus, v)
		}

		mv := debug.SetMemoryLimit(-1)
		if memory != mv {
			tb.Errorf("GOMEMLIMIT expected=%d, got=%d", memory, mv)
		}
	}
}
