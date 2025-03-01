# GaleLB: multi-node L4 load balancer

Supports: 
- [ ] L2-Based Forwarding (Stateless MAC Bridging in `XDP`)
- [ ] L3-Based Forwarding (Stateless IP Routing in `XDP`)
- [ ] L4-Based Forwarding (Stateful `NAT` with Connection Tracking in `XDP + TC`)]

## Requirements 
* Linux Kernel 4.4+
* Clang 18+
* LLVM 18+
* 64 bit architecture and x86 CPU

## Architecture

## Configuration
Load balancer configuration:
```toml
[local]
# port opened to listen for incoming connections from nodes
node_port = 7070
# port opened to listen for incoming connections from clients
clients_port = 8080
# interface used to communicate and re-route network packets to nodes
net_interface_nodes = "wlo1"
# interface used to retrieve and re-route network packets from clients 
net_interface_clients = "wlo1"

[node_health]
# number of continuous health checks that must be passed before being eligible for routing destination
checks_before_routing = 3
# duration of deadline between health checks, after this period, nodes will be removed from the routing ring
checks_timeout = "5s"

# number of times nodes can fail to send health checks before they are blacklisted
# ex: the node will be added and removed 5 times to the routing table before they will start to be completly ignored.
# use -1 if want to disable this option
black_list_after_fails = 5

# duration of the ban
black_list_expiry = "5m"

[load_balancer_quorum]
# enforce_single_configuration is used to enforce that all load balancers must have the same configuration regarding
# nodes (ex: node health timeout). If this option is set to true, the load balancers will have to contain the same
# parameters in order to reach consensus. If a load balancer tries to connect with a different configuration, it will
# be ignored
enforce_single_configuration = false

addresses = [
    { ip = "192.168.1.2", port = 8081 },
    { ip = "192.168.1.3", port = 8081 },
    { ip = "192.168.1.4", port = 8081 }
]
```

Node configuration:
```toml
[load_balancer]
addresses = [
    { ip = "192.168.1.2", port = 8082 },
    { ip = "192.168.1.3", port = 8082 },
    { ip = "192.168.1.4", port = 8082 }
]
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
