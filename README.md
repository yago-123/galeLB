
![Alt text](https://github.com/user-attachments/assets/4418759f-3ec0-40b4-95cd-cf02c1bc8ea9)

# GaleLB: multi-node fault-tolerant load balancer

Supports: 
- [ ] L2-Based Forwarding (Stateless MAC Bridging in `XDP`)
- [ ] L3-Based Forwarding (Stateless IP Routing in `XDP`)
- [ ] L4-Based Forwarding (Stateful `NAT` with Connection Tracking in `XDP + TC`)]

## Architecture
![Alt text](https://github.com/user-attachments/assets/bdca33a4-c6c6-4564-9ba7-9c61f6a5af71)

## Requirements
System requirements:
* Linux Kernel 4.4+
* Clang 18+
* LLVM 18+
* 64 bit architecture and x86 CPU

## Configuration
Load balancer configuration:
```toml
[local]
# port opened to listen for incoming connections from nodes
node_port = 7070
# port opened to listen for incoming connections from clients
clients_port = 8080
# interface used to communicate and re-route network packets to nodes
net_interface_private = "wlo1"
# interface used to retrieve and re-route network packets from clients 
net_interface_public = "wlo1"

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

# Running e2e tests 
## Requirements
* Vagrant    2.4.3
* Ansible    2.16.3
* VirtualBox 7.1.0

## Setting up the environment
Create a new key pair of SSH keys for the vagrant machines:
```bash
$ ssh-keygen -t ed25519  -f ~/.ssh/vagrant -N ""
```

Make sure that the SSH-agent is running and add the newly created key:
```bash
$ eval "$(ssh-agent -s)"
$ ssh-add ~/.ssh/vagrant
```

Create the vagrant machines with the network card connected to the LAN (for example `eth0` or `wlo1`):
```bash
$ NETWORK_CARD=<net-card> SSH_KEY_PATH=~/.ssh/vagrant.pub vagrant up
```

This will spawn 3 load balancer instances and 3 node instances. All machines install `avahi-daemon` package in order to 
enable `mDNS` service discovery. Once instances are up, provision the load balancer and the nodes:
```bash
$ ansible-playbook -i ansible/e2e-hosts.ini \
                      ansible/playbooks/lb.yml -K
$ ansible-playbook -i ansible/e2e-hosts.ini \
                      ansible/playbooks/node.yml -K
```

You can also run the playbooks for Proxmox hosts:
```bash
$ ansible-playbook -i ansible/proxmox-hosts.ini \
                      ansible/playbooks/lb.yml -K
$ ansible-playbook -i ansible/proxmox-hosts.ini \
                      ansible/playbooks/node.yml -K
```

## Running the e2e tests
Run the e2e tests sequentially:
```bash
$ sudo go test ./e2e -parallel 1
```

Once you've finished testing the load balancer, you can destroy the machines:
```bash
$ vagrant destroy -f
```