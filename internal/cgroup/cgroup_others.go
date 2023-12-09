// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build !linux

package cgroup

func getInfo(mountInfoFile string, cgroupFile string) (*Info, error) {
	return nil, fmt.Errorf("cgroup: unsupported platform(%s): %w", runtime.GOOS, errors.ErrUnsupported)
}
