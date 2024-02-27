# Example

> [!IMPORTANT]
>
> This is an _example_ and is **NOT** covered by semver compatibility guarantees.

## Windows Containers

> [!IMPORTANT]
>
> This example uses `mcr.microsoft.com/windows/nanoserver:2004` as the base image,
> change it to suit your base OS. See [Windows container version compatibility]
> for more info.

- Change base directory

  ```powershell
  cd example
  ```

- Build example go binary

  ```console
  go build -o example.exe example.go
  ```

- Run container with process isolation

  ```powershell
  docker run --isolation=process --rm --user=ContainerAdministrator --memory=100M --cpus=2 -v $PWD\:C:\app:ro mcr.microsoft.com/windows/nanoserver:2004 C:\app\example.exe
  ```

  ```console
  GOOS       : windows
  GOMAXPROCS : 2
  NumCPU     : 4
  GOMEMLIMIT : 47185920
  ```

- Run container with Hyper-V isolation

  ```powershell
  docker run --isolation=hyperv --rm --user=ContainerAdministrator --memory=250M --cpus=2 -v $PWD\:C:\app:ro mcr.microsoft.com/windows/nanoserver:2004 C:\app\example.exe
  ```

  ```console
  GOOS       : windows
  GOMAXPROCS : 2
  NumCPU     : 2
  GOMEMLIMIT : 235929600
  ```

## Linux Systemd Services

- Change base directory

  ```console
  cd example
  ```

- Build example go binary

  ```bash
  go build -o example example.go
  ```

- Verify CPU and Memory controller delegation is available for user level units.

  ```bash
  systemctl show user@$(id -u).service --property=DelegateControllers
  ```

  should show something like below. It **must contain** both `cpu` and `memory`.

  ```console
  DelegateControllers=cpu memory pids
  ```

- Run with with resource limits

  ```bash
  systemd-run \
    --user \
    --pipe \
    --quiet \
    --property="CPUQuota=150%" \
    --property=MemoryHigh=250M \
    --property=MemoryMax=300M \
    example
  ```

  ```
  GOOS       : linux
  GOMAXPROCS : 2
  NumCPU     : 4
  GOMEMLIMIT : 262144000
  ```

[Windows container version compatibility]: https://learn.microsoft.com/en-us/virtualization/windowscontainers/deploy-containers/version-compatibility
