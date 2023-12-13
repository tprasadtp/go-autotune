// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux && !windows

package platform

import (
	"errors"
	"fmt"
	"runtime"
)

func getCPUQuota(options ...Option) (float64, error) {
	return 0, fmt.Errorf("platform: unsupported platform(%s): %w", runtime.GOOS, errors.ErrUnsupported)
}

//nolint:nonamedreturns // for docs.
func getMemoryQuota(options ...Option) (max, high int64, err error) {
	return 0, 0, fmt.Errorf("platform: unsupported platform(%s): %w", runtime.GOOS, errors.ErrUnsupported)
}
