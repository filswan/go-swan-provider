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

PKG := "$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: all ffi build clean test help

all: build

test: ## Run unittests
	@go test -short ${PKG_LIST}
	@echo "Done testing."

ffi:
	./extern/filecoin-ffi/install-filcrypto
.PHONY: ffi

build: ## Build the binary file
	@go mod download
	@go mod tidy
	@go build -o $(GOBIN)/$(BINARY_NAME)  main.go
	@echo "Done building."
	@echo "Go to build folder and run \"$(GOBIN)/$(BINARY_NAME)\" to launch swan provider."

install-provider:
	sudo install -C $(GOBIN)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

clean: ## Remove previous build
	@go clean
	@rm -rf $(shell pwd)/build
	@echo "Done cleaning."

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME) -v  main.go
build_win: test
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME) -v  main.go

build_boost:
	git clone https://github.com/filecoin-project/boost
	cd boost && git checkout v1.5.0
	cd boost && make build && sudo mv boostd /usr/local/bin/
	rm -rf boost
.PHONY: build_boost