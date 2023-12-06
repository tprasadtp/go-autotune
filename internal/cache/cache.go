// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cache

import (
	"sync"

	"github.com/tprasadtp/go-autotune/internal/cgroup"
)

var (
	info *cgroup.Info
	err  error
	once sync.Once
)

// GetCgroupInfo is similar to [github.com/tprasadtp/go-autotune/internal/cgroup.GetInfo]
// but executes only once.
func GetCgroupInfo() (*cgroup.Info, error) {
	once.Do(func() {
		info, err = cgroup.GetInfo("", "")
	})
	//nolint:wrapcheck // cached value
	return info, err
}
