// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux

package cgroup_test

import (
	"errors"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/cgroup"
)

func TestGetInfo(t *testing.T) {
	v, err := cgroup.GetInfo("", "")
	if v != nil {
		t.Errorf("expected nil on non linux platform")
	}

	if !errors.Is(err, cgroup.ErrUnsupportedPlatform) {
		t.Errorf("expected error=%s got=%s", cgroup.ErrUnsupportedPlatform, err)
	}
}
