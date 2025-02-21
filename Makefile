
# Define directories
OUTPUT_DIR := bin
CONSENSUS_PROTOBUF_DIR := pkg/consensus

# Define build flags
GCFLAGS := -gcflags "all=-N -l"

# Define executable names
LB_BINARY_NAME := gale-lb
NODE_BINARY_NAME := gale-node

# Define the source files for executables
LB_SOURCE := $(wildcard cmd/lb/*.go)
NODE_SOURCE := $(wildcard cmd/node/*.go)

# Define path files for protoc compiler
CONSENSUS_PROTOBUF_SOURCE := $(wildcard $(CONSENSUS_PROTOBUF_DIR)/*.proto)

.PHONY: lb
lb: protobuf
	@echo "building lb binary"
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(LB_BINARY_NAME) $(LB_SOURCE)
	@echo "lb binary ready at: $(OUTPUT_DIR)/$(LB_BINARY_NAME)"

.PHONY: node
node: protobuf
	@echo "building node binary"
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NODE_BINARY_NAME) $(NODE_SOURCE)
	@echo "node binary ready at: $(OUTPUT_DIR)/$(NODE_BINARY_NAME)"

.PHONY: protobuf
protobuf:
	@echo "generating protobuf code"
	@protoc --go_out=. --go_opt=paths=source_relative $(CONSENSUS_PROTOBUF_SOURCE)

.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

.PHONY: imports
imports:
	@find . -name "*.go" | xargs goimports -w

.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)/*
