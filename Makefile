# Simple Makefile for local dev. `make help` lists everything.

BIN_DIR := bin
LOCAL   := $(BIN_DIR)/meowmine
SSH     := $(BIN_DIR)/meowmine-ssh
SIM     := $(BIN_DIR)/meowmine-sim
WEB     := $(BIN_DIR)/meowmine-web

.PHONY: help build build-sim build-web run run-web run-sim run-debug ssh test lint vet tidy clean

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"}; /^[a-zA-Z_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

tidy: ## Run go mod tidy (fetches deps, regenerates go.sum)
	go mod tidy

build: ## Build all local binaries into bin/
	@mkdir -p $(BIN_DIR)
	go build -o $(LOCAL) ./packages/cli/cmd/meowmine
	go build -o $(SSH)   ./packages/cli/cmd/meowmine-ssh
	go build -o $(SIM)   ./packages/cli/cmd/meowmine-sim
	go build -o $(WEB)   ./packages/web/cmd/meowmine-web

build-sim: ## Build only the headless simulator
	@mkdir -p $(BIN_DIR)
	go build -o $(SIM) ./packages/cli/cmd/meowmine-sim

build-web: ## Build only the 2D web server
	@mkdir -p $(BIN_DIR)
	go build -o $(WEB) ./packages/web/cmd/meowmine-web

run: ## Run the local TUI (equivalent to: go run ./packages/cli/cmd/meowmine)
	go run ./packages/cli/cmd/meowmine

run-debug: ## Run the local TUI with --debug (time multiplier + HUD + cheats)
	go run ./packages/cli/cmd/meowmine --debug

run-web: ## Run the 2D web server on http://localhost:8080
	go run ./packages/web/cmd/meowmine-web

run-sim: ## Run the headless simulator for 1h of virtual time (seed=1)
	go run ./packages/cli/cmd/meowmine-sim --ticks=3600 --seed=1

ssh: build ## Build and run the SSH server on :23234
	./$(SSH)

test: ## Run all tests
	go test ./...

vet: ## Static analysis
	go vet ./...

lint: vet ## Alias for vet

release: ## Cross-compile stripped release binaries for win/linux/macOS x64+arm64
	@mkdir -p $(BIN_DIR)
	@for target in "windows amd64 .exe" "linux amd64 " "darwin amd64 " "darwin arm64 "; do \
		set -- $$target; os=$$1; arch=$$2; ext=$$3; \
		for spec in "meowmine packages/cli/cmd/meowmine" \
		            "meowmine-ssh packages/cli/cmd/meowmine-ssh" \
		            "meowmine-sim packages/cli/cmd/meowmine-sim" \
		            "meowmine-web packages/web/cmd/meowmine-web"; do \
			set -- $$spec; cmd=$$1; pkg=$$2; \
			echo "  building $$cmd-$$os-$$arch$$ext"; \
			GOOS=$$os GOARCH=$$arch go build -ldflags "-s -w" \
				-o $(BIN_DIR)/$$cmd-$$os-$$arch$$ext ./$$pkg; \
		done; \
	done
	@ls -lh $(BIN_DIR)

clean: ## Remove build artefacts
	rm -rf $(BIN_DIR)
