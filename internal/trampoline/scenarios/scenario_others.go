// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !windows && !linux

package scenarios

import "github.com/tprasadtp/go-autotune/internal/trampoline"

// PlatformSpecific scenarios.
func PlatformSpecific() []trampoline.Scenario {
	return []trampoline.Scenario{}
}
