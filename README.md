# go-autotune

[![go-reference](https://img.shields.io/badge/go-reference-00758D?logo=go&logoColor=white)](https://pkg.go.dev/github.com/tprasadtp/go-autotune)
[![test](https://github.com/tprasadtp/go-autotune/actions/workflows/test.yml/badge.svg)](https://github.com/tprasadtp/go-autotune/actions/workflows/test.yml)
[![lint](https://github.com/tprasadtp/go-autotune/actions/workflows/lint.yml/badge.svg)](https://github.com/tprasadtp/go-autotune/actions/workflows/lint.yml)
[![license](https://img.shields.io/github/license/tprasadtp/go-autotune)](https://github.com/tprasadtp/go-autotune/blob/master/LICENSE)
[![latest-version](https://img.shields.io/github/v/tag/tprasadtp/go-autotune?color=7f50a6&label=release&logo=semver&sort=semver)](https://github.com/tprasadtp/go-autotune/releases)

Automatically configure [`GOMAXPROCS`][GOMAXPROCS] and [`GOMEMLIMIT`][GOMEMLIMIT]
for your applications to match CPU quota and memory limits assigned.
Supports _both_ Windows and Linux.

## How

- For Linux CPU and memory limits are obtained from cgroup v2 interface files.
- For Windows, [QueryInformationJobObject] API is used.

## Usage

See [API docs](https://pkg.go.dev/github.com/tprasadtp/go-autotune) for more info and examples.

```go
package main

import (
	_ "github.com/tprasadtp/go-autotune" // Automatically adjusts GOMAXPROCS & GOMEMLIMIT
)
```

## Requirements (Linux)

This module only supports cgroups V2. Following Linux distributions enable it by default.

- Container Optimized OS (since M97)
- Ubuntu (since 21.10)
- Debian (since Debian 11 Bullseye)
- Fedora (since 31)
- Arch Linux (since April 2021)
- RHEL and RHEL-like distributions (since 9)
- Kubernetes 1.25 or later
- containerd v1.4 or later
- cri-o v1.20 or later

For [user level units](https://wiki.archlinux.org/title/systemd/User),
cpu delegation is enabled by default for systemd [252 or later][b8df7f8].
For older versions, It needs to be enabled [manually](https://github.com/systemd/systemd/issues/12362#issuecomment-485762928).
This also affects rootless docker and podman.

As most production workloads use
kubernetes or system level units, this is not an issue in most cases. If you are
running rootless podman/docker and require CPUQuota to be applied to your workloads,
upgrade to a distribution which uses systemd 252 or later or manually delegate
cpu controller to systemd user instance.

## Requirements (Windows)

- Windows 10 20H2 or later
- Windows Server 2019 or later

> [!IMPORTANT]
>
> For Windows containers, only `process` isolation is fully supported.

## Disabling Automatic Configuration

To disable automatic configuration at runtime (for compiled binaries),
Set `GO_AUTOTUNE` environment variable to `false` or `0`.

## Incompatible Modules

This module is incompatible with other modules which also tweak [`GOMAXPROCS`][GOMAXPROCS]
and [`GOMEMLIMIT`][GOMEMLIMIT]. Following [golangci-lint] snippet might help avoid any
issues.

```yml
linters-settings:
  # <snip other linter settings>
  gomodguard:
    blocked:
      modules:
        # <snip other modules>
        - go.uber.org/automaxprocs:
            reason: "Does not handle fractional CPUs well and does not support Windows."
            recommendations:
              - "github.com/tprasadtp/go-autotune"
        - github.com/KimMachineGun/automemlimit:
            reason: |
              Does not support cgroups mounted at non standard location,
              does not support memory.high and does not support Windows.
            recommendations:
              - "github.com/tprasadtp/go-autotune"
linters:
  enabled:
    # <snip other enabled linters>
    - gomodguard
```

## Testing

Testing on Linux requires cgroups v2 support enabled and systemd 252 or later.
Testing on Windows requires Windows 10 20H2/Windows Server 2019 or later.

```console
go test -cover ./...
```

> [!IMPORTANT]
>
> Running _unit tests_ within containers/tasks with already applied resource limits
> is _not supported_.

[GOMEMLIMIT]: https://pkg.go.dev/runtime/debug#SetMemoryLimit
[GOMAXPROCS]: https://pkg.go.dev/runtime#GOMAXPROCS
[golangci-lint]: https://golangci-lint.run/
[b8df7f8]: https://github.com/systemd/systemd/pull/23887
[QueryInformationJobObject]: https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-queryinformationjobobject
