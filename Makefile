PROJECT_NAME=swan-provider
PKG := "$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: all dep build clean test coverage coverhtml lint

all: build

test: ## Run unittests
	@go test -short ${PKG_LIST}
	@echo "Done testing."

dep: ## Get all the dependencies
	@@go get -u -v all
	@echo "Done getting dependencies."

build: ## Build the binary file
	@go mod download
	@go mod tidy
	@mkdir -p ./build/config
	@cp ./config/config.toml.example ./build/config/config.toml
	@go build -o ./build
	@echo "Done building."

clean: ## Remove previous build
	@go clean
	@rm -rf $(shell pwd)/build
	@echo "Done cleaning."

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'