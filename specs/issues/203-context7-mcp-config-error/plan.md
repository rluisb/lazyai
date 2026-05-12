# Plan: 203-context7-mcp-config-error

**Feature ID:** 203
**Spec:** [./spec.md](./spec.md)
**Date:** 2026-05-11
**Status:** Draft
**Owner:** orchestrator
**Constitution:** .specify/memory/constitution.md

> **Purpose.** A plan describes *how* the system in `spec.md` will be built. It selects the tech stack, names the major modules, evaluates the design against the constitution, and breaks the work into tasks. Acceptance criteria stay in the spec; behavior contracts live here only when they are implementation details (data model, internal APIs).

---

## Summary

Fix the context7 MCP server registration by expanding `${VAR}` patterns in header values from environment variables before calling `claude mcp add-json`. Also skip context7 gracefully when `CONTEXT7_API_KEY` is not set.

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language(s) | Go 1.21+ | Existing codebase |
| Framework(s) | Standard library `os` | `os.Getenv` for env var expansion |
| Storage | None | No data changes |
| Deployment | Binary | Same as before |
| Telemetry | Structured logging | Same as existing |

**External dependencies (new):** None

**External dependencies (rejected):** None

---

## Constitution Check

The plan MUST evaluate against every relevant article. Write **PASS** / **FAIL** / **N/A** with one sentence of justification per article.

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | PASS | Using only `os.Getenv` from standard library |
| II — Test-First (NON-NEGOTIABLE) | PASS | Tests added for env expansion and graceful skip |
| III — Docs as Source of Truth | N/A | No documentation changes |
| IV — Anti-Speculation (YAGNI) | PASS | Only expansion and skip behavior added |
| V — Simplicity Over Abstraction | PASS | Single helper function, minimal code change |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | PASS | Direct fix, no abstraction layers |

**Verdict:** APPROVED

---

## Project Structure

```
[repo-root]/
├── packages/cli/internal/adapter/claudecode.go    ← modified (add expandEnvVars)
├── packages/cli/internal/adapter/mcp_compiler.go ← modified (add expandEnvVars)
└── packages/cli/internal/adapter/mcp_utils.go    ← new (shared helper)
```

---

## Implementation Steps

### Step 1: Create shared env var expansion helper
**File:** `packages/cli/internal/adapter/mcp_utils.go` (new)
**Function:** `expandEnvVars(s string) string`
**Logic:**
```go
func expandEnvVars(s string) string {
    // Match ${VAR} patterns and replace with os.Getenv("VAR")
    // If VAR not set, leave ${VAR} as-is
    return os.Expand(s, func(key string) string {
        if val := os.Getenv(key); val != "" {
            return val
        }
        return "${" + key + "}" // leave as-is if not set
    })
}
```

### Step 2: Update mcpServerToJSON for env var expansion
**File:** `packages/cli/internal/adapter/claudecode.go`
**Location:** `mcpServerToJSON()` function (lines 254-332)
**Change:** Call `expandEnvVars()` on `srv.Headers[k]` values before writing to JSON

```go
// In the Headers loop (around line 323):
fmt.Fprintf(&buf, `"%s":"%s"`, k, expandEnvVars(v))
```

### Step 3: Add graceful skip for missing API key
**File:** `packages/cli/internal/adapter/claudecode.go`
**Location:** `installClaudeMCPViaCLI()` (around line 235)
**Change:** Before calling `mcp add-json` for context7, check if key is missing and skip

```go
// For context7, check if CONTEXT7_API_KEY is set
if name == "context7" && os.Getenv("CONTEXT7_API_KEY") == "" {
    adapterLog.Warn("context7 skipped: CONTEXT7_API_KEY not set", "adapter", "claude-code")
    continue
}
```

### Step 4: Apply same changes to mcp_compiler.go
**File:** `packages/cli/internal/adapter/mcp_compiler.go`
**Location:** `useCliForMCP()` function (around line 330)
**Change:** Apply same `expandEnvVars` call and skip check

### Step 5: Add tests
**File:** `packages/cli/internal/adapter/mcp_utils_test.go` (new)
**Coverage:**
- `expandEnvVars("Bearer ${KEY}")` with KEY=abc → `"Bearer abc"`
- `expandEnvVars("Bearer ${MISSING}")` → `"Bearer ${MISSING}"`
- Skip behavior when CONTEXT7_API_KEY empty

---

## Out of Scope

- Adding API key validation
- Changing context7's default enabled state

---

## Downstream Contract

| Produces for | Filename |
|---|---|
| `speckit-tasks` | this file + spec.md → tasks |
| `speckit-implement` | this file + task harnesses |