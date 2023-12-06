# go-autotune

[![Go Reference](https://pkg.go.dev/badge/github.com/tprasadtp/go-autotune.svg)](https://pkg.go.dev/github.com/tprasadtp/go-autotune)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tprasadtp/go-autotune?label=go&logo=go&logoColor=white)
[![test](https://github.com/tprasadtp/go-autotune/actions/workflows/test.yml/badge.svg)](https://github.com/tprasadtp/go-autotune/actions/workflows/test.yml)
[![GitHub](https://img.shields.io/github/license/tprasadtp/go-autotune)](https://github.com/tprasadtp/go-autotune/blob/master/LICENSE)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/tprasadtp/go-autotune?color=7f50a6&label=release&logo=semver&sort=semver)](https://github.com/tprasadtp/go-autotune/releases)

Automatically configure [`GOMAXPROCS`][GOMAXPROCS] and [`GOMEMLIMIT`][GOMEMLIMIT] for your
applications to match to match Container/Cgroup CPU quota and memory limits.

## Usage

See [API docs](https://pkg.go.dev/github.com/tprasadtp/go-autotune) for more info and examples.

```go
package main

import (
	_ "github.com/tprasadtp/go-autotune" // Importing this adjusts GOMAXPROCS and GOMEMLIMIT
)
```

## Testing

Testing requires Linux system with cgroups v2 support enabled.

```
go test -cover ./...
```

## Linux Version

Minimum required linux version is 5.8 or later, as this module only supports Cgroups V2.
Also Cgroupv2 **MUST** be enabled. Following Linux distributions enable Cgroups v2 support
by default.

- Container Optimized OS (since M97)
- Ubuntu (since 21.10)
- Debian GNU/Linux (since Debian 11 Bullseye)
- Fedora (since 31)
- Arch Linux (since April 2021)
- RHEL and RHEL-like distributions (since 9)

## Container Runtime

- Kubernetes 1.25 or later
- containerd v1.4 or later
- cri-o v1.20 or later
- ECS agent [1.61.0](https://github.com/aws/amazon-ecs-agent/pull/3127) or later

## Disabling Automatic Configuration

To disable automatic configuration at runtime(for compiled binaries),
Set `GO_AUTOTUNE` environment variable to `false`.

## Incompatible Modules

This module is incompatible with other modules which also tweak [`GOMAXPROCS`][GOMAXPROCS]
and [`GOMEMLIMIT`][GOMEMLIMIT]. Following [golangci-lint] snippet might help avoid any
issues.

```yml
linters-settings:
  # <snip other linter settings>
  gomodguard:
    blocked:
      - go.uber.org/automaxprocs:
          reason: "Does not handle fractional CPUs well and does not configure GOMEMLIMIT."
          recommendations:
            - "github.com/tprasadtp/go-autotune"
        - github.com/KimMachineGun/automemlimit:
            reason: "Does not support cgroups mounted at non standard location and does not configure GOMAXPROCS."
            recommendations:
              - "github.com/tprasadtp/go-autotune"
linters:
  enabled:
    # <snip other enabled linters>
    - gomodguard # allow and block lists linter for direct Go module dependencies.
```

[GOMEMLIMIT]: https://pkg.go.dev/runtime/debug#SetMemoryLimit
[GOMAXPROCS]: https://pkg.go.dev/runtime#GOMAXPROCS
[golangci-lint]: https://golangci-lint.run/
