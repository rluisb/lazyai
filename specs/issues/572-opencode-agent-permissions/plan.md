# Plan: Issue #572 — OpenCode: canonical agents get no capability-derived permission/mode/tools

**Status:** DRAFT — pending human gate
**Blocked on:** #569 (canonical agent capability model — must merge first)
**Part of:** Epic #568 (cross-CLI agent-tools alignment)
**Date:** 2026-06-29

---

## Blocker

**Do not implement any code in this issue before #569 merges.**

#569 defines the canonical capability vocabulary (the `readonly:` flag or `tools:` allowlist that adapters consume). Encoding speculative mappings before the vocabulary lands would require rework when #569's actual shape is known. All tasks below assume #569 has merged and its capability signal is consumable from `canonical/agents/*.md` frontmatter or an equivalent mechanism.

---

## Scope

Derive `permission:` and optionally `mode:` for OpenCode canonical agent files from the capability model introduced by #569. Implement the mapping in `RewriteAgentForOpenCode` and validate with tests.

**In scope:**
- `packages/cli/internal/adapter/agent_transform.go` — `RewriteAgentForOpenCode` function
- `packages/cli/internal/adapter/opencode_adapter_test.go` — update conflicting test; add capability-derived permission tests
- `packages/cli/internal/adapter/opencode_frontmatter_test.go` — add test for read-only permission emission if not already present

