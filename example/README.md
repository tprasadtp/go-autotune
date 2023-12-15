# Example

## Windows Containers

- Build example go binary

  ```console
  go build -o .\build\example.exe .\example
  ```

- Run with container with process isolation

  ```powershell
  docker run --isolation=process --rm --env=GO_AUTOTUNE=debug --user=ContainerAdministrator --memory=100M --cpus=0.5 -v $PWD\build:C:\app mcr.microsoft.com/windows/nanoserver:2004 C:\app\example.exe
  ```

  ```console
  2023/12/15 04:32:37 INFO Successfully obtained cpu quota cpu.quota=0.5
  2023/12/15 04:32:37 INFO Setting GOMAXPROCS GOMAXPROCS=1
  2023/12/15 04:32:37 INFO Successfully obtained memory limits memory.max=104857600 memory.high=0 memory.reserve.bytes=10485760 memory.reserve.percent=10
  2023/12/15 04:32:37 INFO Setting GOMEMLIMIT GOMEMLIMIT=94371840
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

- Run with systemd-run with resource limits

  ```bash
  systemd-run \
    --user \
    --pipe \
    --quiet \
    --setenv=GO_AUTOTUNE=debug \
    --property="CPUQuota=150%" \
    --property=MemoryHigh=250M \
    --property=MemoryMax=300M \
    build/example
  ```

  ```
  2023/12/15 04:46:22 INFO Successfully obtained cpu quota cpu.quota=1.5
  2023/12/15 04:46:22 INFO Setting GOMAXPROCS GOMAXPROCS=2
  2023/12/15 04:46:22 INFO Successfully obtained memory limits memory.max=314572800 memory.high=262144000 memory.reserve.bytes=31457280 memory.reserve.percent=10
  2023/12/15 04:46:22 INFO Setting GOMEMLIMIT GOMEMLIMIT=262144000
  GOOS       : linux
  GOMAXPROCS : 2
  CPUs       : 4
  GOMEMLIMIT : 262144000
  ```
