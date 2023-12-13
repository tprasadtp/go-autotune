# go-autotune

[![go-reference](https://img.shields.io/badge/go-reference-00758D?logo=go&logoColor=white)](https://pkg.go.dev/github.com/tprasadtp/go-autotune)
[![test](https://github.com/tprasadtp/go-launchd/actions/workflows/test.yml/badge.svg)](https://github.com/tprasadtp/go-autotune/actions/workflows/test.yml)
[![lint](https://github.com/tprasadtp/go-launchd/actions/workflows/lint.yml/badge.svg)](https://github.com/tprasadtp/go-autotune/actions/workflows/lint.yml)
[![license](https://img.shields.io/github/license/tprasadtp/go-launchd)](https://github.com/tprasadtp/go-autotune/blob/master/LICENSE)
[![latest-version](https://img.shields.io/github/v/tag/tprasadtp/go-launchd?color=7f50a6&label=release&logo=semver&sort=semver)](https://github.com/tprasadtp/go-autotune/releases)

Automatically configure [`GOMAXPROCS`][GOMAXPROCS] and [`GOMEMLIMIT`][GOMEMLIMIT]
for your applications to match CPU quota and memory limits assigned.
Supports _both_ Windows and Linux.

## Usage

See [API docs](https://pkg.go.dev/github.com/tprasadtp/go-autotune) for more info and examples.

```go
package main

import (
	_ "github.com/tprasadtp/go-autotune" // Automatically adjusts GOMAXPROCS & GOMEMLIMIT
)
```

## Testing

Testing on Linux requires cgroups v2 support enabled and systemd 249 or later.
Testing on Windows requires Windows 10 or later. Running unit tests on a system
with single CPU core _might error_.

- Create `.gocover` directory to gather coverage data

    ```bash
    mkdir .gocover
    ```

- Run Tests

    ```console
    go test -cover --test.gocoverdir .gocover ./...
    ```

## Requirements

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

For Windows 10 20H2 or later is supported. For Windows Server only 2019 or later is supported.
For Windows containers, only `--isloation=process` is fully supported.

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

[GOMEMLIMIT]: https://pkg.go.dev/runtime/debug#SetMemoryLimit
[GOMAXPROCS]: https://pkg.go.dev/runtime#GOMAXPROCS
[golangci-lint]: https://golangci-lint.run/
