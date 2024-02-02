# Example

## Windows Containers

- Build example go binary

  ```console
  go build -o .\build\example.exe .\example
  ```

- Run container with process isolation

  ```powershell
  docker run --isolation=process --rm --user=ContainerAdministrator --memory=100M --cpus=0.5 -v $PWD\build:C:\app mcr.microsoft.com/windows/nanoserver:2004 C:\app\example.exe
  ```

  ```console
  GOOS       : windows
  GOMAXPROCS : 1
  NumCPU     : 4
  GOMEMLIMIT : 94371840
  ```

## Linux Systemd Services

- Build example go binary

  ```bash
  go build -o build/example ./example/
  ```

- Verify CPU and Memory controller delegation is available for user level units.
  Output from below command **must contain** both `cpu` and `memory`.

  ```bash
  systemctl show user@$(id -u).service --property=DelegateControllers
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
    build/example
  ```

  ```
  GOOS       : linux
  GOMAXPROCS : 2
  NumCPU     : 4
  GOMEMLIMIT : 262144000
  ```
