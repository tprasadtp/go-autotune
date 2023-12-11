// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"testing"
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
			name:   "100kb",
			input:  "100kb",
			expect: 100e3,
		},
		{
			name:   "100KB",
			input:  "100KB",
			expect: 100e3,
		},
		{
			name:   "99.99KB",
			input:  "99.99KB",
			expect: 99990,
		},
		{
			name:   "9.99MB",
			input:  "9.99MB",
			expect: 9990000,
		},
		{
			name:   "9.99GB",
			input:  "9.99GB",
			expect: 9.99e+9,
		},
		{
			name:   "9.99TB",
			input:  "9.99TB",
			expect: 9.99e+12,
		},
		{
			name:   "100KiB",
			input:  "100KiB",
			expect: 100 * kiByte,
		},
		{
			name:   "100Ki",
			input:  "100Ki",
			expect: 100 * kiByte,
		},
		{
			name:   "1.0MiB",
			input:  "1.0MiB",
			expect: 1048576,
		},
		{
			name:   "1.0Mi",
			input:  "1.0Mi",
			expect: 1048576,
		},
		{
			name:   "9.9MiB",
			input:  "9.9MiB",
			expect: 10380903,
		},
		{
			name:   "9.9Mi",
			input:  "9.9Mi",
			expect: 10380903,
		},
		{
			name:   "1.0GiB",
			input:  "1.0GiB",
			expect: 1073741824,
		},
		{
			name:   "1.0Gi",
			input:  "1.0Gi",
			expect: 1073741824,
		},
		{
			name:   "9.9GiB",
			input:  "9.9GiB",
			expect: 10630044058, // 1073741824 * 9.9
		},
		{
			name:   "9.9Gi",
			input:  "9.9Gi",
			expect: 10630044058, // 1073741824 * 9.9
		},
		{
			name:   "9.9TiB",
			input:  "9.9TiB",
			expect: 10885165114983, // 1099511627776 * 9.9
		},
		{
			name:   "9.9Ti",
			input:  "9.9Ti",
			expect: 10885165114983, // 1099511627776 * 9.9
		},
		{
			name:   "0kb",
			input:  "0kb",
			expect: 0,
		},
		{
			name:   "0Ki",
			input:  "0Ki",
			expect: 0,
		},
		{
			name:   "0Mi",
			input:  "0Mi",
			expect: 0,
		},
		{
			name:   "0Gi",
			input:  "0Gi",
			expect: 0,
		},
		{
			name:   "0Ti",
			input:  "0Ti",
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
			name:   "0mb",
			input:  "0mb",
			expect: 0,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s, err := ParseSize(tc.input)
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
