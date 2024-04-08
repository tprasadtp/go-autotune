<div align="center">

# go-autotune

[![go-reference](https://img.shields.io/badge/godoc-reference-5272b4?labelColor=3a3a3a&logo=go&logoColor=959da5)](https://pkg.go.dev/github.com/tprasadtp/go-autotune)
[![go-version](https://img.shields.io/github/go-mod/go-version/tprasadtp/go-autotune?labelColor=3a3a3a&color=00758D&label=go&logo=go&logoColor=959da5)](https://github.com/tprasadtp/go-autotune/blob/master/go.mod)
[![license](https://img.shields.io/github/license/tprasadtp/go-autotune?labelColor=3a3a3a&color=00ADD8&logo=github&logoColor=959da5)](https://github.com/tprasadtp/go-autotune/blob/master/LICENSE)
[![build](https://github.com/tprasadtp/go-autotune/actions/workflows/build.yml/badge.svg)](https://github.com/tprasadtp/go-autotune/actions/workflows/build.yml)
[![lint](https://github.com/tprasadtp/go-autotune/actions/workflows/lint.yml/badge.svg)](https://github.com/tprasadtp/go-autotune/actions/workflows/lint.yml)
[![release](https://github.com/tprasadtp/go-autotune/actions/workflows/release.yml/badge.svg)](https://github.com/tprasadtp/go-autotune/actions/workflows/release.yml)
[![version](https://img.shields.io/github/v/tag/tprasadtp/go-autotune?label=version&sort=semver&labelColor=3a3a3a&color=CE3262&logo=semver&logoColor=959da5)](https://github.com/tprasadtp/go-autotune/releases)

</div>

Automatically configure [`GOMAXPROCS`][GOMAXPROCS] and [`GOMEMLIMIT`][GOMEMLIMIT]
for your applications to match CPU quota and memory limits assigned.
Supports _both_ Windows and Linux.

## How

- For Linux, CPU and memory limits are obtained from cgroup v2 interface files.
- For Windows, [Job Objects API] is used.

## Usage

```go
package main

import (
	_ "github.com/tprasadtp/go-autotune" // Automatically adjusts GOMAXPROCS & GOMEMLIMIT
)
```

See [API docs] and [example](./example/README.md) for more info.

## Requirements (Linux)

This module _only supports cgroups V2_. Following Linux distributions enable it by default.

- Container Optimized OS (since M97)
- Ubuntu (since 21.10)
- Debian (since Debian 11 Bullseye)
- Fedora (since 31)
- Arch Linux (since April 2021)
- RHEL and RHEL-like distributions (since 9)
- Kubernetes 1.25 or later
- containerd v1.4 or later
- cri-o v1.20 or later

For systemd [user level units](https://wiki.archlinux.org/title/systemd/User),
CPU delegation is enabled by default for systemd [252 or later][b8df7f8].
For older versions, It needs to be enabled [manually][enable-cpu-delegation].
This also affects rootless docker and podman.

## Requirements (Windows)

- Windows 10 or later
- Windows Server 2019 or later.

## Disabling Automatic Configuration

To disable automatic configuration at runtime (for compiled binaries),
Set `GOAUTOTUNE` environment variable to `0` or `false`.

## Supporting Kubernetes In-place Resource Resize

This can be done using [time.Ticker](https://pkg.go.dev/time#Ticker)
and a background goroutine. See [API docs] for examples.

See [Kubernetes docs][k8s-resize-docs] for more info.

## Incompatible Modules

This module is incompatible with other modules which _also_ tweak [`GOMAXPROCS`][GOMAXPROCS]
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
            reason: >-
              Does not handle fractional CPUs well and does not support Windows.
            recommendations:
              - github.com/tprasadtp/go-autotune
        - github.com/KimMachineGun/automemlimit:
            reason: >-
              Does not support cgroups mounted at non standard location.
              Also does not support memory.high and does not support Windows.
            recommendations:
              - github.com/tprasadtp/go-autotune
linters:
  enabled:
    # <snip other enabled linters>
    - gomodguard
```

## Example Docker Image

Example docker images are only provided for limited number of platforms/architectures.
However the library will work on all platforms which meet the requirements specified above,
even when running outside of containers. See [example](./example/README.md) for more info.

```console
docker run --rm --cpus=1.5 --memory=250M ghcr.io/tprasadtp/go-autotune
```

### Windows

![windows-docker](./example/screenshots/windows-docker.svg)

### Linux

![linux-docker](./example/screenshots/linux-docker.svg)

## Testing

Testing on Linux requires cgroups v2 support enabled and systemd 252 or later.
Testing on Windows requires Windows 10 20H2/Windows Server 2019 or later.

```console
go test -cover -v ./...
```

> [!IMPORTANT]
>
> Tests extensively use [systemd-run] and [Job Objects API] on Linux and Windows
> respectively. Thus, running unit tests/integration tests within containers is
> _not supported_.


[GOMEMLIMIT]: https://pkg.go.dev/runtime/debug#SetMemoryLimit
[GOMAXPROCS]: https://pkg.go.dev/runtime#GOMAXPROCS
[golangci-lint]: https://golangci-lint.run/
[b8df7f8]: https://github.com/systemd/systemd/pull/23887
[systemd-run]: https://www.freedesktop.org/software/systemd/man/latest/systemd-run.html
[Job Objects API]: https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects
[enable-cpu-delegation]: https://github.com/systemd/systemd/issues/12362#issuecomment-485762928
[pkg-autotune]: https://https://pkg.go.dev/github.com/tprasadtp/go-autotune
[pkg-maxprocs]: https://https://pkg.go.dev/github.com/tprasadtp/go-autotune/maxprocs
[pkg-memlimit]: https://https://pkg.go.dev/github.com/tprasadtp/go-autotune/memlimit
[API docs]: https://pkg.go.dev/github.com/tprasadtp/go-autotune
[k8s-resize-docs]: https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/
[Windows container version compatibility]: https://learn.microsoft.com/en-us/virtualization/windowscontainers/deploy-containers/version-compatibility
