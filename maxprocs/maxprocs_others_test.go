// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux

package maxprocs_test

import (
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/maxprocs"
)

func TestConfigure_NonLinux(t *testing.T) {
	reset()
	maxprocs.Configure(maxprocs.WithLogger(logger()))
	v := maxprocs.Current()
	cpu := runtime.NumCPU()
	if v != cpu {
		t.Errorf("GOMAXPROCS expected=%d, got=%d", cpu, v)
	}
}
