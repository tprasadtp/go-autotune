// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package shared

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
type JOBOBJECT_CPU_RATE_CONTROL_INFORMATION struct {
	ControlFlags uint32
	Value        uint32
}

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
const (
	JOB_OBJECT_CPU_RATE_CONTROL_ENABLE uint32 = 1 << iota
	JOB_OBJECT_CPU_RATE_CONTROL_WEIGHT_BASED
	JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP
	JOB_OBJECT_CPU_RATE_CONTROL_NOTIFY
	JOB_OBJECT_CPU_RATE_CONTROL_MIN_MAX_RATE
)

// https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-updateprocthreadattribute
const (
	PROC_THREAD_ATTRIBUTE_JOB_LIST = 0x2000D
)

// WindowsRun runs test function fn via windows jobobject API.
func WindowsRun(t *testing.T, cpu float64, mem, memProc int64, autoTuneEnv string, fn func(t *testing.T)) {
	if fn == nil {
		t.Fatalf("fn function is nil")
	}

	// If trampoline is true, run the given test function.
	if IsTrue("GO_TEST_EXEC_TRAMPOLINE") {
		t.Logf("Running test function...")
		fn(t)
		return
	}

	// Env variables
	envv := os.Environ()
	for _, item := range envv {
		if strings.Contains(strings.ToUpper(item), "GO_TEST_EXEC_TRAMPOLINE") {
			t.Fatalf("env GO_TEST_EXEC_TRAMPOLINE is already defined")
		}
	}
	envv = append(envv, "GO_TEST_EXEC_TRAMPOLINE=true")
	if autoTuneEnv != "" {
		envv = append(envv, fmt.Sprintf("GOAUTOTUNE=%s", autoTuneEnv))
	}

	var process *os.Process

	// Trampoline exe
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("Failed to find test exe: %s", err)
	}

	// Trampoline args.
	trampolineArgs := []string{
		strconv.Quote(exe),
		fmt.Sprintf(`--test.run=^%s$`, t.Name()),
	}

	// Add verbose flag if test also mentions it.
	if TestingIsVerbose() {
		trampolineArgs = append(trampolineArgs, "--test.v=true")
	}

	// The return value will be empty if test coverage is not enabled.
	if TestingCoverDir(t) != "" {
		trampolineArgs = append(trampolineArgs, fmt.Sprintf("--test.gocoverdir=%s", TestingCoverDir(t)))
	}

	// Task name is derived from test name.
	jobObjectName, err := windows.UTF16PtrFromString(t.Name())
	if err != nil {
		t.Fatalf("Invalid Task name: %s", err)
	}

	// CreateJobObject
	hJobObject, err := windows.CreateJobObject(NewSecurityAttributes(), jobObjectName)
	if err != nil {
		t.Fatalf("CreateJobObject: %s", err)
	}
	t.Cleanup(func() {
		err = windows.TerminateJobObject(hJobObject, 1)
		if err != nil {
			t.Logf("TerminateJobObject: %s", err)
		}
		windows.CloseHandle(hJobObject)
	})

	limit := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE

	// // Add memory limits if any
	if mem > 0 {
		limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_JOB_MEMORY
		limit.JobMemoryLimit = uintptr(mem)
	}

	if memProc > 0 {
		limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY
		limit.ProcessMemoryLimit = uintptr(memProc)
	}

	rv, err := windows.SetInformationJobObject(
		hJobObject,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limit)),
		uint32(unsafe.Sizeof(limit)),
	)
	if err != nil {
		t.Fatalf("SetInformationJobObject(Memory): %s", err)
	}
	if rv == 0 {
		t.Fatalf("SetInformationJobObject(Memory): %d", rv)
	}

	// // Add CPU limits if any.
	if cpu > 0 {
		cpuLimitInfo := JOBOBJECT_CPU_RATE_CONTROL_INFORMATION{
			ControlFlags: JOB_OBJECT_CPU_RATE_CONTROL_ENABLE | JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP,
			Value: func() uint32 {
				v := cpu / float64(runtime.NumCPU()) * 10000
				if v < 10000 {
					return uint32(math.Round(v))
				}
				return 10000
			}(),
		}
		rv, err := windows.SetInformationJobObject(
			hJobObject,
			windows.JobObjectCpuRateControlInformation,
			uintptr(unsafe.Pointer(&cpuLimitInfo)),
			uint32(unsafe.Sizeof(cpuLimitInfo)),
		)
		if err != nil {
			t.Fatalf("SetInformationJobObject(CPU): %s", err)
		}
		if rv == 0 {
			t.Fatalf("SetInformationJobObject(CPU): %d", rv)
		}
	}

	// // Use PROC_THREAD_ATTRIBUTE_JOB_LIST to ensure race-free way on starting
	// // process within a JOBOBJECT.
	procThreadAttrs, err := windows.NewProcThreadAttributeList(3)
	if err != nil {
		t.Fatalf("NewProcThreadAttributeList(Task): %s", err)
	}

	err = procThreadAttrs.Update(
		PROC_THREAD_ATTRIBUTE_JOB_LIST,
		unsafe.Pointer(&hJobObject),
		unsafe.Sizeof(hJobObject),
	)
	if err != nil {
		t.Fatalf("UpdateProcThreadAttribute(Task): %s", err)
	}

	// Extended startup info.
	startupInfoEx := windows.StartupInfoEx{}
	startupInfoEx.Cb = uint32(unsafe.Sizeof(startupInfoEx))
	// startupInfoEx.Flags = windows.STARTF_USESTDHANDLES
	startupInfoEx.ProcThreadAttributeList = procThreadAttrs.List() //nolint:govet // unusedwrite: ProcThreadAttributeList will be read by syscall
	processInfo := &windows.ProcessInformation{}

	// Build args ptr
	argsPtr, err := windows.UTF16PtrFromString(strings.Join(trampolineArgs, " "))
	// argsPtr, err := windows.UTF16PtrFromString(`"C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" dir env:`)
	if err != nil {
		t.Fatalf("UTF16PtrFromString(Args): %s", err)
	}

	t.Logf("Running via trampoline exe=%v", trampolineArgs)
	err = windows.CreateProcess(
		nil,
		argsPtr,
		nil,
		nil,
		false,
		windows.EXTENDED_STARTUPINFO_PRESENT|uint32(windows.CREATE_UNICODE_ENVIRONMENT),
		createEnvBlock(addCriticalEnv((envv))),
		nil,
		&startupInfoEx.StartupInfo,
		processInfo,
	)

	if err != nil {
		t.Fatalf("Failed to create process: %s", err)
	}

	// Don't need the thread handle for anything.
	t.Cleanup(func() {
		_ = windows.CloseHandle(windows.Handle(processInfo.Thread))
	})

	// Re-use *os.Process to avoid reinventing the wheel here.
	process, err = os.FindProcess(int(processInfo.ProcessId))
	if err != nil {
		// If we can't find the process via os.FindProcess,
		// terminate the process as that's what we rely on for all further operations on the
		// object.
		if tErr := windows.TerminateProcess(processInfo.Process, 1); tErr != nil {
			t.Fatalf("failed to terminate process after process not found: %s", tErr)
		}
		t.Fatalf("failed to find process after starting: %s", err)
	}

	if process == nil {
		t.Fatalf("Process did not start")
	}

	procState, err := process.Wait()
	if err != nil {
		t.Fatalf("Error calling Wait: %s", err)
	}
	if !procState.Success() {
		t.Fatalf("Trampoline returned: %d", procState.ExitCode())
	}
}

