// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

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
