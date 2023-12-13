// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux

package platform_test

import (
	"errors"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/platform"
)

func TestGetCPUQuota(t *testing.T) {
	v, err := platform.GetCPUQuota()
	if v != 0 {
		t.Errorf("expected 0 unsupported platform(%s)", runtime.GOOS)
	}

	if !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("expected error=%s got=%s", errors.ErrUnsupported, err)
	}
}

func TestGetMemoryQuota(t *testing.T) {
	max, high, err := platform.GetMemoryQuota()
	if max != 0 {
		t.Errorf("expected 0 unsupported platform(%s)", runtime.GOOS)
	}

	if high != 0 {
		t.Errorf("expected 0 unsupported platform(%s)", runtime.GOOS)
	}

	if !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("expected error=%s got=%s", errors.ErrUnsupported, err)
	}
}
