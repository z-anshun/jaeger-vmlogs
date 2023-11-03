# Jaeger-VictoriaLogs

Store traces in VictoriaLogs for the Jaeger gRPC implementation.

# Why use VictoriaLogs for Jaeger?

Although VictoriaLogs has just come out, I think it is a good product with good prospects and good performance in storage and query, so I try to use it as a link storage here.

# Quick start

Based on the simplicity and convenience of VictoriaLogs, we can start it quickly.

## Build & Run
```shell
export GOOS=linux GOARCH=amd64 
make build
make run
```
Open [localhost:16686](http://localhost:16686/)
