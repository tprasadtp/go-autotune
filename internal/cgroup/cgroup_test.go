// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cgroup

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
)

func TestGetInterfacePath(t *testing.T) {
	tt := []struct {
		name   string
		procfs string
		expect string
		err    bool
	}{
		{
			name:   "cgroup-hybrid",
			procfs: "cgroup-hybrid",
			expect: "/sys/fs/cgroup/unified/user.slice/user-1000.slice/user@1000.service/app.slice/run-u18351.service",
		},
		{
			name:   "cgroup-invalid",
			procfs: "cgroup-invalid",
			err:    true,
		},
		{
			name:   "cgroup-mount-missing-from-mountinfo",
			procfs: "cgroup-mount-missing",
			err:    true,
		},
		{
			name:   "cgroup-v1",
			procfs: "cgroup-v1",
			err:    true,
		},
		{
			name:   "docker-debian",
			procfs: "docker-debian",
			expect: "/sys/fs/cgroup",
		},
		{
			name:   "invalid-cgroup",
			procfs: "invalid-cgroup",
			err:    true,
		},
		{
			name:   "missing-cgroup-file",
			procfs: "missing-cgroup",
			err:    true,
		},
		{
			name:   "missing-mountinfo-file",
			procfs: "missing-mountinfo",
			err:    true,
		},
		{
			name:   "mountinfo-invalid",
			procfs: "mountinfo-invalid",
			err:    true,
		},
		{
			name:   "podman-fedora",
			procfs: "podman-fedora",
			expect: "/sys/fs/cgroup",
		},
		{
			name:   "systemd-debian",
			procfs: "systemd-debian",
			expect: "/sys/fs/cgroup/user.slice/user-0.slice/user@0.service/app.slice/run-u4.service",
		},
		{
			name:   "systemd-nspawn",
			procfs: "systemd-nspawn",
			expect: "/sys/fs/cgroup/user.slice/user-0.slice/session-8.scope",
		},
		{
			name:   "systemd-system",
			procfs: "systemd-system",
			expect: "/sys/fs/cgroup/system.slice/run-u1801.service",
		},
		{
			name:   "systemd-user",
			procfs: "systemd-user",
			expect: "/sys/fs/cgroup/user.slice/user-1000.slice/user@1000.service/app.slice/run-u18351.service",
		},
		{
			name:   "systemd-user-fedora",
			procfs: "systemd-user-fedora",
			expect: "/sys/fs/cgroup/user.slice/user-1000.slice/user@1000.service/app.slice/run-u119.service",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v, err := GetInterfacePath(
				WithProcFSPath(filepath.Join("testdata", "procfs", tc.procfs)),
			)

			if tc.err {
				if err == nil {
					t.Errorf("expected an error, got nil")
				}

				if v != "" {
					t.Errorf("must return empty string when error is expected")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %s", err)
				}
				if tc.expect != v {
					t.Errorf("expected=%s, got=%s", tc.expect, v)
				}
			}
		})
	}
}

func TestGetQuota(t *testing.T) {
	tt := []struct {
		name   string
		path   string
		expect *Quota
		err    bool
	}{
		{
			name:   "no-limits",
			path:   "no-limits",
			expect: &Quota{},
		},
		{
			name: "cpu-50",
			path: "cpu-50",
			expect: &Quota{
				CPU: 0.5,
			},
		},
		{
			name: "cpu-250",
			path: "cpu-250",
			expect: &Quota{
				CPU: 2.5,
			},
		},
		{
			name: "cpu-250-10ms",
			path: "cpu-250-10ms",
			expect: &Quota{
				CPU: 2.5,
			},
		},
		{
			name: "cpu-300",
			path: "cpu-300",
			expect: &Quota{
				CPU: 3,
			},
		},
		{
			name: "mem-high-250",
			path: "mem-high-250",
			expect: &Quota{
				MemoryHigh: shared.MustParseSize("250MiB"),
			},
		},
		{
			name: "mem-max-250",
			path: "mem-max-250",
			expect: &Quota{
				MemoryMax: shared.MustParseSize("250MiB"),
			},
		},
		{
			name: "mem-max-250-high-200",
			path: "mem-max-250-high-200",
			expect: &Quota{
				MemoryMax:  shared.MustParseSize("250MiB"),
				MemoryHigh: shared.MustParseSize("200MiB"),
			},
		},
		{
			name: "mem-max-250-high-250",
			path: "mem-max-250-high-250",
			expect: &Quota{
				MemoryMax:  shared.MustParseSize("250MiB"),
				MemoryHigh: shared.MustParseSize("250MiB"),
			},
		},
		{
			name: "mem-max-300-high-500",
			path: "mem-max-300-high-500",
			expect: &Quota{
				MemoryMax:  shared.MustParseSize("300MiB"),
				MemoryHigh: shared.MustParseSize("500MiB"),
			},
		},
		{
			name: "cpu-250-10ms-mem-max-300-high-250",
			path: "cpu-250-10ms-mem-max-300-high-250",
			expect: &Quota{
				CPU:        2.5,
				MemoryMax:  shared.MustParseSize("300MiB"),
				MemoryHigh: shared.MustParseSize("250MiB"),
			},
		},
		{
			name: "cpu-invalid",
			path: "cpu-invalid",
			err:  true,
		},
		{
			name: "cpu-negative",
			path: "cpu-negative",
			err:  true,
		},
		{
			name: "cpu-negative-interval",
			path: "cpu-negative-interval",
			err:  true,
		},
		{
			name: "mem-high-invalid",
			path: "mem-high-invalid",
			err:  true,
		},
		{
			name: "mem-high-negative",
			path: "mem-high-negative",
			err:  true,
		},
		{
			name: "mem-max-invalid",
			path: "mem-max-invalid",
			err:  true,
		},
		{
			name: "mem-max-negative",
			path: "mem-max-negative",
			err:  true,
		},
		{
			name:   "no-limits-no-files",
			expect: &Quota{},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v, err := GetQuota(
				WithInterfacePath(filepath.Join("testdata", "cgroup", tc.path)),
			)

			if tc.err {
				if err == nil {
					t.Errorf("expected an error, got nil")
				}

				if v != nil {
					t.Errorf("must return nil when error is expected")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %s", err)
				}
				if !reflect.DeepEqual(v, tc.expect) {
					t.Errorf("expected=%+v, got=%+v", tc.expect, v)
				}
			}
		})
	}
}
