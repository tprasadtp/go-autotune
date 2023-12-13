// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux && !windows

package maxprocs_test

import (
	"log/slog"
	"os"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/maxprocs"
)

func TestConfigure_UnsupportedPlatform(t *testing.T) {
	reset()
	os.Unsetenv("GOMAXPROCS")
	maxprocs.Configure(maxprocs.WithLogger(slog.Default()))
	v := maxprocs.Current()
	cpu := runtime.NumCPU()
	if v != cpu {
		t.Errorf("GOMAXPROCS expected=%d, got=%d", cpu, v)
	}
}
