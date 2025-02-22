# GaleLB: multi-node L4 load balancer

## Requirements 

## Architecture

## Configuration
Load balancer configuration:
```toml
node_health_checks_before_routing = 3
node_health_checks_timeout = "5s"
```

Node configuration:
```toml

```

## Example

To run the Gale load balancer, make sure you have the necessary permissions. Elevated privileges are required to adjust the `rlimit` and load the `XDP` module. You can run the load balancer with the following command:

```bash
$ sudo ./bin/gale-lb --config cmd/lb.toml
```


## Dependencies 
Install dependencies for building eBPF programs: 
```bash
$ sudo apt-get install clang llvm libbpf-dev gcc make
$ sudo apt install linux-headers-$(uname -r)
```

Install linter: 
```bash
$ go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5
```

Install `protobuf` and `protoc`: 
```bash 
$ go get google.golang.org/protobuf/cmd/protoc-gen-go
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```
