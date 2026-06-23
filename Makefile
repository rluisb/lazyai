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
	rm -rf $(CLI_DIR)/$(DIST_DIR) $(DIFFVIEWER_DIR)/$(DIST_DIR)

cli-snapshot:
	rm -rf $(CLI_DIR)/$(DIST_DIR) && mkdir -p $(CLI_DIR)/$(DIST_DIR)
	for target in $(PLATFORMS); do \
		os=$${target%/*}; arch=$${target#*/}; ext=; \
		if [ "$$os" = "windows" ]; then ext=.exe; fi; \
		echo "Building lazyai-cli-$$os-$$arch$$ext"; \
		(cd $(CLI_DIR) && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build -ldflags "$(CLI_VERSION_LDFLAGS)" -o $(DIST_DIR)/lazyai-cli-$$os-$$arch$$ext ./cmd/lazyai-cli); \
	done
	cd $(CLI_DIR)/$(DIST_DIR) && shasum -a 256 lazyai-cli-* > checksums.txt

diffviewer-snapshot:
	rm -rf $(DIFFVIEWER_DIR)/$(DIST_DIR) && mkdir -p $(DIFFVIEWER_DIR)/$(DIST_DIR)
	for target in $(PLATFORMS); do \
		os=$${target%/*}; arch=$${target#*/}; ext=; \
		if [ "$$os" = "windows" ]; then ext=.exe; fi; \
		echo "Building lazyai-diffviewer-$$os-$$arch$$ext"; \
		(cd $(DIFFVIEWER_DIR) && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build -ldflags "-s -w" -o $(DIST_DIR)/lazyai-diffviewer-$$os-$$arch$$ext ./cmd/lazyai-diffviewer); \
	done
	cd $(DIFFVIEWER_DIR)/$(DIST_DIR) && shasum -a 256 lazyai-diffviewer-* > checksums.txt

snapshot: cli-snapshot diffviewer-snapshot

all: vet test build

homebrew-formula:
	@echo "=== Rendering Homebrew formula ==="
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION is required (e.g. VERSION=1.3.0)"; exit 1; fi
	@if [ -z "$(SHA256_DARWIN_ARM64)" ]; then echo "ERROR: SHA256_DARWIN_ARM64 is required"; exit 1; fi
	@if [ -z "$(SHA256_DARWIN_AMD64)" ]; then echo "ERROR: SHA256_DARWIN_AMD64 is required"; exit 1; fi
	sed "s/{{VERSION}}/$(VERSION)/g; s/{{SHA256_DARWIN_ARM64}}/$(SHA256_DARWIN_ARM64)/g; s/{{SHA256_DARWIN_AMD64}}/$(SHA256_DARWIN_AMD64)/g" \
		packaging/homebrew/lazyai-cli.rb.tmpl > lazyai-cli.rb
	@echo "Rendered lazyai-cli.rb — copy to rluisb/homebrew-lazyai repository"
