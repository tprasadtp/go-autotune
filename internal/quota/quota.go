// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

// Package quota implements platform specific CPU and memory quota detectors.
//
// For Linux this reads information cgroups v2 interface.
// For Windows this uses [QueryInformationJobObject] API.
//
// This is an internal API and is not covered by compatibility
// guarantees. Select platform from drop down menu on pkg.go.dev
// as some API methods and types are platform specific.
//
// [QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/desktop/api/jobapi2/nf-jobapi2-queryinformationjobobject
package quota
