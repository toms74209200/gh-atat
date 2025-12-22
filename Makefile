.PHONY: all build

all: check test build

build: ## Build the project
	go build -v ./...

.PHONY: clean

clean: ## Clean build artifacts
	rm -f coverage.txt
	rm -f gh-atat

.PHONY: test test-medium test-large test-cover

test-small: ## Run small size tests
	go test -tags=small -v ./...

test-medium: ## Run medium size tests
	go test -tags=medium -v ./...

test-large: ## Run large size tests
	cd tests/acceptance && go test -tags=large -v ./...

test-all: test-small test-medium test-large ## Run all tests

test-cover: ## Run tests with coverage
	go test -tags=small -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

.PHONY: check lint fmt

check: lint fmt ## Run lint and format checks

lint: ## Run lint
	golangci-lint run

fmt: ## Run go fmt
	go fmt ./...

.PHONY: help

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'