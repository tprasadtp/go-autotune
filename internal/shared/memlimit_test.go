// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"testing"
)

func TestParseMemlimit(t *testing.T) {
	tt := []struct {
		input  string
		expect int64
		valid  bool
	}{
		// Good numeric inputs.
		{"1", 1, true},
		{"12345", 12345, true},
		{"012345", 12345, true},
		{"98765432100", 98765432100, true},
		{"9223372036854775807", 1<<63 - 1, true},

		// Good trivial suffix inputs.
		{"1B", 1, true},
		{"12345B", 12345, true},
		{"012345B", 12345, true},
		{"98765432100B", 98765432100, true},
		{"9223372036854775807B", 1<<63 - 1, true},

		// Good binary suffix inputs.
		{"1KiB", 1 << 10, true},
		{"05KiB", 5 << 10, true},
		{"1MiB", 1 << 20, true},
		{"10MiB", 10 << 20, true},
		{"1GiB", 1 << 30, true},
		{"100GiB", 100 << 30, true},
		{"1TiB", 1 << 40, true},
		{"99TiB", 99 << 40, true},

		// Good zero inputs.
		{"0", 0, true},
		{"0B", 0, true},
		{"0KiB", 0, true},
		{"0MiB", 0, true},
		{"0GiB", 0, true},
		{"0TiB", 0, true},

		// Bad inputs.
		{"-0", 0, false},
		{"", 0, false},
		{"-1", 0, false},
		{"a12345", 0, false},
		{"a12345B", 0, false},
		{"12345x", 0, false},
		{"0x12345", 0, false},

		// Bad numeric inputs.
		{"9223372036854775808", 0, false},
		{"9223372036854775809", 0, false},
		{"18446744073709551615", 0, false},
		{"20496382327982653440", 0, false},
		{"18446744073709551616", 0, false},
		{"18446744073709551617", 0, false},
		{"9999999999999999999999", 0, false},

		// Bad trivial suffix inputs.
		{"9223372036854775808B", 0, false},
		{"9223372036854775809B", 0, false},
		{"18446744073709551615B", 0, false},
		{"20496382327982653440B", 0, false},
		{"18446744073709551616B", 0, false},
		{"18446744073709551617B", 0, false},
		{"9999999999999999999999B", 0, false},

		// Bad binary suffix inputs.
		{"1Ki", 0, false},
		{"05Ki", 0, false},
		{"10Mi", 0, false},
		{"100Gi", 0, false},
		{"99Ti", 0, false},
		{"22iB", 0, false},
		{"B", 0, false},
		{"iB", 0, false},
		{"KiB", 0, false},
		{"MiB", 0, false},
		{"GiB", 0, false},
		{"TiB", 0, false},
		{"-120KiB", 0, false},
		{"-891MiB", 0, false},
		{"-704GiB", 0, false},
		{"-42TiB", 0, false},
		{"99999999999999999999KiB", 0, false},
		{"99999999999999999MiB", 0, false},
		{"99999999999999GiB", 0, false},
		{"99999999999TiB", 0, false},
		{"555EiB", 0, false},

		// Mistaken SI suffix inputs.
		{"0KB", 0, false},
		{"0MB", 0, false},
		{"0GB", 0, false},
		{"0TB", 0, false},
		{"1KB", 0, false},
		{"05KB", 0, false},
		{"1MB", 0, false},
		{"10MB", 0, false},
		{"1GB", 0, false},
		{"100GB", 0, false},
		{"1TB", 0, false},
		{"99TB", 0, false},
		{"1K", 0, false},
		{"05K", 0, false},
		{"10M", 0, false},
		{"100G", 0, false},
		{"99T", 0, false},
		{"99999999999999999999KB", 0, false},
		{"99999999999999999MB", 0, false},
		{"99999999999999GB", 0, false},
		{"99999999999TB", 0, false},
		{"99999999999TiB", 0, false},
		{"555EB", 0, false},
	}
	for _, tc := range tt {
		t.Run(tc.input, func(t *testing.T) {
			s, err := ParseMemlimit(tc.input)
			if tc.valid {
				if s != tc.expect {
					t.Errorf("expect value=%d but got=%d", tc.expect, s)
				}
				if err != nil {
					t.Errorf("expected no error but got (%s)", err)
				}
			} else {
				if s != 0 {
					t.Errorf("expect value to be 0 when input is invalid (%q)", tc.input)
				}
				if err == nil {
					t.Errorf("expected error when input is invalid (%q)", tc.input)
				}
			}
		})
	}
}
