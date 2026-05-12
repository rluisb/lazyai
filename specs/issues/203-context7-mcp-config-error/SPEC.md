# Spec: 203-context7-mcp-config-error

**Feature ID:** 203
**Feature name:** context7-mcp-config-error
**Date:** 2026-05-11
**Status:** Draft
**Owner:** orchestrator
**Constitution:** .specify/memory/constitution.md

> **Purpose.** A specification describes *what* the system should do and *why*, not *how*. Tech stack and architecture belong in `plan.md`. Tasks belong in `tasks.md`. This document is the contract every downstream artifact is judged against.

---

## User Scenarios

### P1 — context7 MCP registers with valid API key
**As a** user with `CONTEXT7_API_KEY` set in environment
**I want** context7 MCP server to register successfully when I use Claude Code
**So that** I can use context7's library documentation lookup

**Acceptance criteria**
- [ ] Given `CONTEXT7_API_KEY` is set, when `claude mcp add-json` is called for context7, then the literal `${CONTEXT7_API_KEY}` is replaced with the actual env var value
- [ ] Given valid Bearer token, when registration completes, then no "Invalid configuration" error occurs

### P2 — context7 gracefully skipped when no API key
**As a** user without `CONTEXT7_API_KEY`
**I want** context7 MCP server to be skipped without blocking setup
**So that** I can use Claude Code without context7

**Acceptance criteria**
- [ ] Given `CONTEXT7_API_KEY` is NOT set, when `claude mcp add-json` is called, then context7 registration is skipped with a warning log
- [ ] Given no API key, when init completes, then setup continues without error

---

## Functional Requirements

| ID | Requirement | Priority | Source story |
|---|---|---|---|
| FR-001 | The system MUST expand `${VAR}` patterns in header values from environment variables | P1 | P1 |
| FR-002 | The system MUST skip context7 registration gracefully when `CONTEXT7_API_KEY` is not set | P1 | P2 |
| FR-003 | The expansion MUST happen for both install-time (`installClaudeMCPViaCLI`) and compile-time (`useCliForMCP`) code paths | P1 | P1 |

---

## Key Entities

| Entity | Description | Lifecycle |
|---|---|---|
| EnvVarExpander | Helper function to expand `${VAR}` patterns | Utility function |

---

## Success Criteria

- **SC-001 — Token expansion:** `CONTEXT7_API_KEY=abc123` results in `Authorization: Bearer abc123` in the JSON payload. Measured by: unit test.
- **SC-002 — Graceful skip:** Missing key results in warning log, not error. Measured by: integration test.

---

## Edge Cases

- **EC-001 — Empty env var:** When `CONTEXT7_API_KEY=""`, skip with warning (same as not set)
- **EC-002 — Nonexistent var in pattern:** When `${NONEXISTENT}` is used, leave as-is or warn (prefer leave as-is for future compatibility)
- **EC-003 — Multiple vars in headers:** Each `${VAR}` pattern should be expanded independently

---

## Assumptions

- **A-001:** Environment variables are available at the time of `mcp add-json` call. Confidence: HIGH.
- **A-002:** `os.Getenv` is the correct way to read env vars in this context. Confidence: HIGH.

---

## Out of Scope

- Adding API key validation (context7 server handles this)
- Changing context7's `enabled` default in catalog

---

## Constitutional Notes

- **Article I — Library-First:** N/A — using only `os.Getenv` from standard library.
- **Article IV — YAGNI:** No key rotation or validation features.
- **Article V — Simplicity:** Single helper function for expansion, check for empty key.

---

## Downstream Contract

| Produced for | Filename |
|---|---|
| `speckit-plan` | this file |
| `speckit-tasks` | indirectly via plan |