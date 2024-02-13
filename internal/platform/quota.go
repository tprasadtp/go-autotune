// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package platform

// GetMemoryQuota returns memory quotas in bytes for the current process.
//   - For Linux this reads information cgroups v2 interface.
//   - For Windows this uses [QueryInformationJobObject] API.
//   - For other platforms this always returns nil, [errors.ErrUnsupported].
//
// [QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/desktop/api/jobapi2/nf-jobapi2-queryinformationjobobject
//
//nolint:nonamedreturns // for docs.
func GetMemoryQuota(options ...Option) (max, high int64, err error) {
	return getMemoryQuota(options...)
}

// GetCPUQuota returns CPU quotas set on the current process.
//   - For Linux this reads information cgroups v2 interface.
//   - For Windows this uses [QueryInformationJobObject] API.
//   - For other platforms this always returns nil, [errors.ErrUnsupported].
//
// [QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/desktop/api/jobapi2/nf-jobapi2-queryinformationjobobject
func GetCPUQuota(options ...Option) (float64, error) {
	return getCPUQuota(options...)
}
