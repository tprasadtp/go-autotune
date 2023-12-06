// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package autotune

import (
	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

//nolint:gochecknoinits // ignore
func init() {
	if !shared.IsFalse("GO_AUTOTUNE") && !shared.IsFalse("GOAUTOTUNE") {
		maxprocs.Configure()
		memlimit.Configure()
	}
}
