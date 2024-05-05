// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build windows

package scenarios

import (
	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/internal/trampoline"
)

// windows scenarios.
func PlatformSpecific() []trampoline.Scenario {
	return []trampoline.Scenario{
		{
			Name:   "Env/GOMAXPROCS=OverrideCPULimits",
			Verify: VerifyFunc(2, 0),
			Opts: trampoline.Options{
				Env: []string{"GOMAXPROCS=2", "GOAUTOTUNE=debug"},
				CPU: 0.5,
			},
		},
		{
			Name:   "Windows/CPU=0.5",
			Verify: VerifyFunc(1, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 0.5,
			},
		},
		{
			Name:   "Windows/CPU=1",
			Verify: VerifyFunc(1, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 1,
			},
		},
		{
			Name:   "Windows/CPU=1.5",
			Verify: VerifyFunc(2, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 1.5,
			},
		},
		{
			Name:   "Windows/CPU=2.25",
			Verify: VerifyFunc(3, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 2.25,
			},
		},
		{
			Name:   "Windows/CPU=2.9",
			Verify: VerifyFunc(3, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 2.9,
			},
		},
		{
			Name:   "Windows/JobMemory=250M",
			Verify: VerifyFunc(3, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 2.9,
			},
		},
		{
			Name:   "DisableViaEnv/GOAUTOTUNE=off",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=off"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GOAUTOTUNE=0",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=0"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GOAUTOTUNE=false",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=false"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GO_AUTOTUNE=0",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GO_AUTOTUNE=0"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GO_AUTOTUNE=false",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GO_AUTOTUNE=false"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
	}
}
