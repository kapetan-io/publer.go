.DEFAULT_GOAL := test
LINT = $(GOPATH)/bin/golangci-lint
LINT_VERSION = v1.61.0
VERSION=$(shell git describe --tags --exact-match 2>/dev/null || echo "dev-build")

$(LINT): ## Download Go linter
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(LINT_VERSION)

.PHONY: test
test:
	go test -timeout 10m -v -p=1 -count=1 -race ./...

.PHONY: cover
cover:
	-rm coverage.html coverage.out
	go test -timeout 10m -v -p=1 -count=1 --coverprofile=coverage.out -covermode=atomic -coverpkg=./... ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html


.PHONY: lint
lint: ## Run Go linter
	@which golangci-lint > /dev/null || $(MAKE) $(LINT)
	golangci-lint run -v ./...

.PHONY: tidy
tidy:
	go mod tidy && git diff --exit-code

.PHONY: ci
ci: tidy lint test
	@echo
	@echo "\033[32mEVERYTHING PASSED!\033[0m"

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build: ## Build the project
	go build ./...

