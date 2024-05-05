// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package scenarios provides test scenarios to be tested using trampolines.
//
// This is shared between internal/autotune package and the external autotune
// package which is only used for side effects, to facilitate common tests.
package scenarios

import (
	"github.com/tprasadtp/go-autotune/internal/shared"
	"github.com/tprasadtp/go-autotune/internal/trampoline"
)

// Common scenarios which are not platform specific.
func Common() []trampoline.Scenario {
	return []trampoline.Scenario{
		//---------------------------------------------------
		// GOMAXPROCS
		//---------------------------------------------------
		{
			Name:   "NoLimits",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug"},
			},
		},
		{
			Name:   "Env/GOMAXPROCS=1",
			Verify: VerifyFunc(1, 0),
			Opts: trampoline.Options{
				Env: []string{"GOMAXPROCS=1", "GOAUTOTUNE=debug"},
			},
		},
		{
			Name:   "Env/GOMAXPROCS=2",
			Verify: VerifyFunc(2, 0),
			Opts: trampoline.Options{
				Env: []string{"GOMAXPROCS=2", "GOAUTOTUNE=debug"},
			},
		},
		{
			Name:   "Env/GOMAXPROCS=Negative",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOMAXPROCS=-2", "GOAUTOTUNE=debug"},
			},
		},
		{
			Name:   "Env/GOMAXPROCS=Zero",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOMAXPROCS=0", "GOAUTOTUNE=debug"},
			},
		},
		{
			Name:   "Env/GOMAXPROCS=Fraction",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOMAXPROCS=0.5", "GOAUTOTUNE=debug"},
			},
		},
		{
			Name:   "Env/GOMAXPROCS=NotInteger",
			Verify: VerifyFunc(0, 0),
			Opts: trampoline.Options{
				Env: []string{"GOMAXPROCS=NotInteger", "GOAUTOTUNE=debug"},
			},
		},
		//---------------------------------------------------
		// GOMEMLIMIT
		//---------------------------------------------------
		{
			Name:   "Env/GOMEMLIMIT=235929600",
			Verify: VerifyFunc(0, 225*shared.MiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug", "GOMEMLIMIT=235929600"},
			},
		},
		{
			Name:   "Env/GOMEMLIMIT=225MiB",
			Verify: VerifyFunc(0, 225*shared.MiByte),
			Opts: trampoline.Options{
				Env: []string{"GOAUTOTUNE=debug", "GOMEMLIMIT=225MiB"},
			},
		},
	}
}

func All() []trampoline.Scenario {
	v := Common()
	v = append(v, PlatformSpecific()...)
	return v
}
