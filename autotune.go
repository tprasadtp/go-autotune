// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune

import (
	"github.com/tprasadtp/go-autotune/maxprocs"
	"github.com/tprasadtp/go-autotune/memlimit"
)

func init() {
	maxprocs.Configure()
	memlimit.Configure()
}
