.DEFAULT_GOAL := help

lint: ## Run the linter
	golangci-lint run --out-format=github-actions

test: ## Run tests
test: lint
	@go test -v ./...

build: ## Build for local
	@go build -o deployment/dist/server cmd/server/server.go

build-linux: ## Build for Linux platform
	@CGO_ENABLED=0 GOOS=linux go build -o deployment/dist/server cmd/server/server.go

help: ## List of available commands
	@grep -E '(^[a-zA-Z_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'
