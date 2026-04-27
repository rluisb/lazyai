.PHONY: build test lint vet clean install release cross-build checksums go-build go-test go-vet go-tidy ts-install ts-build ts-test ts-typecheck orch-build orch-test go-all ts-all all

# ── Go (packages/ai-setup-go) ──────────────────────────────────────────────

GO_DIR = packages/ai-setup-go
BINARY_NAME = ai-setup
GO = go

go-build:
	cd $(GO_DIR) && $(GO) build -o $(BINARY_NAME) .

go-test:
	cd $(GO_DIR) && $(GO) test ./... -v -count=1

go-test-short:
	cd $(GO_DIR) && $(GO) test ./... -short -count=1

go-test-coverage:
	cd $(GO_DIR) && $(GO) test ./... -coverprofile=coverage.out -count=1
	cd $(GO_DIR) && $(GO) tool cover -html=coverage.out -o coverage.html

go-vet:
	cd $(GO_DIR) && $(GO) vet ./...

go-tidy:
	cd $(GO_DIR) && $(GO) mod tidy

go-fmt:
	cd $(GO_DIR) && gofmt -w . && goimports -w .

go-clean:
	rm -f $(GO_DIR)/$(BINARY_NAME) $(GO_DIR)/coverage.out $(GO_DIR)/coverage.html

go-cross-build:
	cd $(GO_DIR) && \
		GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 . && \
		GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 . && \
		GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 . && \
		GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 . && \
		GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

go-checksums: go-cross-build
	cd $(GO_DIR)/dist && shasum -a 256 * > checksums.txt

go-all: go-vet go-test go-build

# ── TypeScript (pnpm workspaces) ────────────────────────────────────────────

ts-install:
	pnpm install

ts-build:
	pnpm run build

ts-test:
	pnpm run test

ts-typecheck:
	pnpm run typecheck

# ── Orchestrator ────────────────────────────────────────────────────────────

orch-build:
	pnpm --filter @ai-setup/orchestrator run build

orch-test:
	pnpm --filter @ai-setup/orchestrator run test

# ── Convenience ─────────────────────────────────────────────────────────────

build: ts-build go-build
test: ts-test go-test
lint: go-vet ts-typecheck

clean:
	rm -rf packages/ai-setup-go/dist packages/ai-setup-ts/dist packages/orchestrator/dist
	rm -f packages/ai-setup-go/$(BINARY_NAME) packages/ai-setup-go/coverage.out packages/ai-setup-go/coverage.html

all: lint test build
