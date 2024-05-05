// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Constants for IEC size.
const (
	_ = 1 << (iota * 10)
	KiByte
	MiByte
	GiByte
	TiByte
)

// ParseMemlimit parses given human readable string to bytes.
// This only accepts valid values for GOMEMLIMIT.
func ParseMemlimit(s string) (int64, error) {
	// As special case if file size empty return zero value.
	if s == "" {
		return 0, nil
	}

	// Save index of lastDigit to parse unit.
	var i int
	for _, r := range s {
		if !(unicode.IsDigit(r) || r == '.') {
			break
		}
		i++
	}

	// Try to parse s[0:i] as floating point value.
	f, err := strconv.ParseInt(s[:i], 10, 64)
	if err != nil || f < 0 {
		return 0, fmt.Errorf("invalid size: %w", err)
	}

	// Remove spaces and case insensitive, so "100 mb" is same as "100MB"
	unit := strings.ToLower(strings.TrimSpace(s[i:]))
	multiplier := int64(1)

	switch unit {
	case "", "b":
		// already in bytes
	case "kib":
		multiplier = KiByte
	case "mib":
		multiplier = MiByte
	case "gib":
		multiplier = GiByte
	case "tib":
		multiplier = TiByte
	default:
		return 0, fmt.Errorf("invalid size unit: %q", unit)
	}

	return f * multiplier, nil
}
