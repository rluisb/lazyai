# Implementation Plan: Independent Packages, CI/CD, and Hexagonal Architecture

This document outlines the step-by-step plan to implement independent CI/CD pipelines, manage monorepo dependencies, and refactor the codebase towards a Hexagonal Architecture.

---

## Phase 1: CI/CD Independence & GoReleaser Setup

**Goal:** Break apart the monolithic CI/CD workflows so each package builds, tests, and releases independently.

1.  **Remove Monolithic Workflows:**
    *   Delete `.github/workflows/go-ci.yml`.
    *   Delete `.github/workflows/release.yml`.
2.  **Create Package-Specific CI Workflows:**
    *   Create `.github/workflows/ci-cli.yml` (triggered on paths `packages/cli/**`).
    *   Create `.github/workflows/ci-orchestrator.yml` (triggered on paths `packages/orchestrator/**`).
    *   Create `.github/workflows/ci-diffviewer.yml` (triggered on paths `packages/diffviewer/**`).
    *   *Note:* Each workflow will run `go test` and `go build` with `GOWORK=off` to ensure `go.mod` integrity.
3.  **Configure GoReleaser for Each Package:**
    *   Create `packages/cli/.goreleaser.yaml`.
    *   Create `packages/orchestrator/.goreleaser.yaml`.
    *   Create `packages/diffviewer/.goreleaser.yaml`.
    *   Configure cross-platform builds (Linux, macOS, Windows) for `amd64` and `arm64` in each config.
4.  **Create Package-Specific Release Workflows:**
    *   Create `.github/workflows/release-cli.yml` (triggered on tags `packages/cli/v*`).
    *   Create `.github/workflows/release-orchestrator.yml` (triggered on tags `packages/orchestrator/v*`).
    *   Create `.github/workflows/release-diffviewer.yml` (triggered on tags `packages/diffviewer/v*`).
    *   These workflows will use the `goreleaser/goreleaser-action`.

---

## Phase 2: Monorepo Dependency Management

**Goal:** Ensure packages explicitly depend on published versions of each other, while maintaining fast local development.

1.  **Audit `go.mod` Files:**
    *   Ensure `packages/cli/go.mod` explicitly requires `github.com/rluisb/lazyai/packages/orchestrator` and `github.com/rluisb/lazyai/packages/diffviewer` at specific versions (e.g., `v1.0.0`).
    *   Ensure `GOWORK=off go mod tidy` passes in all package directories.
2.  **Update Makefile:**
    *   Refactor the `Makefile` to support the new independent structure, removing the monolithic `release-dry-run` in favor of GoReleaser local builds (`goreleaser build --snapshot --clean`).

---

## Phase 3: Hexagonal Architecture Refactoring

**Goal:** Decouple core domain logic from external frameworks (Cobra, Charm, OS, APIs).

### Step 3.1: Refactor `orchestrator`
1.  **Create Directory Structure:** `domain/`, `ports/`, `service/`, `adapters/`.
2.  **Define Ports:** Extract interfaces for external dependencies (e.g., `FileStore`, `Database`, `MCPClient`) into `ports/`.
3.  **Isolate Domain/Service:** Move core workflow execution and state management into `service/` and `domain/`. Ensure these packages do not import `os`, `net/http`, or `modernc.org/sqlite`.
4.  **Implement Adapters:** Move SQLite logic, MCP-go logic, and file system operations into specific packages under `adapters/` (e.g., `adapters/sqlite`, `adapters/mcp`).

### Step 3.2: Refactor `diffviewer`
1.  **Create Directory Structure:** `domain/`, `ports/`, `adapters/`.
2.  **Isolate Domain:** Separate the diff parsing and hunk management logic into `domain/`.
3.  **Implement Adapters:** Move the Bubble Tea UI rendering into `adapters/tui/`.

### Step 3.3: Refactor `cli` (Primary Adapter)
1.  **Clean up `cmd/`:** Ensure Cobra commands contain zero business logic. They should only parse flags and call the injected orchestrator/diffviewer services.
2.  **Dependency Injection (Composition Root):** Update `packages/cli/cmd/lazyai-cli/main.go` to initialize all secondary adapters (SQLite, FileStore), inject them into the orchestrator/diffviewer services, and pass those services to the Cobra root command.

---

## Phase 4: Meaningful Testing Strategy

**Goal:** Implement tests focused on behavior, utilizing the new decoupled architecture.

1.  **Domain Tests:** Write pure, table-driven unit tests for the `orchestrator` and `diffviewer` domain logic. These should be extremely fast and require no mocks.
2.  **Application Tests:** Write unit tests for the `orchestrator` services using in-memory "Fakes" for the `ports` (e.g., an in-memory `FileStore` fake).
3.  **Adapter Tests:** Write narrow integration tests for the `adapters` (e.g., testing the SQLite adapter against a real, temporary SQLite database file).
4.  **Remove Brittle Tests:** Identify and remove any existing tests that heavily mock standard libraries or test trivial implementation details.

---

## Execution Strategy

To minimize disruption, the implementation will be executed in the following order:
1.  **Phase 1 & 2 (CI/CD & Dependencies):** Establish the independent pipelines first. This ensures that subsequent architectural changes can be verified independently.
2.  **Phase 3.2 (Diffviewer Refactor):** Refactor the smallest package first as a proof-of-concept for the Hexagonal Architecture.
3.  **Phase 3.1 & 3.3 (Orchestrator & CLI Refactor):** Refactor the core engine and wire it up to the CLI.
4.  **Phase 4 (Testing):** Write the new behavior-focused tests alongside the refactoring in Phase 3.
