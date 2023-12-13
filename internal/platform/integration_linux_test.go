// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package platform_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/internal/platform"
	"github.com/tprasadtp/go-autotune/internal/shared"
)

func TestGetQuota(t *testing.T) {
	t.Run("NoQuota", func(t *testing.T) {
		args := []string{"--pipe"}
		shared.SystemdRun(t, args, func(t *testing.T) {
			path, err := platform.GetCgroupInterfacePath()
			if err != nil {
				t.Fatalf("failed to get cgroup path")
			}

			cpu, err := platform.GetCPUQuota(platform.WithCgroupInterfacePath(path))
			if err != nil {
				t.Errorf("expected no error, got=%s", err)
			}

			if cpu != 0 {
				t.Errorf("expected 0 error, got=%f", cpu)
			}

			max, high, err := platform.GetMemoryQuota(platform.WithCgroupInterfacePath(path))
			if err != nil {
				t.Errorf("expected no error, got=%s", err)
			}

			if max != 0 {
				t.Errorf("expected max=0, got=%d", max)
			}

			if high != 0 {
				t.Errorf("expected high=0, got=%d", high)
			}
		})
	})
	t.Run("WithQuota", func(t *testing.T) {
		args := []string{
			"--pipe",
			"--property=CPUQuota=50%",
			"--property=MemoryHigh=2.5G",
			"--property=MemoryMax=3G",
		}
		shared.SystemdRun(t, args, func(t *testing.T) {
			path, err := platform.GetCgroupInterfacePath()
			if err != nil {
				t.Fatalf("failed to get cgroup path")
			}

			cpu, err := platform.GetCPUQuota(platform.WithCgroupInterfacePath(path))
			if err != nil {
				t.Errorf("expected no error, got=%s", err)
			}

			if cpu != 0.5 {
				t.Errorf("expected 0.5 error, got=%f", cpu)
			}

			max, high, err := platform.GetMemoryQuota(platform.WithCgroupInterfacePath(path))
			if err != nil {
				t.Errorf("expected no error, got=%s", err)
			}

			if max != 3*shared.GiByte {
				t.Errorf("expected max=%d, got=%d", int64(3*shared.GiByte), max)
			}

			if high != 2.5*shared.GiByte {
				t.Errorf("expected high=%d, got=%d", int64(2.5*shared.GiByte), high)
			}
		})
	})
}
