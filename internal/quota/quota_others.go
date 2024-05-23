// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux && !windows

package quota

import (
	"context"
	"errors"
)

type Detector struct{}

func (d *Detector) DetectCPUQuota(_ context.Context) (float64, error) {
	return 0, errors.ErrUnsupported
}

//nolint:nonamedreturns // for docs.
func (d *Detector) DetectMemoryQuota(_ context.Context) (max, high int64, err error) {
	return 0, 0, errors.ErrUnsupported
}
