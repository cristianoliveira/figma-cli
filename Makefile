# Makefile for figma-cli
# Only targets that work RIGHT NOW

# Variables
BINARY_NAME := figma
BUILD_DIR := bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-X main.version=$(VERSION)"

# Help
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build
.PHONY: build clean run
build: ## Build binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/figma

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

run: ## Run application with go run
	go run ./cmd/figma

# Testing
.PHONY: test test-short test-race test-cover
test: ## Run all tests
	go test ./...

test-cover: ## Run coverage for packages that contain tests
	@go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./... | xargs -r go test -cover

test-short: ## Run short tests only
	go test ./... -short

test-race: ## Run tests with race detection
	go test ./... -race

# Code Quality
.PHONY: fmt vet
fmt: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

# Dependencies
.PHONY: deps tidy
deps: ## Download dependencies
	go mod download

tidy: ## Run go mod tidy
	go mod tidy

# Pre-commit Hooks
.PHONY: install-hooks run-hooks
install-hooks: ## Install pre-commit hooks
	lefthook install

run-hooks: ## Run all pre-commit hooks manually
	lefthook run pre-commit

# Utility
.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"

.DEFAULT_GOAL := help