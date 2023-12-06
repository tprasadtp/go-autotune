// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package maxprocs

import "log/slog"

type config struct {
	Logger *slog.Logger
}

type CPUQuota interface {
	CPUQuota() (float64, bool, error)
}
