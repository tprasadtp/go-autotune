// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package platform

func getCPUQuota(options ...Option) (float64, error) {
	return getCPUQuotaFromCgroup(options...)
}

//nolint:nonamedreturns // for docs.
func getMemoryQuota(options ...Option) (max, high int64, err error) {
	return getMemoryQuotaFromCgroup(options...)
}
