// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit_test

import (
	"log/slog"
	"math"
	"os"
	"testing"

	"github.com/tprasadtp/go-autotune/memlimit"
)

func TestConfigure_UnsupportedPlatform(t *testing.T) {
	os.Unsetenv("GOMEMLIMIT")
	memlimit.Configure(memlimit.WithLogger(slog.Default()))
	v := memlimit.Current()
	if v != math.MaxInt64 {
		t.Errorf("GOMEMLIMIT expected=%d, got=%d", math.MaxInt64, v)
	}
}
