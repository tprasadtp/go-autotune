// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package maxprocs

var (
	_ error = Error("")
)

// Error is immutable error representation.
//
// Error strings themselves are NOT part of semver compatibility guarantees.
// Use exported symbols instead of directly using error strings.
type Error string

// Implements Error() interface.
func (e Error) Error() string {
	return string(e)
}

const (
	errUnsupportedPlatform = Error("maxprocs: Only supported on Linux")
)
