// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/internal/autotune"
	"github.com/tprasadtp/go-autotune/internal/trampoline"
	"github.com/tprasadtp/go-autotune/internal/trampoline/scenarios"
)

func TestIntegration(t *testing.T) {
	tt := scenarios.All()
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			trampoline.Trampoline(t, tc.Opts, tc.Verify, autotune.Configure)
		})
	}
}
