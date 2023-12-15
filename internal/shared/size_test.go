// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
)

func TestParseFileSize(t *testing.T) {
	type testCase struct {
		name    string
		input   string
		expect  int64
		invalid bool
	}
	tt := []testCase{
		{
			name:   "empty-string",
			input:  "",
			expect: 0,
		},
		{
			name:    "spaces",
			input:   "     ",
			invalid: true,
		},
		{
			name:   "zero",
			input:  "0",
			expect: 0,
		},
		{
			name:   "zero-bytes",
			input:  "0B",
			expect: 0,
		},
		{
			name:    "invalid-string",
			input:   "foo-bar",
			invalid: true,
		},
		{
			name:    "hexadecimal",
			input:   "0x1ffffff",
			invalid: true,
		},
		{
			name:    "negative",
			input:   "-2.5MB",
			invalid: true,
		},
		{
			name:   "bytes(100mb)",
			input:  "100000000",
			expect: 1e8,
		},
		{
			name:   "100kib",
			input:  "100kib",
			expect: shared.KiByte * 100,
		},
		{
			name:    "100KB",
			input:   "100KB",
			invalid: true,
		},
		{
			name:    "99.99KB",
			input:   "99.99KB",
			invalid: true,
		},
		{
			name:    "9.99MB",
			input:   "9.99MB",
			invalid: true,
		},
		{
			name:    "9.99GB",
			input:   "9.99GB",
			invalid: true,
		},
		{
			name:    "9.99TB",
			input:   "9.99TB",
			invalid: true,
		},
		{
			name:   "100KiB",
			input:  "100KiB",
			expect: 100 * shared.KiByte,
		},
		{
			name:    "100Ki",
			input:   "100Ki",
			invalid: true,
		},
		{
			name:   "1MiB",
			input:  "1MiB",
			expect: shared.MiByte,
		},
		{
			name:    "1.0Mi",
			input:   "1.0Mi",
			invalid: true,
		},
		{
			name:    "9.9MiB",
			input:   "9.9MiB",
			invalid: true,
		},
		{
			name:    "9.9Mi",
			input:   "9.9Mi",
			invalid: true,
		},
		{
			name:   "1GiB",
			input:  "1GiB",
			expect: 1073741824,
		},
		{
			name:    "1.0Gi",
			input:   "1.0Gi",
			invalid: true,
		},
		{
			name:    "9.9GiB",
			input:   "9.9GiB",
			invalid: true,
		},
		{
			name:    "9.9Gi",
			input:   "9.9Gi",
			invalid: true,
		},
		{
			name:    "9.9TiB",
			input:   "9.9TiB",
			invalid: true,
		},
		{
			name:    "9.9Ti",
			input:   "9.9Ti",
			invalid: true,
		},
		{
			name:  "0KiB",
			input: "0KiB",
		},
		{
			name:  "0MiB",
			input: "0MiB",
		},
		{
			name:  "0GiB",
			input: "0GiB",
		},
		{
			name:  "0TiB",
			input: "0TiB",
		},
		{
			name:  "0",
			input: "0",
		},
		{
			name:  "0B",
			input: "0B",
		},
		{
			name:  "0b",
			input: "0b",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s, err := shared.ParseSize(tc.input)
			if tc.invalid {
				if s != 0 {
					t.Errorf("expect value to be 0 when input is invalid (%q)", tc.input)
				}
				if err == nil {
					t.Errorf("expected error when input is invalid (%q)", tc.input)
				}
			} else {
				if s != tc.expect {
					t.Errorf("expect value=%d but got=%d", tc.expect, s)
				}
				if err != nil {
					t.Errorf("expected no error but got (%s)", err)
				}
			}
		})
	}
}
