// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package platform_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/internal/platform"
)

// VerifyQuotaFunc returns a function which can be used by [Trampoline].
func VerifyQuotaFunc(cpus float64, max, high int64) func(tb testing.TB) {
	return func(tb testing.TB) {
		tb.Helper()
		qcpu, err := platform.GetCPUQuota()
		if err != nil {
			tb.Errorf("expected no error, got=%s", err)
		}

		if qcpu != float64(cpus) {
			tb.Errorf("expected=%f, got=%f", cpus, qcpu)
		}

		qmax, qhigh, err := platform.GetMemoryQuota()
		if err != nil {
			tb.Errorf("expected no error, got=%s", err)
		}

		if qmax != max {
			tb.Errorf("expected max=%d, got=%d", max, qmax)
		}

		if qhigh != high {
			tb.Errorf("expected high=%d, got=%d", high, qhigh)
		}
	}
}
