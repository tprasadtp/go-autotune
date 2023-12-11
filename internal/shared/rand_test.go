// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared_test

import (
	"testing"

	"github.com/tprasadtp/go-autotune/internal/shared"
)

func TestRandomString(t *testing.T) {
	s := shared.RandomString()
	if s == "" {
		t.Errorf("Expected non empty string")
	}
}
