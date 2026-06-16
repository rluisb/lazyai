.PHONY: build test lint vet clean cli-build cli-test cli-vet cli-tidy diffviewer-build diffviewer-test diffviewer-vet diffviewer-tidy go-work-sync cli-snapshot diffviewer-snapshot snapshot all

GO = go
CLI_DIR = packages/cli
DIFFVIEWER_DIR = packages/diffviewer
DIST_DIR = dist
VERSION ?= 0.0.0-dev
CLI_VERSION_LDFLAGS = -s -w -X github.com/rluisb/lazyai/packages/cli/cmd.Version=$(VERSION) -X github.com/rluisb/lazyai/packages/cli/internal/version.Version=$(VERSION)
PLATFORMS = darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64

cli-build:
	cd $(CLI_DIR) && $(GO) build -o lazyai-cli ./cmd/lazyai-cli

cli-test:
	cd $(CLI_DIR) && $(GO) test ./... -count=1

cli-vet:
	cd $(CLI_DIR) && $(GO) vet ./...

cli-tidy:
	cd $(CLI_DIR) && $(GO) mod tidy

diffviewer-build:
	cd $(DIFFVIEWER_DIR) && $(GO) build -o lazyai-diffviewer ./cmd/lazyai-diffviewer

diffviewer-test:
	cd $(DIFFVIEWER_DIR) && $(GO) test ./... -count=1

diffviewer-vet:
	cd $(DIFFVIEWER_DIR) && $(GO) vet ./...

diffviewer-tidy:
	cd $(DIFFVIEWER_DIR) && $(GO) mod tidy

go-work-sync:
	$(GO) work sync

build: cli-build diffviewer-build

test: cli-test diffviewer-test

vet: cli-vet diffviewer-vet

lint: vet

clean:
	rm -f $(CLI_DIR)/lazyai-cli $(CLI_DIR)/coverage.out $(CLI_DIR)/coverage.html
	rm -f $(DIFFVIEWER_DIR)/lazyai-diffviewer
	rm -f $(CLI_DIR)/dist $(DIFFVIEWER_DIR)/dist

cli-snapshot:
	cd $(CLI_DIR) && goreleaser build --snapshot --clean

diffviewer-snapshot:
	cd $(DIFFVIEWER_DIR) && goreleaser build --snapshot --clean

snapshot: cli-snapshot diffviewer-snapshot

all: vet test build
