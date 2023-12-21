// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package autotune

import (
	"log/slog"

	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

//nolint:gochecknoinits // ignore
func init() {
	if shared.IsFalse("GO_AUTOTUNE") || shared.IsFalse("GOAUTOTUNE") {
		return
	}

	var logger *slog.Logger
	if shared.IsDebug("GO_AUTOTUNE") || shared.IsDebug("GOAUTOTUNE") {
		logger = slog.Default()
	}
	maxprocs.Configure(
		maxprocs.WithLogger(logger),
	)
	memlimit.Configure(
		memlimit.WithLogger(logger),
	)
}
