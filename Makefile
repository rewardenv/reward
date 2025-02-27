.DEFAULT_GOAL = help

SHELL         = bash
project       = reward
GIT_AUTHOR    = janosmiko
GO_DOCKER           = docker run --rm -v $(PWD):/app -w /app golang:1.24.0
GO					= $(GO_DOCKER) go
GOLANGCI_LINT		= docker run --rm -v $(PWD):/app -v $(HOME)/Library/Caches/golangci-lint:/tmp/golangci-lint -e GOLANGCI_LINT_CACHE=/tmp/golangci-lint -w /app golangci/golangci-lint:v1.64.4 golangci-lint
GORELEASER          = docker run --rm -v $(PWD):/app -w /app goreleaser/goreleaser:v2.5.1
BASHUNIT            = docker run --rm -v $(PWD):/app -w /app rewardenv/docker-toolbox "curl -s https://bashunit.typeddevs.com/install.sh | bash -s -- /usr/local/bin && find images -name "*_test.sh" -type f -print0 | xargs -0 -t bashunit"

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
gomod: ## Download Go dependencies
	$(GO) mod tidy

goupdate: ## Update Go dependencies
	$(GO) get -u ./...
	$(GO) mod tidy

lint: ## Lint Go code
	$(GOLANGCI_LINT) run ./...

lint-fix: ## Lint Go code
	$(GOLANGCI_LINT) run --fix ./...

test-go: ## Run Go tests
	$(GO) test -v -race ./...

test-bash: ## Run Bash tests
	$(BASHUNIT)

test: ## Run Go tests
	$(MAKE) test-go
	$(MAKE) test-bash
