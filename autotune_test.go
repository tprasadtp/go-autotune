// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func verify(t *testing.T, procs int, mem int64) {
	t.Helper()
	v := maxprocs.Current()
	if v != procs {
		t.Errorf("GOMAXPROCS expected=%d, got=%d", procs, v)
	}

	mv := memlimit.Current()
	if mem != mv {
		t.Errorf("GOMEMLMIT expected=%d, got=%d", mem, mv)
	}
}
