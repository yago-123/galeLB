
# Define directories
OUTPUT_DIR := bin
CONSENSUS_PROTOBUF_DIR := pkg/consensus/v1

# Define eBPF compiler and CFlags
BPF_CLANG = clang
BPF_CFLAGS = -O2 -emit-llvm -c -g -target bpf -I/usr/include/linux -I/usr/include
BPF_APP_IMPORTS = -Ipkg/common
BPF_LLC_FLAGS = -march=bpf -filetype=obj

# Define build flags for Go
GCFLAGS := -gcflags "all=-N -l"

# Define output path for the XDP router program
XDP_OBJECT_OUTPUT := pkg/routing/xdp_obj/xdp_router.o

# Define executable names
LB_BINARY_NAME := gale-lb
NODE_BINARY_NAME := gale-node

# Define the source files for executables
LB_SOURCE := $(wildcard cmd/lb/*.go)
NODE_SOURCE := $(wildcard cmd/node/*.go)

# Define path files for protoc compiler
CONSENSUS_PROTOBUF_SOURCE := $(wildcard $(CONSENSUS_PROTOBUF_DIR)/*.proto)

.PHONY: build 
build: lb node

.PHONY: lb
lb: protobuf xdp_router
	@echo "building lb binary"
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(LB_BINARY_NAME) $(LB_SOURCE)
	@echo "lb binary ready at: $(OUTPUT_DIR)/$(LB_BINARY_NAME)"

.PHONY: xdp_router
xdp_router:
	$(BPF_CLANG) $(BPF_CFLAGS) $(BPF_APP_IMPORTS) pkg/routing/router.c -o - | llc $(BPF_LLC_FLAGS) -o $(XDP_OBJECT_OUTPUT)

.PHONY: node
node: protobuf
	@echo "building node binary"
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NODE_BINARY_NAME) $(NODE_SOURCE)
	@echo "node binary ready at: $(OUTPUT_DIR)/$(NODE_BINARY_NAME)"

.PHONY: protobuf
protobuf:
	@echo "generating protobuf code"
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $(CONSENSUS_PROTOBUF_SOURCE)

.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

.PHONY: imports
imports:
	@find . -name "*.go" | xargs goimports -w

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)/*
