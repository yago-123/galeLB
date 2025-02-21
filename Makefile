
# Define directories
OUTPUT_DIR := bin

# Define build flags
GCFLAGS := -gcflags "all=-N -l"

# Define executable names
LB_BINARY_NAME := galelb
NODE_BINARY_NAME := gale-node

# Define the source files for executables
LB_SOURCE := $(wildcard cmd/lb/*.go)
NODE_SOURCE := $(wildcard cmd/node/*.go)

.PHONY: lb
lb:
	@echo "building lb binary"
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(LB_BINARY_NAME) $(LB_SOURCE)
	@echo "lb binary ready at: $(OUTPUT_DIR)/$(LB_BINARY_NAME)"

.PHONY: node
node:
	@echo "building node binary"
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NODE_BINARY_NAME) $(NODE_SOURCE)
	@echo "node binary ready at: $(OUTPUT_DIR)/$(NODE_BINARY_NAME)"
