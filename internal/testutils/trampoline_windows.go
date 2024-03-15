// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package shared

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/tprasadtp/go-autotune/internal/env"
	"github.com/tprasadtp/go-autotune/internal/types"
	"golang.org/x/sys/windows"
)

var (
	kernel32          = windows.NewLazySystemDLL("kernel32.dll")
	procPeekNamedPipe = kernel32.NewProc("PeekNamedPipe")
)

func peekNamedPipe(h windows.Handle, buf []byte, lpBytesRead, lpTotalBytesAvail, lpBytesLeftThisMessage *uint32) error {
	var _p0 *byte
	if len(buf) > 0 {
		_p0 = &buf[0]
	}

	r1, _, e1 := syscall.SyscallN(
		procPeekNamedPipe.Addr(),
		uintptr(h),
		uintptr(unsafe.Pointer(_p0)),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(lpBytesRead)),
		uintptr(unsafe.Pointer(lpTotalBytesAvail)),
		uintptr(unsafe.Pointer(lpBytesLeftThisMessage)),
	)
	if r1 == 0 {
		return e1
	}
	return nil
}

func readPipe(h windows.Handle) ([]byte, error) {
	var pending uint32
	var err error

	// Peek a named pipe and check if it has any pending bytes.
	// Do not copy anything to buffer yet.
	err = peekNamedPipe(h, nil, nil, &pending, nil)
	if err != nil && !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, fmt.Errorf("PeekNamedPipe: %w", err)
	}

	// If there are pending bytes, read via ReadFile.
	if pending > 0 {
		buf := make([]byte, pending)
		var n uint32
		err = windows.ReadFile(h, buf, &n, nil)
		if err != nil {
			return nil, fmt.Errorf("ReadFile: %w", err)
		}
		return buf, nil
	}
	return nil, nil
}

func pipeLogger(ctx context.Context, wg *sync.WaitGroup, read windows.Handle, logger *OutputLogger) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			// Read any pending data in pipe.
			buf, err := readPipe(read)
			if err != nil {
				logger.Errorf("Failed to read pending data from pipe: %s", err)
			}
			if len(buf) > 0 {
				logger.LogOutput(buf)
			}
			return
		default:
			buf, err := readPipe(read)
			if err != nil {
				logger.Errorf("Failed to read from pipe: %s", err)
				return
			}
			if len(buf) > 0 {
				logger.LogOutput(buf)
			}
			// Avoid CPU bounds, as anonymous pipes do not support overlapped i/o.
			//nolint:forbidigo //ignore
			time.Sleep(time.Millisecond * 10)
		}
	}
}

