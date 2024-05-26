// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package shared provides utilities shared across multiple internal packages.
package shared

import (
	"fmt"
	"math"
	"strconv"
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
//
// s must match the following regular expression:
//
//	^[0-9]+(([KMGT]i)?B)?$
//
// Return value is int64 because of Runtime API. returned value is always positive.
func ParseMemlimit(s string) (int64, error) {
	// Save index of lastDigit to parse unit.
	var i int
	for _, r := range s {
		if !unicode.IsDigit(r) {
			break
		}
		i++
	}

	// Try to parse s[0:i] as an integer value.
	v, err := strconv.ParseUint(s[:i], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid input(%q): %w", s, err)
	}

	// Because value is parsed as uint64, check if it overflows int64.
	if v > math.MaxInt64 {
		return 0, fmt.Errorf("integer overflow: %q", s)
	}

	// Parse units.
	unit := s[i:]
	multiplier := uint64(1)

	switch unit {
	case "", "B":
		// already in bytes
	case "KiB":
		multiplier = KiByte
	case "MiB":
		multiplier = MiByte
	case "GiB":
		multiplier = GiByte
	case "TiB":
		multiplier = TiByte
	default:
		return 0, fmt.Errorf("invalid size unit: %q", unit)
	}

	rv := v * multiplier

	if rv > 0 {
		if rv > math.MaxInt64 || rv/multiplier != v {
			return 0, fmt.Errorf("integer overflow: %q", s)
		}
	}

	return int64(rv), nil
}
