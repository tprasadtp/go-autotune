package autotune_test

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/tprasadtp/go-autotune/maxprocs"
)

// Custom init function which configures GOMAXPROCS.
// This is different from blank importing github.com/tprasadtp/go-autotune
// as user can customize options. In this only sets GOMAXPROCS.
//
// Do not use this together with blank import of github.com/tprasadtp/go-autotune.
//
//nolint:gochecknoinits // ignore
func init() {
	maxprocs.Configure(maxprocs.WithLogger(slog.Default()))
}

func Example_onlyGOMAXPROCS() {
	fmt.Printf("GOOS      : %s\n", runtime.GOOS)
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(-1))
	fmt.Printf("CPUs      : %d\n", runtime.NumCPU())
}
