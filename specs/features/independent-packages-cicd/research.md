# Research Report: Independent Packages, CI/CD, and Hexagonal Architecture

This document synthesizes the research conducted by specialized agents on three core topics: CI/CD Independence, Hexagonal Architecture, and Meaningful Testing Strategies for the LazyAI monorepo.

---

## 1. CI/CD & Dependency Strategy

To support independent releases and Go's module system, the repository must be structured as a **Multi-Module Monorepo**.

### Repository Architecture
*   **Individual `go.mod` files:** Each package (`packages/cli`, `packages/orchestrator`, `packages/diffviewer`) must have its own `go.mod` file.
*   **`go.work` for Local Development:** A `go.work` file links these modules together for local development.
*   **Important CI Rule:** `go.work` is strictly for local development. CI workflows should run *without* `go.work` (`GOWORK=off`) to ensure the `go.mod` files accurately reflect the required published versions.

### GitHub Actions CI/CD Strategy
*   **Splitting Workflows:** Create discrete workflow files for each package (e.g., `.github/workflows/ci-cli.yml`). Use GitHub Actions `paths` filtering to trigger CI only when relevant files change.
*   **Cross-Platform Build Matrix:** Use a build matrix to compile for Linux, macOS, and Windows across required architectures (amd64, arm64).

### Release and Tagging Scheme
*   **Go-Compliant Monorepo Tagging:** To version a module located in a subdirectory, the Git tag **must** be prefixed with the directory path (e.g., `packages/cli/v1.2.0`).
*   **Automated Release Workflows:** Trigger release workflows based on these specific tag patterns. Using GoReleaser is highly recommended for cross-platform Go releases.

### Managing Inter-Dependencies
When a change is made in a lower-level package (`orchestrator`) that an upper-level package (`cli`) needs, you must perform a **Topological Release**:
1.  **Release the Dependency:** Merge changes for `packages/orchestrator` and tag the release (`packages/orchestrator/v1.3.0`).
2.  **Update the Consumer:** Update `packages/cli/go.mod` to require the new version (`GOWORK=off go get github.com/rluisb/lazyai/packages/orchestrator@v1.3.0`).
3.  **Release the Consumer:** Tag the new CLI release (`packages/cli/v2.1.0`).

---

## 2. Hexagonal Architecture Strategy

Hexagonal architecture (Ports and Adapters) divides the system into concentric layers. Dependencies always point inward toward the core domain.

### Decoupling Strategy
*   **The `cli` Package (Primary Adapter):** Acts solely as the entry point and delivery mechanism. Contains zero business logic. CLI commands receive their dependencies (primary ports) during initialization.
*   **The `orchestrator` Package (Core Domain):** The engine that manages workflows and handles state. Defines what it needs from the outside world via **Driven Ports** (e.g., `FileStore`, `APIClient`).
*   **The `diffviewer` Package (Domain + Adapter):** Separate the core logic (parsing patch formats) from the terminal UI rendering (Bubble Tea).

### Proposed Directory Structure
```text
packages/
├── cli/                        # PRIMARY ADAPTER
│   ├── cmd/                    # Cobra commands setup
│   └── tui/                    # Charm/Bubbletea models
├── orchestrator/               # CORE DOMAIN
│   ├── domain/                 # Pure Go domain models
│   ├── ports/                  # Go Interfaces (FileStore, APIClient)
│   ├── service/                # Application logic
│   └── adapters/               # SECONDARY ADAPTERS (fsstore, llmclient)
├── diffviewer/                 # CORE DOMAIN + UI
│   ├── domain/                 # Diff logic, hunk management
│   ├── ports/                  # Interfaces for rendering
│   └── adapters/               # Bubble Tea specific rendering
└── cmd/
    └── lazyai/
        └── main.go             # THE WIRE-UP LAYER (Dependency Injection)
```

### The Role of `main.go` (Dependency Injection)
The `main.go` file is the ultimate composition root. It initializes secondary adapters, injects them into the core domain, and injects the core domains into the primary adapter (CLI).

---

## 3. Testing Strategy: Behavior Over Coverage

In Hexagonal Architecture, the testing philosophy shifts to testing behaviors, not implementations.

### Testing the Domain Logic (The Core)
*   **Pure, Blazing Fast Unit Tests:** Because domain entities have no external dependencies, you do not need mocks here.
*   **State and Invariants:** Test that domain objects maintain their internal invariants.
*   **Table-Driven Tests:** Define multiple scenarios (success, edge cases, failure states) in a single slice of structs.

### Testing the Application Layer (Use Cases)
*   **Interfaces are the Boundary:** Define small, focused interfaces in the application layer.
*   **Fakes over Mocks:** Writing an in-memory "Fake" implementation of a repository is often better than using a mocking framework.
*   **Mocking Frameworks:** Use when you need to simulate specific error conditions or verify side-effects.

### Testing Adapters
*   **Outbound (Driven) Adapters:** Use narrow integration tests. Do not use unit tests with mocked database drivers. Use tools like Testcontainers to test the contract with real infrastructure.
*   **Inbound (Driving) Adapters:** Use handler/controller tests. Mock the Application Service interface and use `httptest` to simulate HTTP requests.

### The Balanced Testing Pyramid
1.  **Domain Unit Tests:** High volume, extremely fast. Near 100% coverage expected.
2.  **Application Unit Tests:** High volume, fast. Uses fakes/mocks for outbound ports.
3.  **Adapter Integration Tests:** Medium volume, slower. Uses real infrastructure.
4.  **End-to-End (E2E) Tests:** Low volume, slowest. Tests the fully assembled application.
