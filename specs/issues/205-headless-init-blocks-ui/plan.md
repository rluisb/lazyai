# Plan: 205-headless-init-blocks-ui

**Feature ID:** 205
**Spec:** [./spec.md](./spec.md)
**Date:** 2026-05-11
**Status:** Draft
**Owner:** orchestrator
**Constitution:** .specify/memory/constitution.md

> **Purpose.** A plan describes *how* the system in `spec.md` will be built. It selects the tech stack, names the major modules, evaluates the design against the constitution, and breaks the work into tasks. Acceptance criteria stay in the spec; behavior contracts live here only when they are implementation details (data model, internal APIs).

---

## Summary

Refactor `runHeadlessInit()` in `packages/cli/cmd/init.go` to run headless init for each tool concurrently using goroutines and `sync.WaitGroup`, instead of sequentially. This unblocks the CLI prompt immediately after scaffolding completes while headless init runs in the background.

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language(s) | Go 1.21+ | Existing codebase |
| Framework(s) | Standard library `sync` | No external dependencies needed |
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
| I — Library-First | PASS | Using only standard library `sync` package |
| II — Test-First (NON-NEGOTIABLE) | PASS | Tests added for parallel execution behavior |
| III — Docs as Source of Truth | N/A | No documentation changes |
| IV — Anti-Speculation (YAGNI) | PASS | Only parallel execution added, no extra features |
| V — Simplicity Over Abstraction | PASS | Goroutines + WaitGroup is the simplest Go pattern |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | PASS | Straightforward refactor, no abstraction layers |

**Verdict:** APPROVED

---

## Project Structure

```
[repo-root]/
└── packages/cli/cmd/init.go            ← modified (lines 249-266)
```

---

## Implementation Steps

### Step 1: Add `sync` import
**File:** `packages/cli/cmd/init.go`
**Location:** Around line 10, where imports are declared
**Change:** Add `"sync"` to the import block

### Step 2: Refactor the headless init loop
**File:** `packages/cli/cmd/init.go`
**Location:** Lines 249-266 (`runHeadlessInit` function)
**Change:** Wrap the sequential loop in goroutines with `sync.WaitGroup`

**Before:**
```go
for _, tool := range ctx.Tools {
    adapt, err := reg.Get(tool)
    if err != nil {
        continue
    }
    adapterCtx := &adapter.AdapterContext{...}
    cmdLog.Info("running headless populate", "tool", tool)
    if err := adapt.RunHeadlessInit(adapterCtx, prompt); err != nil {
        cmdLog.Warn("headless init failed", "tool", tool, "error", err)
    }
}
```

**After:**
```go
var wg sync.WaitGroup
for _, tool := range ctx.Tools {
    wg.Add(1)
    go func(tool string) {
        defer wg.Done()
        adapt, err := reg.Get(tool)
        if err != nil {
            return
        }
        adapterCtx := &adapter.AdapterContext{
            TargetDir:  ctx.TargetDir,
            HomeDir:    ctx.HomeDir,
            SetupScope: ctx.SetupScope,
            LibraryFS:  ctx.LibraryFS,
        }
        cmdLog.Info("running headless populate", "tool", tool)
        if err := adapt.RunHeadlessInit(adapterCtx, prompt); err != nil {
            cmdLog.Warn("headless init failed", "tool", tool, "error", err)
        }
    }(tool)
}
wg.Wait()
```

### Step 3: Add tests
**File:** `packages/cli/cmd/init_headless_test.go` (new file)
**Coverage:**
- Verify goroutines are spawned for multiple tools
- Verify WaitGroup.Wait blocks until all goroutines complete
- Verify error logging still works

---

## Out of Scope

- Adding outer timeout (each adapter already has 120s timeout)
- Changing `mechanicalFill` or `updatePopulateNeeded`

---

## Downstream Contract

| Produces for | Filename |
|---|---|
| `speckit-tasks` | this file + spec.md → tasks |
| `speckit-implement` | this file + task harnesses |