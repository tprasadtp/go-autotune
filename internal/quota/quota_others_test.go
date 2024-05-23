// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux && !windows

package quota_test

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/quota"
)

func TestGetCPUQuota(t *testing.T) {
	d := &quota.Detector{}
	v, err := d.DetectCPUQuota(context.Background())
	if v != 0 {
		t.Errorf("expected 0 unsupported platform(%s)", runtime.GOOS)
	}

	if !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("expected error=%s got=%s", errors.ErrUnsupported, err)
	}
}

func TestGetMemoryQuota(t *testing.T) {
	d := &quota.Detector{}
	max, high, err := d.DetectMemoryQuota(context.Background())
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
