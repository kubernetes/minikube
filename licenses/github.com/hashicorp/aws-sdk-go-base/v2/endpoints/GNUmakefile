TIMEOUT ?= 30s

default: test lint

cleantidy: ## Tidy go modules
	@echo "make: tidying Go mods..."
	@cd tools && go mod tidy && cd ..
	@cd v2/awsv1shim && go mod tidy && cd ../..
	@go mod tidy
	@echo "make: Go mods tidied"

fmt: ## Run gofmt
	gofmt -s -w ./

gen: ## Run generators
	@echo "make: Running Go generators..."
	@go generate ./...

golangci-lint: ## Run golangci-lint
	@golangci-lint run ./...
	@cd v2/awsv1shim && golangci-lint run ./...

help: ## Display this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-13s\033[0m %s\n", $$1, $$2}'

importlint: ## Lint imports
	@impi --local . --scheme stdThirdPartyLocal ./...

lint: golangci-lint importlint ## Run all linters

semgrep: ## Run semgrep checks
	@docker run --rm --volume "${PWD}:/src" returntocorp/semgrep semgrep --config .semgrep --no-rewrite-rule-ids

test: ## Run unit tests
	go test -timeout=$(TIMEOUT) -parallel=4 ./...
	cd v2/awsv1shim && go test -timeout=$(TIMEOUT) -parallel=4 ./...

tools: ## Install tools
	cd tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd tools && go install github.com/pavius/impi/cmd/impi

# Please keep targets in alphabetical order
.PHONY: \
	cleantidy \
	fmt \
	gen \
	golangci-lint \
	help \
	importlint \
	lint \
	test \
	test \
	tools \
