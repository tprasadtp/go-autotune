// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package quota

import (
	"path/filepath"
	"testing"
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
			v, err := GetCgroupInterfacePath(filepath.Join("testdata", "procfs", tc.procfs))

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
				if tc.expect != filepath.ToSlash(v) {
					t.Errorf("expected=%s, got=%s", tc.expect, v)
				}
			}
		})
	}
}
