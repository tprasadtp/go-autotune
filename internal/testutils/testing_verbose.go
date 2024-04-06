// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package testutils

import (
	"flag"
	"sync"
)

var (
	testVerboseCache bool
	testVerboseOnce  sync.Once
)

// TestingIsVerbose returns true if test.v flag is set.
func TestingIsVerbose() bool {
	testVerboseOnce.Do(func() {
		v := flag.Lookup("test.v")
		if v != nil {
			if v.Value.String() == "true" {
				testVerboseCache = true
			}
		}
	})
	return testVerboseCache
}