**Out of scope:**
- The hardcoded `plan`/`build`/`explore` blocks in `opencode.json` — those are OpenCode session modes, not canonical agents; do not touch.
- MCP server names — must never appear in the `tools:` gate map; see Constraint §5.
- `opencode_frontmatter.go` — `OpenCodeAgentOpts` struct and `BuildOpenCodeAgentFrontmatter` already support all needed fields; no changes expected unless schema changes.
- Source `mode: all` in canonical agent frontmatter — stripped by current code; no change to that behavior.
- Other adapters (#570/#571/#573/#574/#575) — handled in their respective issues.

---

## Capability → OpenCode Mapping (draft, subject to #569 vocabulary)

The exact field names below depend on #569's delivery. The mapping logic is:

### Read-only agents
Agents marked read-only by #569 (expected: researcher, reviewer, evidence-verifier):

```yaml
permission:
  edit: deny
  bash: deny
```

These agents read files and produce text; they must never modify the working tree or execute shell commands.

### Planning agents
Agents that write plans or specs (expected: planner):

```yaml
permission:
  edit: ask   # may write a plan doc, but user confirmation required
  bash: deny  # no shell execution
```

### Full-write agents
Agents that implement code or operate infrastructure (expected: implementer, deployer, responder):

```yaml
permission:
  edit: ask   # writes code; user confirms destructive ops
  bash: ask   # shell access; user confirms
```

### General-purpose agents
Agents with no specific restriction signal (expected: guide):

Omit `permission:` entirely — OpenCode defaults to allowing all tools. This preserves the existing behavior for unspecialized agents.

### `mode:` derivation (conditional)
If #569 introduces a subagent flag or equivalent, read-only agents may be emitted with `mode: subagent` to restrict them from acting as the primary agent. This mapping is optional and must only be applied if #569's capability model explicitly supports it. Do not emit `mode:` speculatively.

**Note:** the current test `TestOpenCodeAdapter_Install_FromFS` (line 99-101) asserts `mode` must NOT be present in baseline output. If `mode:` derivation is added, this assertion must be updated to validate per-agent expected values rather than blanket absence.

### MCP server names — hard constraint
The `Tools map[string]bool` field in `OpenCodeAgentOpts` uses **OpenCode built-in tool names** (`bash`, `edit`, `write`, `read`, etc.). It must never be populated with MCP server identifiers (e.g., `filesystem`, `ripgrep`, `memory`). Source frontmatter `tools:` lines containing MCP server names are already stripped by `BuildOpenCodeAgentFrontmatter` — this must remain true after #572. The capability-derived gate map must only reference OpenCode-native names.

---

## Tasks

These tasks are ordered; do not begin implementation of any task before #569 merges.

### Task 0 — Read #569 delivery (gate)
After #569 merges, read the resulting capability field shape in `canonical/agents/*.md`. Confirm:
- What field name carries the capability signal (e.g., `readonly:`, `tools:`, `capability:`).
- Which agents are marked read-only.
- Whether the vocabulary includes a `subagent` signal usable for `mode:` derivation.

Adjust the mapping table above if #569's actual shape differs.

### Task 1 — Implement capability-derived permission in `RewriteAgentForOpenCode`

**File:** `packages/cli/internal/adapter/agent_transform.go`

Current state: `_ = mode` and `_ = ctx` — both discarded.

After #569:
1. Parse the capability signal from the source frontmatter (or from `ctx` if #569 attaches the model there).
2. Derive the `permission` map:
   - Read-only capability → `{ edit: deny, bash: deny }`
   - Write-capable planning role → `{ edit: ask, bash: deny }`
   - Write-capable execution role → `{ edit: ask, bash: ask }`
   - No signal → omit `permission` (nil map; `BuildOpenCodeAgentFrontmatter` skips nil)
3. Optionally derive `mode:` if #569's vocabulary supports it.
4. Pass the derived values to `BuildOpenCodeAgentFrontmatter` via `OpenCodeAgentOpts`.

No changes to `BuildOpenCodeAgentFrontmatter` are anticipated — it already emits `Permission` and `Mode` when non-zero.

### Task 2 — Update `TestOpenCodeAdapter_Install_FromFS`

**File:** `packages/cli/internal/adapter/opencode_adapter_test.go:74-102`

The current loop asserts `mode` must be absent from all agent files. After #572:
- If `mode:` is derived and emitted: replace the blanket absence check with per-agent assertions.
- If `mode:` is not derived (only `permission:` is added): the `mode` absence check remains valid; add a `permission` presence check for read-only agents instead.

Either way, add per-agent assertions for read-only permission:
```go
// researcher, reviewer, evidence-verifier must deny edit and bash
```

### Task 3 — Add read-only permission test

Add a focused test that installs a real agent file and verifies the emitted frontmatter has `permission.edit: deny` and `permission.bash: deny`. This covers the acceptance criterion: "Read-only agents deny edit/bash (verified by test)."

Candidates:
- `TestOpenCodeAdapter_CanonicalReadOnlyAgentsGetPermission` in `opencode_adapter_test.go`
- Test `researcher.md`, `reviewer.md`, and `evidence-verifier.md` in a single table-driven test.

### Task 4 — Verify MCP non-leakage invariant is preserved

Run `TestBuildOpenCodeAgentFrontmatter_DropsSourceExtraKeys` — it must still pass unchanged. If the implementation accidentally populates `Tools` from source frontmatter `tools:` (MCP names), this test will catch it.

### Task 5 — Verify FR-013 compliance

Ensure no `tools:` (deprecated form) or `maxSteps:` keys are emitted by the new code path. `TestOpenCodeAdapter_DefaultConfigIncludesSkillSurface` already guards this at the config level; add an equivalent check in the new canonical-agent test if `BuildOpenCodeAgentFrontmatter` is called in a path that might emit `maxSteps`.

---

## Acceptance Criteria

From issue #572:

1. Canonical OpenCode agents carry capability-derived `permission:` and/or `mode:` in their emitted frontmatter.
2. Read-only agents (researcher, reviewer, evidence-verifier) emit `permission: { edit: deny, bash: deny }` — verified by test.
3. MCP server names do not appear in the `tools:` gate map — existing test `TestBuildOpenCodeAgentFrontmatter_DropsSourceExtraKeys` continues to pass.
4. Only files in `specs/issues/572-opencode-agent-permissions/` and `packages/cli/internal/adapter/` are changed (no touches to other adapter files, library agents, or unrelated surfaces).
5. All existing adapter tests pass.

---

## Test Plan Summary

| Test | Action | Rationale |
|---|---|---|
| `TestOpenCodeAdapter_Install_FromFS` (line 99-101) | Update | Currently forbids `mode` — must be reconciled with derived mode (or extended to check `permission`) |
| `TestOpenCodeAdapter_CanonicalReadOnlyAgentsGetPermission` | Add | Core acceptance criterion for read-only restriction |
| `TestBuildOpenCodeAgentFrontmatter_DropsSourceExtraKeys` | Preserve | MCP non-leakage guard — must pass unchanged |
| `TestOpenCodeAdapter_DefaultConfigIncludesSkillSurface` | Preserve | FR-013 + hardcoded mode permissions — unaffected by #572 |

---

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| #569 uses a field name different from the draft mapping | Medium | Task 0 re-reads #569 delivery before any code is written |
| `mode:` derivation conflicts with OpenCode's primary-loop behavior | Low | Only derive `mode: subagent` if capability model explicitly signals it; default is omit |
| Permission map serialization changes byte-identity of tracked files | Low | `BuildOpenCodeAgentFrontmatter` already uses `sortedStringKeys` for determinism; no change needed |
| `TestOpenCodeAdapter_Install_FromFS` update overfits to specific agents | Medium | Use table-driven checks per role, not per file name, so new agents don't break the test |

---

## Human Gate

Human Gate: PENDING

<!-- The human approver records approval here before implementation begins.
     Format: "Approved by <name> on YYYY-MM-DD — proceed to implementation."
     AI-generated approvals are rejected by the pre-commit hook. -->
