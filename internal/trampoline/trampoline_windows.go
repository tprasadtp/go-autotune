// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package trampoline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
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

	"github.com/tprasadtp/go-autotune/internal/shared"
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

func pipeStream(ctx context.Context, tb testing.TB, wg *sync.WaitGroup, h windows.Handle, w io.Writer) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			// Read any pending data in pipe.
			buf, err := readPipe(h)
			if err != nil {
				tb.Errorf("Failed to read pending data from pipe: %s", err)
			}
			if len(buf) > 0 {
				_, _ = w.Write(buf)
			}
			return
		default:
			buf, err := readPipe(h)
			if err != nil {
				tb.Errorf("Failed to read from pipe: %s", err)
				return
			}
			if len(buf) > 0 {
				_, _ = w.Write(buf)
			}
			// Avoid CPU bounds, as anonymous pipes do not support overlapped i/o.
			//nolint:forbidigo //ignore
			time.Sleep(time.Millisecond * 10)
		}
	}
}

// trampoline runs test function fn via Windows jobobject API.
//
//nolint:funlen // ignore
func trampoline(tb testing.TB, opts Options, verify func(tb testing.TB), configure func()) {
	if verify == nil {
		tb.Fatalf("verify function is nil")
	}

	// If trampoline is defined, run the given test function.
	if _, ok := os.LookupEnv("GO_TEST_EXEC_TRAMPOLINE"); ok {
		// If fn hook is specified, then, run it.
		// This is typically the function which sets GOMAXPROCS and GOMEMLIMIT.
		// This can be nil, if its already set via import side effects.
		if configure != nil {
			configure()
		}

		// verify is a test assertion function.
		verify(tb)
		return
	}

	// Options default overrides.
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second * 30
	}

	// Env variables
	envv := os.Environ()
	envv = append(envv, opts.Env...)

	for _, item := range envv {
		if strings.Contains(strings.ToUpper(item), "GO_TEST_EXEC_TRAMPOLINE") {
			tb.Fatalf("env GO_TEST_EXEC_TRAMPOLINE is already defined")
		}
	}
	envv = append(envv, "GO_TEST_EXEC_TRAMPOLINE=true")

	// Skip if available CPUs < configured CPUs.
	if opts.CPU > float64(runtime.NumCPU()) {
		tb.Skipf("CPU=%f > runtime.NumCPU(%d)", opts.CPU, runtime.NumCPU())
	}

	// Set timeouts.
	//
	// Ideally we would set per set timeouts, but they are not available yet.
	// See https://github.com/golang/go/issues/48157 for more info.
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// Trampoline exe
	exe, err := os.Executable()
	if err != nil {
		tb.Fatalf("Failed to find test exe: %s", err)
	}

	// Trampoline args.
	args := []string{
		strconv.Quote(exe),
		fmt.Sprintf(`-test.run=^%s$`, tb.Name()),
		fmt.Sprintf("-test.timeout=%s", opts.Timeout),
		"--test.v",
	}

	// The return value will be empty if test coverage is not enabled.
	if v := CoverDir(tb); v != "" {
		args = append(args, fmt.Sprintf("-test.gocoverdir=%s", v))
	}

	// Generate a random task name.
	//nolint:gosec // ignore
	jname, err := windows.UTF16PtrFromString(fmt.Sprintf("go-autotune-trampoline-%d", rand.Int()))
	if err != nil {
		tb.Fatalf("UTF16PtrFromString: %s", err)
	}

	// CreateJobObject
	hJobObject, err := windows.CreateJobObject(
		newSecurityAttributes(false),
		jname,
	)
	if err != nil {
		tb.Fatalf("CreateJobObject: %s", err)
	}
	tb.Cleanup(func() {
		err = windows.TerminateJobObject(hJobObject, 1)
		if err != nil {
			tb.Logf("TerminateJobObject: %s", err)
		}
		_ = windows.CloseHandle(hJobObject)
	})

	limit := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE

	// Add memory limits if any
	if opts.M1 > 0 {
		limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY
		limit.ProcessMemoryLimit = uintptr(opts.M1)
	}

	if opts.M2 > 0 {
		limit.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_JOB_MEMORY
		limit.JobMemoryLimit = uintptr(opts.M2)
	}

	v1, err := windows.SetInformationJobObject(
		hJobObject,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limit)),
		uint32(unsafe.Sizeof(limit)),
	)
	if err != nil {
		tb.Fatalf("SetInformationJobObject(Memory): %s", err)
	}
	if v1 == 0 {
		tb.Fatalf("SetInformationJobObject(Memory): %d", v1)
	}

	// Add CPU limits if specified.
	if opts.CPU > 0 {
		cpuLimitInfo := shared.JOBOBJECT_CPU_RATE_CONTROL_INFORMATION{
			ControlFlags: shared.JOB_OBJECT_CPU_RATE_CONTROL_ENABLE | shared.JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP,
			Value: func() uint32 {
				// Scaled for 10000.
				v := opts.CPU / float64(runtime.NumCPU()) * 10000
				if v < 10000 {
					return uint32(math.Round(v)) // round off to 4 digits.
				}
				panic(
					fmt.Sprintf("CPUs=%f, JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP > 10000", opts.CPU),
				)
			}(),
		}
		ret, err := windows.SetInformationJobObject(
			hJobObject,
			windows.JobObjectCpuRateControlInformation,
			uintptr(unsafe.Pointer(&cpuLimitInfo)),
			uint32(unsafe.Sizeof(cpuLimitInfo)),
		)
		// Return value is non zero if SetInformationJobObject succeeds.
		// https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-setinformationjobobject#return-value
		if err != nil || ret == 0 {
			tb.Fatalf("SetInformationJobObject(CPU): %s", err)
		}
	}

	// Use PROC_THREAD_ATTRIBUTE_JOB_LIST to ensure
	// race-free way on starting a process within a JOBOBJECT.
	// Inspired by https://devblogs.microsoft.com/oldnewthing/20230209-00/?p=107812
	procThreadAttrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		tb.Fatalf("NewProcThreadAttributeList(Task): %s", err)
	}

	err = procThreadAttrs.Update(
		shared.PROC_THREAD_ATTRIBUTE_JOB_LIST,
		unsafe.Pointer(&hJobObject),
		unsafe.Sizeof(hJobObject),
	)
	if err != nil {
		tb.Fatalf("UpdateProcThreadAttribute(Task): %s", err)
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
		tb.Fatalf("Failed to create pipe: %s", err)
	}
	tb.Cleanup(func() {
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
		tb.Fatalf("Failed to create pipe: %s", err)
	}
	tb.Cleanup(func() {
		_ = windows.CloseHandle(stderrPipeRead)
		_ = windows.CloseHandle(stderrPipeWrite)
	})

	// Extended startup info.
	processInfo := &windows.ProcessInformation{}
	startupInfoEx := windows.StartupInfoEx{}
	startupInfoEx.Cb = uint32(unsafe.Sizeof(startupInfoEx))
	startupInfoEx.Flags = windows.STARTF_USESTDHANDLES
	startupInfoEx.StdOutput = stdoutPipeWrite
	startupInfoEx.StdErr = stderrPipeWrite
	//nolint:govet // unusedwrite: ProcThreadAttributeList will be read by syscall
	startupInfoEx.ProcThreadAttributeList = procThreadAttrs.List()

	// Build args ptr
	// argsPtr, err := windows.UTF16PtrFromString(`"C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" dir env:`)
	argsPtr, err := windows.UTF16PtrFromString(strings.Join(args, " "))
	if err != nil {
		tb.Fatalf("UTF16PtrFromString(Args): %s", err)
	}

	tb.Logf("Running via trampoline =%v", args)
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
		tb.Fatalf("Failed to create process: %s", err)
	}

	// Don't need the thread handle for anything.
	tb.Cleanup(func() {
		_ = windows.CloseHandle(processInfo.Thread)
	})

	// Re-use *os.Process to avoid reinventing the wheel here.
	process, err := os.FindProcess(int(processInfo.ProcessId))
	if err != nil {
		// If we can't find the process via os.FindProcess,
		// terminate the process as that's what we rely on for all further operations
		// on the object.
		if tErr := windows.TerminateProcess(processInfo.Process, 1); tErr != nil {
			tb.Fatalf("failed to terminate process after process not found: %s", tErr)
		}
		tb.Fatalf("failed to find process after starting: %s", err)
	}

	if process == nil {
		tb.Fatalf("Process did not start")
	}

	// Stream output from trampoline to t.Log via windows pipes.
	var wg sync.WaitGroup
	stdout := NewWriter(tb, "stdout")
	wg.Add(1)
	go pipeStream(ctx, tb, &wg, stdoutPipeRead, stdout)

	stderr := NewWriter(tb, "stderr")
	wg.Add(1)
	go pipeStream(ctx, tb, &wg, stderrPipeRead, stderr)

	procState, err := process.Wait()
	if err != nil {
		tb.Errorf("Error calling Wait: %s", err)
	}

	// Stop pipe i/o go routines.
	cancel()

	// Wait for pipe reader goroutines to complete
	// we wait first then check procState to give time for pipes
	// to write to the logger.
	wg.Wait()
	if !procState.Success() {
		tb.Errorf("Trampoline returned: %d", procState.ExitCode())
	}
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
