// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux

package autotune

import (
	"log/slog"

	"github.com/tprasadtp/go-autotune/internal/env"
	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func configure() {
	if env.IsFalse("GO_AUTOTUNE") || env.IsFalse("GOAUTOTUNE") {
		return
	}

	var logger *slog.Logger
	if env.IsDebug("GO_AUTOTUNE") || env.IsDebug("GOAUTOTUNE") {
		logger = slog.Default()
	}
	maxprocs.Configure(
		maxprocs.WithLogger(logger),
	)
	memlimit.Configure(
		memlimit.WithLogger(logger),
	)
}
