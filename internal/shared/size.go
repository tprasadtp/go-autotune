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
	kByte = 1000
	mByte = kByte * 1000
	gByte = mByte * 1000
	tByte = gByte * 1000
)

// Constants for IEC size.
const (
	_ = 1 << (iota * 10)
	kiByte
	miByte
	giByte
	tiByte
)

func MustParseSize(s string) int64 {
	v, err := ParseSize(s)
	if err != nil {
		panic(err)
	}
	return v
}

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
		multiplier = kByte
	case "m", "mb", "megabyte", "megabytes":
		multiplier = mByte
	case "g", "gb", "gigabyte", "gigabytes":
		multiplier = gByte
	case "t", "tb", "terabyte", "terabytes":
		multiplier = tByte
	case "kib":
		multiplier = kiByte
	case "mib":
		multiplier = miByte
	case "gib":
		multiplier = giByte
	case "tib":
		multiplier = tiByte
	default:
		return 0, fmt.Errorf("invalid size unit: %q", unit)
	}

	return int64(math.Ceil(f * multiplier)), nil
}
