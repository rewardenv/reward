.DEFAULT_GOAL = help

SHELL         = bash
project       = reward
GIT_AUTHOR    = janosmiko
GO_DOCKER           = docker run --rm -v $(PWD):/app -w /app golang:1.23
GO					= $(GO_DOCKER) go
GOLANGCI_LINT		= docker run --rm -v $(PWD):/app -v $(HOME)/Library/Caches/golangci-lint:/tmp/golangci-lint -e GOLANGCI_LINT_CACHE=/tmp/golangci-lint -w /app golangci/golangci-lint:v1.60.1 golangci-lint
GORELEASER          = docker run --rm -v $(PWD):/app -w /app goreleaser/goreleaser:v2.3.0

help: ## Outputs this help screen
	@grep -E '(^[\/a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

# If the first argument is "gen"...
ifeq (gen,$(firstword $(MAKECMDGOALS)))
	# use the rest as arguments for "run"
	GEN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
	# ...and turn them into do-nothing targets
	$(eval $(GEN_ARGS):;@:)
endif

## —— Commands —————————————————————————————————————————————————————————
build: ## Build the command to ./dist
	$(GO_DOCKER) /bin/bash -c '\
	go mod download && \
	go generate ./... && \
	CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/reward ./cmd/reward/main.go'

package: ## Build the binaries and packages using goreleaser (without releasing it)
	$(GORELEASER) --clean --snapshot

build-local: ## Build the binaries only using goreleaser (without releasing it)
	$(GORELEASER) --clean --snapshot --config .local.goreleaser.yml

## —— Go Commands —————————————————————————————————————————————————————————
gomod: ## Update Go Dependencies
	$(GO) mod tidy

lint: ## Lint Go Code
	$(GOLANGCI_LINT) run ./...

lint-fix: ## Lint Go Code
	$(GOLANGCI_LINT) run --fix ./...

test: ## Run Go tests
	$(GO) test -v -race ./...