// NewSecurityAttributes creates a SECURITY_ATTRIBUTES structure, that specifies the
// security descriptor for the job object and determines that child
// processes can not inherit the handle.
//
// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/legacy/aa379560(v=vs.85)
func NewSecurityAttributes() *windows.SecurityAttributes {
	var sa windows.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 0
	return &sa
}

// createEnvBlock converts an array of environment strings into
// the representation required by CreateProcess: a sequence of NUL
// terminated strings followed by a nil.
// Last bytes are two UCS-2 NULs, or four NUL bytes.
func createEnvBlock(envv []string) *uint16 {
	if len(envv) == 0 {
		return &utf16.Encode([]rune("\x00\x00"))[0]
	}
	length := 0
	for _, s := range envv {
		length += len(s) + 1
	}
	length += 1

	b := make([]byte, length)
	i := 0
	for _, s := range envv {
		l := len(s)
		copy(b[i:i+l], []byte(s))
		copy(b[i+l:i+l+1], []byte{0})
		i = i + l + 1
	}
	copy(b[i:i+1], []byte{0})

	return &utf16.Encode([]rune(string(b)))[0]
}

// addCriticalEnv adds any critical environment variables that are required
// (or at least almost always required) on the operating system.
func addCriticalEnv(env []string) []string {
	for _, kv := range env {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			continue
		}
		k := kv[:eq]
		if strings.EqualFold(k, "SYSTEMROOT") {
			// We already have it.
			return env
		}
	}
	return append(env, "SYSTEMROOT="+os.Getenv("SYSTEMROOT"))
}
