package autotune_test

import (
	"fmt"
	"log/slog"
	"runtime"
	"runtime/debug"

	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

// Custom init function which configures GOMAXPROCS and GOMEMLIMIT.
// This is different from blank importing github.com/tprasadtp/go-autotune
// as user can customize options. In this example default logger is being used.
//
// Do not use this together with blank import of github.com/tprasadtp/go-autotune.
//
//nolint:gochecknoinits // ignore
func init() {
	maxprocs.Configure(maxprocs.WithLogger(slog.Default()))
	memlimit.Configure(memlimit.WithLogger(slog.Default()))
}

func Example_customLogger() {
	fmt.Printf("GOOS      : %s\n", runtime.GOOS)
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(-1))
	fmt.Printf("GOMEMLIMIT: %d\n", debug.SetMemoryLimit(-1))
	fmt.Printf("CPUs      : %d\n", runtime.NumCPU())
}
