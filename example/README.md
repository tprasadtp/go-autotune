# Example

## Windows Containers

- Build go binary

  ```console
  go build example/example.go -o example/example.exe
  ```

- Run with docker process isolation

  ```powershell
  docker run --isolation=process --rm --env=GO_AUTOTUNE=debug -it --user=ContainerAdministrator --memory=100M --cpus=0.5 -v $PWD\example:C:\Shared mcr.microsoft.com/windows/nanoserver:2004 C:\shared\example.exe
  ```

  ``````console
  time=2023-12-13T01:40:25.022+01:00 level=INFO msg="Successfully obtained CPU Quota" CPUQuota=2
  time=2023-12-13T01:40:25.026+01:00 level=INFO msg="Setting GOMAXPROCS" GOMAXPROCS=2
  time=2023-12-13T01:40:25.027+01:00 level=INFO msg="Successfully obtained memory limits" memory.max=104857600 memory.high
  =0
  time=2023-12-13T01:40:25.030+01:00 level=INFO msg="Using default reserve percent value" ReservePercent=10
  time=2023-12-13T01:40:25.033+01:00 level=ERROR msg="Max allowed memory (with reserve)" memeory.max=94371840
  time=2023-12-13T01:40:25.033+01:00 level=INFO msg="Only memory.max (with reserve) is specififed" GOMEMLIMIT=94371840
  time=2023-12-13T01:40:25.034+01:00 level=INFO msg="Setting GOMEMLIMIT" GOMEMLIMIT=94371840
  Env (GOMAXPROCS)  :
  Env (GOMEMLIMIT)  :
  Env (GO_AUTOTUNE) : debug

  GOMAXPROCS        : 2
  Memory Limit      : 94371840
  ```
