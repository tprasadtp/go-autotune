// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package memlimit

import "log/slog"

type config struct {
	Logger *slog.Logger
}

type MemLimit interface {
	MemLimit() (uint64, bool, error)
}
