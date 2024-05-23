// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package quota

import (
	"fmt"
	"strings"
)

// A few specific characters in mountinfo path entries (root and mountpoint)
// are escaped using a backslash followed by a character's ascii code in octal.
//
//	space              -- as \040
//	tab (aka \t)       -- as \011
//	newline (aka \n)   -- as \012
//	backslash (aka \\) -- as \134
//
// This function un-escapes the above sequences.
func unescape(path string) (string, error) {
	if strings.IndexByte(path, '\\') == -1 {
		return path, nil
	}

	// The following code is UTF-8 transparent as it only looks for some
	// specific characters (backslash and 0..7) with values < utf8.RuneSelf,
	// and everything else is passed through as is.
	buf := make([]byte, len(path))
	bufLen := 0
	for i := 0; i < len(path); i++ {
		if path[i] != '\\' {
			buf[bufLen] = path[i]
			bufLen++
			continue
		}
		s := path[i:]
		if len(s) < 4 {
			// too short
			return "", fmt.Errorf("bad escape sequence %q: too short", s)
		}
		c := s[1]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			v := c - '0'
			for j := 2; j < 4; j++ {
				// one digit already; two more
				if s[j] < '0' || s[j] > '7' {
					return "", fmt.Errorf("bad escape sequence %q: not a digit", s[:3])
				}
				x := s[j] - '0'
				v = (v << 3) | x
			}
			if v > 255 {
				return "", fmt.Errorf("bad escape sequence %q: out of range", s[:3])
			}
			buf[bufLen] = v
			bufLen++
			i += 3
			continue
		default:
			return "", fmt.Errorf("bad escape sequence %q: not a digit", s[:3])
		}
	}

	return string(buf[:bufLen]), nil
}
