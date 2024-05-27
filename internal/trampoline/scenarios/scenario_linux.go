// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

//go:build linux

package scenarios

import (
	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/internal/trampoline"
)

// Linux scenarios.
func PlatformSpecific() []trampoline.Scenario {
	return []trampoline.Scenario{
		{
			Name:   "Env/GOMAXPROCS=OverrideCPULimits",
			Verify: trampoline.VerifyFunc(2, 0),
			Opts: trampoline.Options{
				Env: []string{"GOMAXPROCS=2", "GOAUTOTUNE=debug"},
				CPU: 0.5,
			},
		},
		{
			Name:   "Linux/CPU=0.5",
			Verify: trampoline.VerifyFunc(1, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 0.5,
			},
		},
		{
			Name:   "Linux/CPU=1",
			Verify: trampoline.VerifyFunc(1, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 1,
			},
		},
		{
			Name:   "Linux/CPU=1.5",
			Verify: trampoline.VerifyFunc(2, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 1.5,
			},
		},
		{
			Name:   "Linux/CPU=2.25",
			Verify: trampoline.VerifyFunc(3, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 2.25,
			},
		},
		{
			Name:   "Linux/CPU=2.9",
			Verify: trampoline.VerifyFunc(3, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				CPU: 2.9,
			},
		},
		{
			Name:   "Linux/MemoryMax=250M",
			Verify: trampoline.VerifyFunc(0, 225*shared.MiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				M1:  250 * shared.MiByte,
			},
		},
		{
			Name:   "Linux/MemoryMax=5G",
			Verify: trampoline.VerifyFunc(0, 5*shared.GiByte-100*shared.MiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				M1:  5 * shared.GiByte,
			},
		},
		{
			Name:   "Linux/MemoryHigh=250M",
			Verify: trampoline.VerifyFunc(0, 250*shared.MiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "Linux/MemoryHigh=5G",
			Verify: trampoline.VerifyFunc(0, 5*shared.GiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				M2:  5 * shared.GiByte,
			},
		},
		{
			Name:   "Linux/MemoryMax=250M/MemoryHigh=250M",
			Verify: trampoline.VerifyFunc(0, 225*shared.MiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "Linux/MemoryMax=300M/MemoryHigh=250M",
			Verify: trampoline.VerifyFunc(0, 250*shared.MiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				M1:  300 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		// This is probably an invalid configuration. MemoryHigh should take precedence here.
		{
			Name:   "Linux/MemoryMax=250M/MemoryHigh=300M",
			Verify: trampoline.VerifyFunc(0, 225*shared.MiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
				M1:  250 * shared.MiByte,
				M2:  300 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GOAUTOTUNE=off",
			Verify: trampoline.VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=off"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GOAUTOTUNE=0",
			Verify: trampoline.VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=0"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GOAUTOTUNE=false",
			Verify: trampoline.VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=false"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GO_AUTOTUNE=0",
			Verify: trampoline.VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GO_AUTOTUNE=0"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
		{
			Name:   "DisableViaEnv/GO_AUTOTUNE=false",
			Verify: trampoline.VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GO_AUTOTUNE=false"},
				CPU: 1.5,
				M1:  250 * shared.MiByte,
				M2:  250 * shared.MiByte,
			},
		},
	}
}
