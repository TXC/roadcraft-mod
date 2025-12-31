.PHONY: all build test clean install lint lint-fix fmt help

# Variables
MAIN_NAME=main.go
BINARY_NAME=roadcraft-mod
GO=go
GOFLAGS=-v
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

pkgs          = ./...

FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOHOSTOS     ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH   ?= $(shell $(GO) env GOHOSTARCH)

SKIP_GOLANGCI_LINT :=
GOLANGCI_LINT :=
GOLANGCI_LINT_OPTS ?=
GOLANGCI_LINT_VERSION ?= v2.6.1
# golangci-lint only supports linux, darwin and windows platforms on i386/amd64/arm64.
# windows isn't included here because of the path separator being different.
ifeq ($(GOHOSTOS),$(filter $(GOHOSTOS),linux darwin))
	ifeq ($(GOHOSTARCH),$(filter $(GOHOSTARCH),amd64 i386 arm64))
		# If we're in CI and there is an Actions file, that means the linter
		# is being run in Actions, so we don't need to run it here.
		ifneq (,$(SKIP_GOLANGCI_LINT))
			GOLANGCI_LINT :=
		else ifeq (,$(CIRCLE_JOB))
			GOLANGCI_LINT := $(FIRST_GOPATH)/bin/golangci-lint
		else ifeq (,$(wildcard .github/workflows/golangci-lint.yml))
			GOLANGCI_LINT := $(FIRST_GOPATH)/bin/golangci-lint
		endif
	endif
endif

all: test build

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/$(MAIN_NAME)
	@echo "Build complete!"

## install: Install the binaries to GOPATH/bin
install:
	@echo "Installing..."
	$(GO) install $(LDFLAGS) ./cmd/$(MAIN_NAME)
## test: Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -cover $(pkgs)

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out $(pkgs)
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linters
lint: $(GOLANGCI_LINT)
ifdef GOLANGCI_LINT
	@echo "Running golangci-lint..."
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_OPTS) $(pkgs)
endif

lint-fix: $(GOLANGCI_LINT)
ifdef GOLANGCI_LINT
	@echo "Running golangci-lint fix"
	$(GOLANGCI_LINT) run --fix $(GOLANGCI_LINT_OPTS) $(pkgs)
endif

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt $(pkgs)
	gofmt -s -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet $(pkgs)

## tidy: Tidy go modules
tidy:
	@echo "Tidying go modules..."
	$(GO) mod tidy

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	$(GO) clean

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

ifdef GOLANGCI_LINT
$(GOLANGCI_LINT):
	mkdir -p $(FIRST_GOPATH)/bin
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh \
		| sed -e '/install -d/d' \
		| sh -s -- -b $(FIRST_GOPATH)/bin $(GOLANGCI_LINT_VERSION)
endif