// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/internal/trampoline"
	"github.com/tprasadtp/go-autotune/internal/trampoline/scenarios"
)

func TestIntegration(t *testing.T) {
	for _, tc := range scenarios.All() {
		t.Run(tc.Name, func(t *testing.T) {
			// configure is nil, as test package already imports
			// github.com/tprasadtp/go-autotune for side effects.
			trampoline.Trampoline(t, tc.Opts, tc.Verify, nil)
		})
	}
}
