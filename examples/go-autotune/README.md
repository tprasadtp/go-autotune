# Example

> [!IMPORTANT]
>
> This is an _example_ and is **NOT** covered by semver compatibility guarantees.

If `PORT` env variable is specified and is a valid port, a simple http server
is started on that port, listening on all available interfaces. Alternatively,
`-listen` flag can be used to specify listening address. If both are not specified,
then container simply prints `GOMAXPROCS` and `GOMEMLIMIT` values and some runtime/platform
data to stdout and exits.

## Docker

Example docker images are only provided for limited number of platforms/architectures.

As examples are not covered by semver compatibility guarantees, semver tagged images are
not provided. However, images are tagged with both short and full commit hashes to test
a specific commit. `latest` tag corresponds to `HEAD` of the default branch.

[SLSA build level 3][slsa-build-l3] provenance is attached to the images. This project
_also_ provides [GitHub native provenance](https://github.com/tprasadtp/go-autotune/attestations)
for images, though only at [SLSA build level 2][slsa-build-l2], [due to isolation requirements][slsa-levels-github-native].

<div align="center">

[![slsa-level3-badge](../logos/slsa-level3-logo.svg)][slsa-build-l3]

</div>


```console
docker run --rm --cpus=1.5 --memory=250M ghcr.io/tprasadtp/go-autotune
```

### Docker (Linux)

![linux-stdout](../screenshots/linux-docker.svg)

### Docker (Windows)

> [!IMPORTANT]
>
> _Example docker images_ are only provided for Server 2019, Server 2022 and
> Server 2025 because of [Windows container version compatibility].

![windows-stdout](../screenshots/windows-docker.svg)

![windows-server](../screenshots/windows-http-server.png)

## Systemd

- Install the example binary.

  ```bash
  go install github.com/tprasadtp/go-autotune/example@latest
  ```

- Verify that CPU and memory controllers are available for user level units.
  If output does not contain strings `cpu` and `memory`, CPU controllers are not available for
  user level units. Install the binary to a root accessible location (like `/usr/local/bin`)
  by setting `GOBIN` environment and run the `systemd-run` commands without the `--user` flag.

  ```bash
  systemctl show user@$(id -u).service -P DelegateControllers
  ```

- Run the example binary as a transient unit with with resource limits applied.

  ```bash
  systemd-run -Pq --user -p "CPUQuota=150%" -p MemoryHigh=250M -p MemoryMax=300M go-autotune
  ```

  ![linux-systemd](../screenshots/linux-systemd-run.svg)

[Windows container version compatibility]: https://learn.microsoft.com/en-us/virtualization/windowscontainers/deploy-containers/version-compatibility
[slsa-build-l3]: https://slsa.dev/spec/v1.0/levels#build-l3
[slsa-build-l2]: https://slsa.dev/spec/v1.0/levels#build-l2
[slsa-levels-github-native]: https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds#about-slsa-levels-for-artifact-attestations
