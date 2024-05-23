// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package quota_test

import (
	"context"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/quota"
)

// VerifyQuotaFunc returns a function which can be used by [Trampoline].
func VerifyQuotaFunc(cpus float64, hard, soft int64) func(tb testing.TB) {
	return func(tb testing.TB) {
		tb.Helper()
		ctx := context.Background()
		d := &quota.Detector{}

		qcpu, err := d.DetectCPUQuota(ctx)
		if err != nil {
			tb.Errorf("expected no error, got=%s", err)
		}

		if qcpu != float64(cpus) {
			tb.Errorf("expected=%f, got=%f", cpus, qcpu)
		}

		qmax, qhigh, err := d.DetectMemoryQuota(ctx)
		if err != nil {
			tb.Errorf("expected no error, got=%s", err)
		}

		if qmax != hard {
			tb.Errorf("expected max=%d, got=%d", hard, qmax)
		}

		if qhigh != soft {
			tb.Errorf("expected high=%d, got=%d", soft, qhigh)
		}
	}
}
