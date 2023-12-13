// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// Constant for bytes.
const (
	KByte = 1000
	MByte = KByte * 1000
	GByte = MByte * 1000
	TByte = GByte * 1000
)

// Constants for IEC size.
const (
	_ = 1 << (iota * 10)
	KiByte
	MiByte
	GiByte
	TiByte
)

// ParseSize parses given human readable string to bytes.
func ParseSize(s string) (int64, error) {
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
	f, err := strconv.ParseFloat(s[:i], 64)
	if err != nil || f < 0 {
		return 0, fmt.Errorf("invalid size: %w", err)
	}

	// Remove spaces and case insensitive, so "100 mb" is same as "100MB"
	unit := strings.ToLower(strings.TrimSpace(s[i:]))
	multiplier := float64(1)

	switch unit {
	case "", "b":
		// already in bytes
	case "k", "kb", "kilobyte", "kilobytes":
		multiplier = KByte
	case "m", "mb", "megabyte", "megabytes":
		multiplier = MByte
	case "g", "gb", "gigabyte", "gigabytes":
		multiplier = GByte
	case "t", "tb", "terabyte", "terabytes":
		multiplier = TByte
	case "kib", "ki":
		multiplier = KiByte
	case "mib", "mi":
		multiplier = MiByte
	case "gib", "gi":
		multiplier = GiByte
	case "tib", "ti":
		multiplier = TiByte
	default:
		return 0, fmt.Errorf("invalid size unit: %q", unit)
	}

	return int64(math.Ceil(f * multiplier)), nil
}
