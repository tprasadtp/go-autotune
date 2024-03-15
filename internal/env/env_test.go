// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package env_test

import (
	"fmt"
	"testing"

	"github.com/tprasadtp/go-autotune/internal/env"
)

func TestIsTrue(t *testing.T) {
	tt := []struct {
		env    string
		expect bool
	}{
		{"true", true},
		{"yes", true},
		{"1", true},
		{"TRUE", true},
		{"enable", true},
		{"enabled", true},
		{"on", true},
		{"", false},
		{"off", false},
		{"0", false},
		{"FALSE", false},
		{"hey-this-sort-of-nonsense-can-only-be-written-by-a-software-developer", false},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("env=%s", tc.env), func(t *testing.T) {
			t.Setenv("GO_TEST_PKG_SHARED_ENV_FOO", tc.env)
			v := env.IsTrue("GO_TEST_PKG_SHARED_ENV_FOO")
			if tc.expect != v {
				t.Errorf("expected=%t, got=%t", tc.expect, v)
			}
		})
	}
}

func TestIsFalse(t *testing.T) {
	tt := []struct {
		env    string
		expect bool
	}{
		{"true", false},
		{"yes", false},
		{"false", true},
		{"no", true},
		{"disable", true},
		{"off", true},
		{"0", true},
		{"true", false},
		{"1", false},
		{"hey-this-sort-of-nonsense-can-only-be-written-by-a-software-developer", false},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("env=%s", tc.env), func(t *testing.T) {
			t.Setenv("GO_TEST_PKG_SHARED_ENV_FOO", tc.env)
			v := env.IsFalse("GO_TEST_PKG_SHARED_ENV_FOO")
			if tc.expect != v {
				t.Errorf("expected=%t, got=%t", tc.expect, v)
			}
		})
	}
}

func TestIsDebug(t *testing.T) {
	tt := []struct {
		env    string
		expect bool
	}{
		{"true", false},
		{"yes", false},
		{"false", false},
		{"no", false},
		{"disable", false},
		{"off", false},
		{"0", false},
		{"true", false},
		{"1", false},
		{"debug", true},
		{"DEBUG", true},
		{" DEBUG ", true},
		{"hey-this-sort-of-nonsense-can-only-be-written-by-a-software-developer", false},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("env=%s", tc.env), func(t *testing.T) {
			t.Setenv("GO_TEST_PKG_SHARED_ENV_FOO", tc.env)
			v := env.IsDebug("GO_TEST_PKG_SHARED_ENV_FOO")
			if tc.expect != v {
				t.Errorf("expected=%t, got=%t", tc.expect, v)
			}
		})
	}
}
