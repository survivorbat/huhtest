MAKEFLAGS := --no-print-directory --silent

default: help

help:
	@echo "Please use 'make <target>' where <target> is one of"
	awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z\._-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format go code, tidy the go.mod file and run the linter
	@go mod tidy
	@gofumpt -l -w .
	@golangci-lint run --fix

tools: ## Install extra tools for development
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: ## Lint the code locally
	golangci-lint run

test: ## Run unit-tests, but not the integration tests defined in integration_test.go
	go test ./... -short -timeout=10s -parallel=20

