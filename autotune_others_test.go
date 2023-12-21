// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux && !windows

package autotune_test

import (
	"math"
	"os"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func TestOthers(t *testing.T) {
	t.Run("GOMAXPROCS", func(t *testing.T) {
		if os.Getenv("GOMAXPROCS") != "" {
			t.Skipf("GOMAXPROCS env variable is set")
		}

		v := maxprocs.Current()
		if v != runtime.NumCPU() {
			t.Errorf("expected=%d, got=%d", runtime.NumCPU(), v)
		}
	})

	t.Run("GOMEMLIMIT", func(t *testing.T) {
		if os.Getenv("GOMEMLIMIT") != "" {
			t.Skipf("GOMEMLIMIT env variable is set")
		}
		v := memlimit.Current()
		if v != math.MaxInt64 {
			t.Errorf("expected=%d, got=%d", math.MinInt64, v)
		}
	})
}
