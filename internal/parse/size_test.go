// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package parse

import (
	"testing"
)

func TestSize(t *testing.T) {
	type testCase struct {
		name    string
		input   string
		expect  uint64
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
			name:   "bytes(100mb)",
			input:  "100000000",
			expect: 1e8,
		},
		{
			name:   "100KiB",
			input:  "100KiB",
			expect: 100 * kiByte,
		},
		{
			name:   "1.0MiB",
			input:  "1.0MiB",
			expect: 1048576,
		},
		{
			name:   "9.9MiB",
			input:  "9.9MiB",
			expect: 10380903,
		},
		{
			name:   "1.0GiB",
			input:  "1.0GiB",
			expect: 1073741824,
		},
		{
			name:   "9.9GiB",
			input:  "9.9GiB",
			expect: 10630044058, // 1073741824 * 9.9
		},
		{
			name:   "9.9TiB",
			input:  "9.9TiB",
			expect: 10885165114983, // 1099511627776 * 9.9
		},
		{
			name:   "0kib",
			input:  "0kib",
			expect: 0,
		},
		{
			name:   "0",
			input:  "0",
			expect: 0,
		},
		{
			name:   "0B",
			input:  "0B",
			expect: 0,
		},
		{
			name:   "0b",
			input:  "0b",
			expect: 0,
		},
		{
			name:   "0mib",
			input:  "0mib",
			expect: 0,
		},
		{
			name:   "0.0Gib",
			input:  "0.0Gib",
			expect: 0,
		},
		{
			name:   "0.0Tib",
			input:  "0.0Tib",
			expect: 0,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s, err := Size(tc.input)
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
