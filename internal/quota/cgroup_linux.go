// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package quota

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetCgroupInterfacePath returns base path of cgroup interface files.
// If procfs is empty, /self/proc is assumed.
func GetCgroupInterfacePath(procfs string) (string, error) {
	if procfs == "" {
		procfs = "/proc/self"
	}

	mount, err := cgroupMountPointFromFile(filepath.Join(procfs, "mountinfo"))
	if err != nil {
		return "", fmt.Errorf("quota(cgroup): failed to get cgroup2 mountpoint: %w", err)
	}

	name, err := cgroupNameFromFile(filepath.Join(procfs, "cgroup"))
	if err != nil {
		return "", fmt.Errorf("quota(cgroup): failed to get cgroup name: %w", err)
	}

	return filepath.Join(mount, name), nil
}

// cgroupNameFromFile returns cgroup name for given cgroup file.
func cgroupNameFromFile(path string) (string, error) {
	// Try to open file.
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer file.Close()

	// If file is too large do not read it.
	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to check file size: %w", err)
	}

	if stat.Size() > 1e3 {
		return "", fmt.Errorf("file too large: %d", stat.Size())
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read contents for cgroup file %s: %w", path, err)
	}

	// /proc/self/cgroup (since Linux 2.6.24)
	// This file describes control groups to which the process with the corresponding PID
	// belongs. The displayed information differs for cgroups version 1 and version 2 hierarchies.
	// For each cgroup hierarchy of which the process is a member, there is one entry containing
	// three colon-separated fields:
	//
	// hierarchy-ID:controller-list:cgroup-path
	//
	// The colon-separated fields are, from left to right:
	//
	// 1. For the cgroups version 2 hierarchy, this field contains the value 0.
	// 2. For the cgroups version 2 hierarchy, this field is empty.
	// 3. This field contains the pathname of the control group in the hierarchy to which the
	//    process belongs. This pathname is relative to the mount point of the hierarchy.
	//
	// https://manpages.debian.org/buster/manpages/cgroups.7.en.html
	if !bytes.HasPrefix(contents, []byte("0::")) {
		return "", fmt.Errorf("missing prefix '0::' from cgroup file(%s)", path)
	}

	rv := string(bytes.TrimSpace(bytes.TrimPrefix(contents, []byte("0::"))))
	if strings.Contains(rv, "\n") {
		return "", fmt.Errorf("cgroup file(%s) contains newlines", path)
	}

	return rv, nil
}

// cgroupMountPointFromFile parses given mountinfo file and extracts cgroup
// v2 mountpoint from it.
func cgroupMountPointFromFile(mountInfo string) (string, error) {
	file, err := os.Open(mountInfo)
	if err != nil {
		return "", fmt.Errorf("failed to open: %w", err)
	}
	defer file.Close()

	// Read mount info file
	// See https://manpages.debian.org/buster/manpages/proc.5.en.html
	//
	// The file contains lines of the form:
	//
	// 36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
	// (1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)
	//
	// The numbers in parentheses are labels for the descriptions
	// below:
	//
	// (1)  mount ID: a unique ID for the mount (may be reused
	//      after umount(2)).
	//
	// (2)  parent ID: the ID of the parent mount (or of self for
	//      the root of this mount namespace's mount tree).
	//
	// (3)  major:minor: the value of st_dev for files on this
	//      filesystem.
	//
	// (4)  root: the pathname of the directory in the filesystem
	//      which forms the root of this mount.
	//
	// (5)  mount point: the pathname of the mount point relative
	//      to the process's root directory.
	//
	// (6)  mount options: per-mount options .
	//
	// (7)  optional fields: zero or more fields of the form
	//      "tag[:value]"; see below.
	//
	// (8)  separator: the end of the optional fields is marked
	//      by a single hyphen.
	//
	// (9)  filesystem type: the filesystem type in the form
	//      "type[.subtype]".
	//
	// (10) mount source: filesystem-specific information or
	//      "none".
	//
	// (11) super options: per-superblock options.
	s := bufio.NewScanner(file)
	for s.Scan() {
		var err error
		text := s.Text()
		fields := strings.Split(text, " ")
		numFields := len(fields)
		if numFields < 10 {
			// Should be at least 10 fields
			return "", fmt.Errorf("parsing '%s' failed: not enough fields (%d)", text, numFields)
		}

		// Separator field
		sepIdx := numFields - 4

		// Check if type is valid
		fsType, err := unescape(fields[sepIdx+1])
		if err != nil {
			return "", fmt.Errorf("parsing '%s' failed: fstype: %w", fields[sepIdx+1], err)
		}

		mountpoint, err := unescape(fields[4])
		if err != nil {
			return "", fmt.Errorf("parsing '%s' failed: mount point: %w", fields[4], err)
		}

		if fsType == "cgroup2" {
			return mountpoint, nil
		}
	}
	return "", errors.New("unable to find cgroup2 mountpoint")
}

func memLimitFromFile(path string) (int64, error) {
	file, err := os.Open(path)
	base := filepath.Base(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
	}

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 1 {
			return 0, fmt.Errorf("invalid format %s", base)
		}

		// No memory limits.
		if fields[0] == "max" {
			return 0, nil
		}

		max, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil || max < 0 {
			return 0, fmt.Errorf("invalid format %s: %w", base, err)
		}

		return max, nil
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan %s: %w", base, err)
	}

	return 0, io.ErrUnexpectedEOF
}
