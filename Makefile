# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT


# Variables
BINARY_NAME=smd2
BINARY_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION?=0.0.1
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

CONTAINER_CMD=docker

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: clean build

# Build the binary
.PHONY: build
build: $(BINARY_DIR)/$(BINARY_NAME)

$(BINARY_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "$(GO_FILES)"
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	go build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/server/

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BINARY_DIR)
	rm -rf data
	rm -f coverage.out coverage.html

# Clean build artifacts
.PHONY: image
image: build
	@echo "Building container image..."
	$(CONTAINER_CMD) build -t $(BINARY_NAME):latest .


# Clean build artifacts
.PHONY: image-alpine
image-alpine:
	@echo "Building container alpine image..."
	$(CONTAINER_CMD) build -f Dockerfile-alpine -t $(BINARY_NAME):latest .
