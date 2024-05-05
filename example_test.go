// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune_test

import (
	"fmt"
	"runtime"
	"runtime/debug"

	_ "github.com/tprasadtp/go-autotune" // Adjusts GOMAXPROCS & GOMEMLIMIT
)

// To render a whole-file example, a package-level declaration is required.
var _ = ""

// This simply prints current GOMAXPROCS and GOMEMLIMIT values.
// If [github.com/tprasadtp/go-autotune] package is imported as specified,
// It should show appropriately configured values for GOMAXPROCS and GOMEMLIMIT.
func Example_simple() {
	fmt.Printf("GOOS       : %s\n", runtime.GOOS)
	fmt.Printf("GOMAXPROCS : %d\n", runtime.GOMAXPROCS(-1))
	fmt.Printf("GOMEMLIMIT : %d\n", debug.SetMemoryLimit(-1))
	fmt.Printf("CPUs       : %d\n", runtime.NumCPU())
}
