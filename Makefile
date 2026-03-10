PROJECT ?= xenorchestra-k8s-common

TESTARGS ?= "-v"

############

# Help Menu

define HELP_MENU_HEADER
# Getting Started

To work with this project, you must have the following installed:

- make
- golang 1.24+
- golangci-lint

endef

export HELP_MENU_HEADER

.PHONY: help
help: ## This help menu.
	@echo "$$HELP_MENU_HEADER"
	@grep -E '^[a-zA-Z0-9%_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

############
#
# Code Quality
#

.PHONY: fmt
fmt: ## Format code with gofmt
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Lint code with golangci-lint
	golangci-lint run --config .golangci.yml

.PHONY: tidy
tidy: ## Tidy go modules
	go mod tidy

.PHONY: vuln
vuln: ## Check vulnerabilities with govulncheck
	govulncheck ./...

############
#
# Code Generation
#

.PHONY: mock
mock: ## Generate mocks with go.uber.org/mock
	go generate ./...

############
#
# Testing
#

.PHONY: unit
unit: ## Run unit tests
	go test $(shell go list ./...) $(TESTARGS)

.PHONY: test
test: vet unit ## Run vet + unit tests

############
#
# All-in-one
#

.PHONY: check
check: fmt vet lint unit ## Run all checks (fmt, vet, lint, unit)
