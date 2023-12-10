// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cgroup

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
			name:   "Hybrid-Cgroup",
			procfs: "cgroup-hybrid",
			expect: "/sys/fs/cgroup/unified/user.slice/user-1000.slice/user@1000.service/app.slice/run-u18351.service",
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
