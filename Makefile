.PHONY: build test lint vet clean install release cross-build

BINARY_NAME=ai-setup
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-dev")
LDFLAGS=-ldflags "-s -w -X cmd.Version=$(VERSION)"
GO=go

build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) .

test:
	$(GO) test ./... -v -count=1

test-short:
	$(GO) test ./... -short -count=1

test-coverage:
	$(GO) test ./... -coverprofile=coverage.out -count=1
	$(GO) tool cover -html=coverage.out -o coverage.html

vet:
	$(GO) vet ./...

lint: vet
	@echo "Linting Go files..."

fmt:
	gofmt -w .
	goimports -w .

tidy:
	$(GO) mod tidy

clean:
	rm -f $(BINARY_NAME) coverage.out coverage.html

install: build
	cp $(BINARY_NAME) /usr/local/bin/

cross-build:
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

dev:
	$(GO) build -o $(BINARY_NAME) . && ./$(BINARY_NAME) $(ARGS)

all: vet test build
