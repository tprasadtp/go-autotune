// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package shared

// import (
// 	"fmt"
// 	"os"
// 	"strings"
// 	"testing"

// 	"golang.org/x/sys/windows"
// )

// // https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
// type JOBOBJECT_CPU_RATE_CONTROL_INFORMATION struct {
// 	ControlFlags uint32
// 	Value        uint32
// }

// // https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
// const (
// 	JOB_OBJECT_CPU_RATE_CONTROL_ENABLE uint32 = 1 << iota
// 	JOB_OBJECT_CPU_RATE_CONTROL_WEIGHT_BASED
// 	JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP
// 	JOB_OBJECT_CPU_RATE_CONTROL_NOTIFY
// 	JOB_OBJECT_CPU_RATE_CONTROL_MIN_MAX_RATE
// )

// // SystemdRun runs test function fn via systemd-run.
// func WindowsRun(t *testing.T, cpu, mem, memProc uintptr, fn func(t *testing.T)) {
// 	t.Helper()
// 	if fn == nil {
// 		t.Fatalf("fn function is nil")
// 	}

// 	trampoline := strings.TrimSpace(strings.ToLower(os.Getenv("GO_TEST_EXEC_TRAMPOLINE")))

// 	// If trampoline is true, run the given test function.
// 	if trampoline == "true" {
// 		fn(t)
// 		return
// 	}

// 	// Trampoline exe
// 	arg0, err := os.Executable()
// 	if err != nil {
// 		t.Fatalf("Failed to find exe: %s", err)
// 	}

// 	// Trampoline args.
// 	args := []string{
// 		fmt.Sprintf("--test.run=^%s$", t.Name()),
// 	}

// 	// Add verbose flag if test also mentions it.
// 	if TestingIsVerbose() {
// 		args = append(args, "--test.v=true")
// 	}

// 	// The return value will be empty if test coverage is not enabled.
// 	if TestingCoverDir(t) != "" {
// 		args = append(args, fmt.Sprintf("--test.gocoverdir=%s", TestingCoverDir(t)))
// 	}

// 	binaryPtr, err := windows.UTF16PtrFromString(arg0)
// 	if err != nil {
// 		t.Fatalf("UTF16PtrFromString: %s", err)
// 	}

// 	argsPtr, err := windows.UTF16PtrFromString(strings.Join(args, " "))
// 	if err != nil {
// 		t.Fatalf("UTF16PtrFromString: %s", err)
// 	}

// 	objHandle, err := windows.CreateJobObject(nil, nil)
// 	if err != nil {
// 		t.Fatalf("CreateJobObject: %s", err)
// 	}

// 	elimit := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
// 		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
// 			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
// 		},
// 	}

// 	if mem > 0 {
// 		elimit.BasicLimitInformation.LimitFlags = elimit.BasicLimitInformation.LimitFlags | windows.JOB_OBJECT_LIMIT_JOB_MEMORY
// 	}
// 	if memProc > 0 {
// 		elimit.BasicLimitInformation.LimitFlags = elimit.BasicLimitInformation.LimitFlags | windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY
// 	}

// 	attrs, err := windows.NewProcThreadAttributeList(3)
// 	if err != nil {
// 		t.Fatalf("NewProcThreadAttributeList: %s", err)
// 	}

// }
