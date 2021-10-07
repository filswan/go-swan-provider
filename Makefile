# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOBIN=$(shell pwd)/build

# Binary names
PROJECT_NAME=swan-provider
BINARY_NAME=$(PROJECT_NAME)
BINARY_UNIX=$(BINARY_NAME)_unix

PKG := "$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: all build clean test help

all: build

test: ## Run unittests
	@go test -short ${PKG_LIST}
	@echo "Done testing."

build: ## Build the binary file
	@go mod download
	@go mod tidy
	@mkdir -p ./build/config
	@go build -o $(GOBIN)/$(BINARY_NAME)  main.go
	@echo "Done building."
	@echo "Go to build folder and run \"$(GOBIN)/$(BINARY_NAME)\" to launch swan provider."

clean: ## Remove previous build
	@go clean
	@rm -rf $(shell pwd)/build
	@echo "Done cleaning."

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_UNIX) -v  main.go
build_win: test
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_UNIX) -v  main.go
