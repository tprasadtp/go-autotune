// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package testutils

import (
	"runtime"
	"testing"
)

func SkipIfCPUCountLessThan(t *testing.T, n int) {
	t.Helper()
	if v := runtime.NumCPU(); v < n {
		t.Skipf("runtime.NumCPU() %d < %d", v, n)
	}
}
