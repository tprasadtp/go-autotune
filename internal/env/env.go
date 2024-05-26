// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package env provides utilities for processing environment variables.
package env

import (
	"os"
	"strings"
)

// IsTrue checks if environment variable env is set to truthy value.
func IsTrue(env string) bool {
	value := os.Getenv(env)

	switch strings.ToLower(value) {
	case "true", "1", "yes", "enable", "enabled", "on":
		return true
	default:
		return false
	}
}

// IsFalse checks if environment variable env is set to false value.
func IsFalse(env string) bool {
	value := os.Getenv(env)

	switch strings.ToLower(value) {
	case "false", "0", "no", "disable", "disabled", "off":
		return true
	default:
		return false
	}
}

// IsDebug checks if environment variable env is set to "debug".
func IsDebug(env string) bool {
	switch strings.TrimSpace(strings.ToLower(os.Getenv(env))) {
	case "debug":
		return true
	default:
		return false
	}
}
