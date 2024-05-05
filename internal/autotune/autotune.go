// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package autotune which implements autotune functions.
//
// This is an internal package with only single method exported
// to allow easily benchmark and test without side effects of init.
package autotune

// Configure configures GOMAXPROCS and GOMEMLIMIT. This is only intended
// to be used for testing and use in init function of the public package.
func Configure() {
	configure()
}