// WindowsRun runs test function fn via Windows jobobject API.
//
//nolint:funlen // ignore
func WindowsRun(t *testing.T, cpu float64, mem, memProc int64, autoTuneEnv string, fn func(t *testing.T)) {
	if fn == nil {
		t.Fatalf("fn function is nil")
	}

	// If trampoline is true, run the given test function.
	if env.IsTrue("GO_TEST_EXEC_TRAMPOLINE") {
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

	// Set timeouts.
	//
	// Ideally we would set per set timeouts, but they are not available yet.
	// See https://github.com/golang/go/issues/48157 for more info.
	var ctx context.Context
	var cancel context.CancelFunc
	var timeout time.Duration
	if ts, ok := t.Deadline(); ok {
		// Timeout is derived from test's own timeout.
		timeout = time.Since(ts).Abs()
		ctx, cancel = context.WithDeadline(context.Background(), ts)
	} else {
		timeout = time.Second * 30
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*30)
	}
	defer cancel()

	// Trampoline exe
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("Failed to find test exe: %s", err)
	}

	// Trampoline args.
	trampolineArgs := []string{
		strconv.Quote(exe),
		fmt.Sprintf(`-test.run=^%s$`, t.Name()),
		fmt.Sprintf("-test.timeout=%s", timeout),
		"--test.v",
	}

	// The return value will be empty if test coverage is not enabled.
	if TestingCoverDir(t) != "" {
		trampolineArgs = append(trampolineArgs, fmt.Sprintf("--test.gocoverdir=%s", TestingCoverDir(t)))
	}

	// Generate a random task name.
	rb := make([]byte, 8)
	_, err = rand.Read(rb)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %s", err)
	}
	jobObjectName, err := windows.UTF16PtrFromString(hex.EncodeToString(rb))
	if err != nil {
		t.Fatalf("UTF16PtrFromString: %s", err)
	}

	// CreateJobObject
	hJobObject, err := windows.CreateJobObject(
		newSecurityAttributes(false),
		jobObjectName,
	)
	if err != nil {
		t.Fatalf("CreateJobObject: %s", err)
	}
	t.Cleanup(func() {
		err = windows.TerminateJobObject(hJobObject, 1)
		if err != nil {
			t.Logf("TerminateJobObject: %s", err)
		}
		_ = windows.CloseHandle(hJobObject)
	})

	limit := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE

	// Add memory limits if any
	if mem > 0 {
		limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_JOB_MEMORY
		limit.JobMemoryLimit = uintptr(mem)
	}

	if memProc > 0 {
		limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY
		limit.ProcessMemoryLimit = uintptr(memProc)
	}

	v1, err := windows.SetInformationJobObject(
		hJobObject,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limit)),
		uint32(unsafe.Sizeof(limit)),
	)
	if err != nil {
		t.Fatalf("SetInformationJobObject(Memory): %s", err)
	}
	if v1 == 0 {
		t.Fatalf("SetInformationJobObject(Memory): %d", v1)
	}

	// Add CPU limits if specified.
	if cpu > 0 {
		cpuLimitInfo := types.JOBOBJECT_CPU_RATE_CONTROL_INFORMATION{
			ControlFlags: types.JOB_OBJECT_CPU_RATE_CONTROL_ENABLE | types.JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP,
			Value: func() uint32 {
				v := cpu / float64(runtime.NumCPU()) * 10000
				if v < 10000 {
					return uint32(math.Round(v))
				}
				return 10000
			}(),
		}
		v2, err := windows.SetInformationJobObject(
			hJobObject,
			windows.JobObjectCpuRateControlInformation,
			uintptr(unsafe.Pointer(&cpuLimitInfo)),
			uint32(unsafe.Sizeof(cpuLimitInfo)),
		)
		if err != nil {
			t.Fatalf("SetInformationJobObject(CPU): %s", err)
		}
		if v2 == 0 {
			t.Fatalf("SetInformationJobObject(CPU): %d", v2)
		}
	}

	// Use PROC_THREAD_ATTRIBUTE_JOB_LIST to ensure
	// race-free way on starting a process within a JOBOBJECT.
	procThreadAttrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		t.Fatalf("NewProcThreadAttributeList(Task): %s", err)
	}

	err = procThreadAttrs.Update(
		types.PROC_THREAD_ATTRIBUTE_JOB_LIST,
		unsafe.Pointer(&hJobObject),
		unsafe.Sizeof(hJobObject),
	)
	if err != nil {
		t.Fatalf("UpdateProcThreadAttribute(Task): %s", err)
	}

	// Create pipes for stdout and stderr
	var stdoutPipeWrite windows.Handle
	var stdoutPipeRead windows.Handle
	err = windows.CreatePipe(
		&stdoutPipeRead,
		&stdoutPipeWrite,
		newSecurityAttributes(true),
		0,
	)
	if err != nil {
		t.Fatalf("Failed to create pipe: %s", err)
	}
	t.Cleanup(func() {
		_ = windows.CloseHandle(stdoutPipeRead)
		_ = windows.CloseHandle(stdoutPipeWrite)
	})

	var stderrPipeWrite windows.Handle
	var stderrPipeRead windows.Handle
	err = windows.CreatePipe(
		&stderrPipeRead,
		&stderrPipeWrite,
		newSecurityAttributes(true),
		0,
	)
	if err != nil {
		t.Fatalf("Failed to create pipe: %s", err)
	}
	t.Cleanup(func() {
		_ = windows.CloseHandle(stderrPipeRead)
		_ = windows.CloseHandle(stderrPipeWrite)
	})

	// Extended startup info.
	startupInfoEx := windows.StartupInfoEx{}
	startupInfoEx.Cb = uint32(unsafe.Sizeof(startupInfoEx))
	startupInfoEx.Flags = windows.STARTF_USESTDHANDLES
	startupInfoEx.StdOutput = stdoutPipeWrite
	startupInfoEx.StdErr = stderrPipeWrite
	startupInfoEx.ProcThreadAttributeList = procThreadAttrs.List() //nolint:govet // unusedwrite: ProcThreadAttributeList will be read by syscall
	processInfo := &windows.ProcessInformation{}

	// Build args ptr
	// argsPtr, err := windows.UTF16PtrFromString(`"C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" dir env:`)
	argsPtr, err := windows.UTF16PtrFromString(strings.Join(trampolineArgs, " "))
	if err != nil {
		t.Fatalf("UTF16PtrFromString(Args): %s", err)
	}

	t.Logf("Running via trampoline =%v", trampolineArgs)
	err = windows.CreateProcess(
		nil,
		argsPtr,
		nil,
		nil,
		true,
		windows.EXTENDED_STARTUPINFO_PRESENT|windows.CREATE_UNICODE_ENVIRONMENT,
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
		_ = windows.CloseHandle(processInfo.Thread)
	})

	// Re-use *os.Process to avoid reinventing the wheel here.
	process, err := os.FindProcess(int(processInfo.ProcessId))
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

	var wg sync.WaitGroup

	stdoutLogger := NewOutputLogger(t, "stdout")
	wg.Add(1)
	go pipeLogger(ctx, &wg, stdoutPipeRead, stdoutLogger)

	stderrLogger := NewOutputLogger(t, "stderr")
	wg.Add(1)
	go pipeLogger(ctx, &wg, stderrPipeRead, stderrLogger)

	procState, err := process.Wait()
	if err != nil {
		t.Errorf("Error calling Wait: %s", err)
	}
	if !procState.Success() {
		t.Errorf("Trampoline returned: %d", procState.ExitCode())
	}

	// Stop reader and writer go routines.
	cancel()

	// Wait for pipe reader goroutines to complete
	wg.Wait()
}

// newSecurityAttributes creates a SECURITY_ATTRIBUTES structure, that specifies the
// security descriptor for the job object and determines that child
// processes cannot inherit the handle.
//
// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/legacy/aa379560(v=vs.85)
func newSecurityAttributes(inherit bool) *windows.SecurityAttributes {
	var sa windows.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	if inherit {
		sa.InheritHandle = 1
	}
	return &sa
}

// createEnvBlock converts an array of environment strings into
// the representation required by CreateProcess: a sequence of NUL
// terminated strings followed by a nil.
func createEnvBlock(envv []string) *uint16 {
	if len(envv) == 0 {
		return &utf16.Encode([]rune("\x00\x00"))[0]
	}
	length := 0
	for _, s := range envv {
		length += len(s) + 1
	}
	length++

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
