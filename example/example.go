// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

// Example CLI which can show GOMAXPROCS and GOMEMLIMIT values.
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"

	_ "github.com/tprasadtp/go-autotune"
)

func main() {
	fmt.Printf("GOOS       : %s\n", runtime.GOOS)
	fmt.Printf("GOMAXPROCS : %d\n", runtime.GOMAXPROCS(-1))
	fmt.Printf("CPUs       : %d\n", runtime.NumCPU())
	fmt.Printf("GOMEMLIMIT : %d\n", debug.SetMemoryLimit(-1))
}
