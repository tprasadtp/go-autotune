package autotune_test

import (
	"fmt"
	"runtime"
	"runtime/debug"

	_ "github.com/tprasadtp/go-autotune" // Importing this adjusts GOMAXPROCS & GOMEMLIMIT
)

// To render a whole-file example, a package-level declaration is required.
var _ = ""

func Example() {
	fmt.Printf("GOOS       : %s\n", runtime.GOOS)
	fmt.Printf("GOMAXPROCS : %d\n", runtime.GOMAXPROCS(-1))
	fmt.Printf("GOMEMLIMIT : %d\n", debug.SetMemoryLimit(-1))
	fmt.Printf("CPUs       : %d\n", runtime.NumCPU())
}
