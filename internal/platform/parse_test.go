// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package platform

import "testing"

func TestUnescape(t *testing.T) {
	tt := []struct {
		name     string
		path     string
		expected string
		err      bool
	}{
		{"empty-path", "", "", false},
		{"root", "/", "/", false},
		{"non-root", "/some/longer/path", "/some/longer/path", false},
		{"with-spaces", "/path\\040with\\040spaces", "/path with spaces", false},
		{"with-backslash", "/path/with\\134backslash", "/path/with\\backslash", false},
		{"with-tab", "/tab\\011in/path", "/tab\tin/path", false},
		{"with-quotes", `/path/"with'quotes`, `/path/"with'quotes`, false},
		{"with-quotes-tab-space", `/path/"with'quotes,\040space,\011tab`, `/path/"with'quotes, space,	tab`, false},
		{"backslash", `\134`, `\`, false},
		{"invalid-quotes-1", `"'"'"'`, `"'"'"'`, false},
		{"invalid-1", `\12`, "", true},
		{"invalid-2", `/\12x`, "", true},
		{"invlaid-3", `\0`, "", true},
		{"invalid-4", `\x`, "", true},
		{"invalid-5", "\\\\", "", true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v, err := unescape(tc.path)
			if tc.err {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if v != "" {
					t.Errorf("expected emoty output on error, got=%s", v)
				}
			} else {
				if err != nil {
					t.Errorf("expected=nil, got=%s", err)
				}
				if v != tc.expected {
					t.Errorf("expected=%q, got=%q", tc.expected, v)
				}
			}
		})
	}
}
