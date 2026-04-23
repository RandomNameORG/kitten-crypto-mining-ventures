# Simple Makefile for local dev. `make help` lists everything.

BIN_DIR := bin
LOCAL   := $(BIN_DIR)/meowmine
SSH     := $(BIN_DIR)/meowmine-ssh

.PHONY: help build run ssh test lint vet tidy clean

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"}; /^[a-zA-Z_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

tidy: ## Run go mod tidy (fetches deps, regenerates go.sum)
	go mod tidy

build: ## Build both binaries into bin/
	@mkdir -p $(BIN_DIR)
	go build -o $(LOCAL) ./cmd/meowmine
	go build -o $(SSH)   ./cmd/meowmine-ssh

run: ## Run the local TUI (equivalent to: go run ./cmd/meowmine)
	go run ./cmd/meowmine

ssh: build ## Build and run the SSH server on :23234
	./$(SSH)

test: ## Run all tests
	go test ./...

vet: ## Static analysis
	go vet ./...

lint: vet ## Alias for vet

clean: ## Remove build artefacts
	rm -rf $(BIN_DIR)
