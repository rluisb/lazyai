.PHONY: build test lint vet clean cli-build cli-test cli-vet cli-tidy orchestrator-build orchestrator-test orchestrator-vet orchestrator-tidy diffviewer-build diffviewer-test diffviewer-vet diffviewer-tidy go-work-sync release-dry-run all

GO = go
CLI_DIR = packages/cli
ORCHESTRATOR_DIR = packages/orchestrator
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

orchestrator-build:
	cd $(ORCHESTRATOR_DIR) && $(GO) build -o lazyai-orchestrator ./cmd/lazyai-orchestrator

orchestrator-test:
	cd $(ORCHESTRATOR_DIR) && $(GO) test ./... -count=1

orchestrator-vet:
	cd $(ORCHESTRATOR_DIR) && $(GO) vet ./...

orchestrator-tidy:
	cd $(ORCHESTRATOR_DIR) && $(GO) mod tidy

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

build: cli-build orchestrator-build diffviewer-build

test: cli-test orchestrator-test diffviewer-test

vet: cli-vet orchestrator-vet diffviewer-vet

lint: vet

clean:
	rm -f $(CLI_DIR)/lazyai-cli $(CLI_DIR)/coverage.out $(CLI_DIR)/coverage.html
	rm -f $(ORCHESTRATOR_DIR)/lazyai-orchestrator
	rm -f $(DIFFVIEWER_DIR)/lazyai-diffviewer
	rm -rf $(CLI_DIR)/dist $(ORCHESTRATOR_DIR)/dist $(DIFFVIEWER_DIR)/dist
	rm -rf $(DIST_DIR)

release-dry-run:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	@for target in $(PLATFORMS); do \
		os=$${target%/*}; \
		arch=$${target#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "Building lazyai-cli-$$os-$$arch$$ext"; \
		(cd $(CLI_DIR) && GOOS=$$os GOARCH=$$arch $(GO) build -ldflags "$(CLI_VERSION_LDFLAGS)" -o ../../$(DIST_DIR)/lazyai-cli-$$os-$$arch$$ext ./cmd/lazyai-cli) || exit 1; \
		echo "Building lazyai-orchestrator-$$os-$$arch$$ext"; \
		(cd $(ORCHESTRATOR_DIR) && GOOS=$$os GOARCH=$$arch $(GO) build -ldflags "-s -w" -o ../../$(DIST_DIR)/lazyai-orchestrator-$$os-$$arch$$ext ./cmd/lazyai-orchestrator) || exit 1; \
		echo "Building lazyai-diffviewer-$$os-$$arch$$ext"; \
		(cd $(DIFFVIEWER_DIR) && GOOS=$$os GOARCH=$$arch $(GO) build -ldflags "-s -w" -o ../../$(DIST_DIR)/lazyai-diffviewer-$$os-$$arch$$ext ./cmd/lazyai-diffviewer) || exit 1; \
	done
	cd $(DIST_DIR) && shasum -a 256 lazyai-* > checksums.txt

all: vet test build
