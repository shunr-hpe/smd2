# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT


# Variables
BINARY_NAME=inventory
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
	go build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-service ./cmd/server/
	go build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-cli ./cmd/client/

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BINARY_DIR)
	rm -rf data
	rm -rf data-resttests
	rm -f coverage.out coverage.html

# Build container image
.PHONY: image
image: build
	@echo "Building container image..."
	$(CONTAINER_CMD) build -t $(BINARY_NAME)-service:latest .

unittest:
	go test -cover -v ./apis/... ./cmd/... ./internal/... ./pkg/...

resttest:
	go test -cover -v ./resttests/...

# Build test container image
.PHONY: test-image
test-image:
	@echo "Building test container image..."
	$(CONTAINER_CMD) build -t $(BINARY_NAME)-test:latest tests/pytests


# Start compose environment
.PHONY: start-compose
start-compose:
	@echo "Starting docker compose environment for testing...";
	tests/compose/generate-config;
	cd tests/compose && $(CONTAINER_CMD) compose -p inventory -f networks.yml -f postgres.yml -f smd.yml -f inventory-service.yml -f computes.yml up -d;


# Stop compose environment
.PHONY: stop-compose
stop-compose:
	@echo "Stoping docker compose environment for testing..."
	$(CONTAINER_CMD) compose -p inventory down -v


# Restart compose environment
.PHONY: restart-compose
restart-compose: stop-compose start-compose
