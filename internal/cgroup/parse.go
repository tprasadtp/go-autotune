// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package cgroup

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

// A few specific characters in mountinfo path entries (root and mountpoint)
// are escaped using a backslash followed by a character's ascii code in octal.
//
//	space              -- as \040
//	tab (aka \t)       -- as \011
//	newline (aka \n)   -- as \012
//	backslash (aka \\) -- as \134
//
// This function un-escapes the above sequences.
func unescape(path string) (string, error) {
	if strings.IndexByte(path, '\\') == -1 {
		return path, nil
	}

	// The following code is UTF-8 transparent as it only looks for some
	// specific characters (backslash and 0..7) with values < utf8.RuneSelf,
	// and everything else is passed through as is.
	buf := make([]byte, len(path))
	bufLen := 0
	for i := 0; i < len(path); i++ {
		if path[i] != '\\' {
			buf[bufLen] = path[i]
			bufLen++
			continue
		}
		s := path[i:]
		if len(s) < 4 {
			// too short
			return "", fmt.Errorf("bad escape sequence %q: too short", s)
		}
		c := s[1]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			v := c - '0'
			for j := 2; j < 4; j++ { // one digit already; two more
				if s[j] < '0' || s[j] > '7' {
					return "", fmt.Errorf("bad escape sequence %q: not a digit", s[:3])
				}
				x := s[j] - '0'
				v = (v << 3) | x
			}
			if v > 255 {
				return "", fmt.Errorf("bad escape sequence %q: out of range" + s[:3])
			}
			buf[bufLen] = v
			bufLen++
			i += 3
			continue
		default:
			return "", fmt.Errorf("bad escape sequence %q: not a digit" + s[:3])
		}
	}

	return string(buf[:bufLen]), nil
}

// nameFromFile returns cgroup name for given cgroup file.
func nameFromFile(path string) (string, error) {
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
	// This file describes control groups to which the process with the corresponding PID belongs. The displayed information differs for cgroups version 1 and version 2 hierarchies.
	// For each cgroup hierarchy of which the process is a member, there is one entry containing three colon-separated fields:
	//
	// hierarchy-ID:controller-list:cgroup-path
	//
	// The colon-separated fields are, from left to right:
	//
	// 1. For the cgroups version 2 hierarchy, this field contains the value 0.
	// 2. For the cgroups version 2 hierarchy, this field is empty.
	// 3. This field contains the pathname of the control group in the hierarchy to which the process belongs.
	//    This pathname is relative to the mount point of the hierarchy.
	//
	// https://manpages.debian.org/buster/manpages/cgroups.7.en.html
	if !bytes.HasPrefix(contents, []byte("0::")) {
		return "", fmt.Errorf("cgroup file(%s) is missing prefix 0::", path)
	}

	rv := string(bytes.TrimSpace(bytes.TrimPrefix(contents, []byte("0::"))))
	if strings.Contains(rv, "\n") {
		return "", fmt.Errorf("cgroup file(%s) contains newlines", path)
	}

	return rv, nil
}

// mountPointFromFile parses given mountinfo file and extracts cgroup v2 mountpoint
// from it.
func mountPointFromFile(mountInfo string) (string, error) {
	// Try to open file.
	file, err := os.Open(mountInfo)
	if err != nil {
		return "", fmt.Errorf("failed to open: %w", err)
	}
	defer file.Close()

	// If file is too large do not read it.
	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to check file size: %w", err)
	}

	if stat.Size() > 1e6 {
		return "", fmt.Errorf("mountinfo file too large: %d", stat.Size())
	}

	// Read mount info file
	// See https://manpages.debian.org/buster/manpages/proc.5.en.html
	// The file contains lines of the form:

	// 36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
	// (1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

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
			return "", fmt.Errorf("parsing '%q' failed: not enough fields (%d)", text, numFields)
		}

		// separator field
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
	return "", fmt.Errorf("unable to find mountpoint for cgroupv2")
}

// getCPUQuotaFromFile reads from path cpu.max quota specified.
// This never returns Nan or Inf.
func getCPUQuotaFromFile(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		// If file is missing then cpu controller is not enabled
		// or cpu limits are not defined.
		if errors.Is(err, os.ErrNotExist) {
			return -1, nil
		}
	}

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || len(fields) > 2 {
			return -1, fmt.Errorf("invalid format cpu.max")
		}

		// No CPU limits.
		if fields[0] == "max" {
			return 0, nil
		}

		// Get Maximum CPU quota
		max, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil || max == 0 {
			return -1, fmt.Errorf("invalid format cpu.max")
		}

		// Check if period is defined.
		var period uint64
		if len(fields) == 2 {
			period, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return -1, fmt.Errorf("invalid format cpu.max: %w", err)
			}
		} else {
			// Default CPU period value.
			period = 100000
		}

		return float64(max) / float64(period), nil
	}

	if err := scanner.Err(); err != nil {
		return -1, fmt.Errorf("failed to scan cpu.max: %w", err)
	}

	return -1, io.ErrUnexpectedEOF
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

		// Get Value
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
