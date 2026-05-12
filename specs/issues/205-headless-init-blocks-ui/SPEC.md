# Spec: 205-headless-init-blocks-ui

**Feature ID:** 205
**Feature name:** headless-init-blocks-ui
**Date:** 2026-05-11
**Status:** Draft
**Owner:** orchestrator
**Constitution:** .specify/memory/constitution.md

> **Purpose.** A specification describes *what* the system should do and *why*, not *how*. Tech stack and architecture belong in `plan.md`. Tasks belong in `tasks.md`. This document is the contract every downstream artifact is judged against.

---

## User Scenarios

### P1 — Interactive init returns promptly
**As a** user running `lazyai-cli init` interactively
**I want** the CLI to return to the prompt immediately after scaffolding completes
**So that** I can continue working while headless init runs in the background

**Acceptance criteria**
- [ ] Given `--no-interactive` mode, when init completes, then total time is ≤ 30 seconds regardless of number of tools (parallel headless init)
- [ ] Given interactive mode, when scaffold completes, then user sees "Setup complete!" and prompt returns within 5 seconds

### P2 — Error reporting not blocked
**As a** user
**I want** any headless init errors to be logged and reported after all tools complete
**So that** I know what succeeded/failed without blocking on success

**Acceptance criteria**
- [ ] Given multiple tools with one failing, when init completes, then failures are logged and summary is shown
- [ ] Given all tools succeed, when init completes, then no error summary is shown

---

## Functional Requirements

| ID | Requirement | Priority | Source story |
|---|---|---|---|
| FR-001 | The headless init loop MUST run tools concurrently using goroutines | P1 | P1 |
| FR-002 | The system MUST use `sync.WaitGroup` to wait for all goroutines before returning | P1 | P1 |
| FR-003 | Errors from individual tools MUST be collected and logged without blocking | P2 | P2 |
| FR-004 | The `mechanicalFill` phase MUST complete before the parallel fan-out | P1 | P1 |
| FR-005 | The `updatePopulateNeeded` phase MUST run after all goroutines complete | P1 | P1 |

---

## Key Entities

| Entity | Description | Lifecycle |
|---|---|---|
| HeadlessInitWaitGroup | sync.WaitGroup tracking parallel init completion | created per runHeadlessInit call |

---

## Success Criteria

- **SC-001 — Prompt return time:** Interactive init returns to prompt within 5 seconds after scaffold completion. Measured by: manual timing.
- **SC-002 — Parallel execution:** With 3 tools enabled, total headless init time is ≤ time of slowest tool (not sum). Measured by: `time lazyai-cli init`.

---

## Edge Cases

- **EC-001 — All tools fail:** When all `RunHeadlessInit` calls fail, the function returns without error (non-fatal, same as current behavior).
- **EC-002 — One tool hangs:** When one tool exceeds its internal 120s timeout, other tools complete normally; timeout is logged.

---

## Assumptions

- **A-001:** `RunHeadlessInit` is thread-safe (each call creates its own `exec.CommandContext`). Confidence: HIGH.
- **A-002:** `AdapterContext` is read-only during headless init. Confidence: HIGH.
- **A-003:** No shared mutable state in the loop that could cause race conditions. Confidence: HIGH.

---

## Out of Scope

- Adding outer timeout to the entire headless init phase (each adapter already has 120s timeout)
- Changing `mechanicalFill` or `updatePopulateNeeded` behavior

---

## Constitutional Notes

- **Article I — Library-First:** N/A — using standard library `sync` package.
- **Article IV — YAGNI:** No additional features beyond parallel execution with error aggregation.
- **Article V — Simplicity:** Goroutines + WaitGroup is the simplest Go pattern for this.

---

## Downstream Contract

| Produced for | Filename |
|---|---|
| `speckit-plan` | this file |
| `speckit-tasks` | indirectly via plan |