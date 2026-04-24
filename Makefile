.PHONY: help build test typecheck lint clean \
        build-go test-go build-ts test-ts typecheck-ts lint-ts \
        install

help:
	@echo "ai-setup monorepo — root targets"
	@echo ""
	@echo "Cross-language:"
	@echo "  make install      - Install JS deps (pnpm) and tidy Go modules"
	@echo "  make build        - Build both runtimes"
	@echo "  make test         - Run both test suites"
	@echo "  make typecheck    - TS typecheck (no Go equivalent)"
	@echo "  make lint         - Lint both"
	@echo "  make clean        - Remove build artifacts"
	@echo ""
	@echo "Go-only (packages/ai-setup-go):"
	@echo "  make build-go | test-go | cross-build-go"
	@echo ""
	@echo "TS-only (packages/ai-setup-ts):"
	@echo "  make build-ts | test-ts | typecheck-ts | lint-ts"

install:
	pnpm install
	$(MAKE) -C packages/ai-setup-go tidy

build: build-go build-ts

test: test-go test-ts

typecheck: typecheck-ts

lint: lint-ts
	$(MAKE) -C packages/ai-setup-go vet

clean:
	$(MAKE) -C packages/ai-setup-go clean
	rm -rf packages/ai-setup-ts/dist
	rm -rf packages/orchestrator/dist

# ---------- Go targets (delegate) ----------

build-go:
	$(MAKE) -C packages/ai-setup-go build

test-go:
	$(MAKE) -C packages/ai-setup-go test

cross-build-go:
	$(MAKE) -C packages/ai-setup-go cross-build

# ---------- TS targets (delegate via pnpm) ----------

build-ts:
	pnpm --filter ./packages/ai-setup-ts run build

test-ts:
	pnpm --filter ./packages/ai-setup-ts run test

typecheck-ts:
	pnpm --filter ./packages/ai-setup-ts run typecheck

lint-ts:
	pnpm --filter ./packages/ai-setup-ts run lint
