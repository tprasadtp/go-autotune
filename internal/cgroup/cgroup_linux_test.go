// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cgroup_test

import (
	"reflect"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/cgroup"
	"github.com/tprasadtp/go-autotune/internal/shared"
)

func TestSystemdRun(t *testing.T) {
	t.Run("NoQuota", func(t *testing.T) {
		args := []string{"--pipe"}
		shared.SystemdRun(t, args, func(t *testing.T) {
			expected := &cgroup.Quota{}
			v, err := cgroup.GetQuota()
			if err != nil {
				t.Errorf("expected no error, got=%s", err)
			}
			if !reflect.DeepEqual(v, expected) {
				t.Errorf("expected=%+v, got=%+v", expected, v)
			}
		})
	})
	t.Run("WithQuota", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--property=CPUQuota=50%",
			"--property=MemoryHigh=250M",
			"--property=MemoryMax=300M",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			expected := &cgroup.Quota{
				CPU:        0.5,
				MemoryMax:  shared.MustParseSize("300Mi"),
				MemoryHigh: shared.MustParseSize("250Mi"),
			}
			v, err := cgroup.GetQuota()
			if err != nil {
				t.Errorf("expected no error, got=%s", err)
			}
			if !reflect.DeepEqual(v, expected) {
				t.Errorf("expected=%+v, got=%+v", expected, v)
			}
		})
	})
}
