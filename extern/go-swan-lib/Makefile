# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOBIN=$(shell pwd)/build

PROJECT_NAME=swan-lib
BINARY_NAME=$(PROJECT_NAME)
BINARY_UNIX=$(BINARY_NAME)_unix

PKG := "$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: all build clean test help

all: build

test: ## Run unittests
	@go test -short ${PKG_LIST}
	@echo "Done testing."
.PHONY: test

build: ## Build the binary file
	@go mod download
	@go mod tidy
	@go build -o $(GOBIN)/$(BINARY_NAME)  main.go
	@echo "Done building."
	@echo "Go to build folder and run \"$(GOBIN)/$(BINARY_NAME)\" to launch swan client."
.PHONY: build

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_UNIX) -v  main.go
.PHONY: build_linux

build_win: test
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_UNIX) -v  main.go
.PHONY: build_win

clean: ## Remove previous build
	@go clean
	@rm -rf $(shell pwd)/build
	@echo "Done cleaning."
.PHONY: clean

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: clean
