# Plan: Issue #571 — Copilot per-agent tools instead of hardcoded blanket list

**Issue:** [#571](https://github.com/rluisb/lazyai/issues/571)
**Status:** DRAFT — blocked on [#569](https://github.com/rluisb/lazyai/issues/569)
**Scope:** Replace hardcoded blanket `tools:` in `copilotAgentMarkdownContent` with grant-derived Copilot tool names.
**Files changed (expected):** `packages/cli/internal/adapter/copilot.go`, `packages/cli/internal/adapter/adapter_frontmatter_test.go`


---

## Prerequisite

**#569 must merge first.** This plan describes how to consume the canonical tool-capability model that #569 delivers. Do not begin implementation until #569 is merged and its capability field (`tools:`, `readonly:`, or equivalent) is available on canonical agents.

---

## Objective

Replace the single hardcoded string:

```go
b.WriteString("\ntools: [\"read\", \"search\", \"edit\", \"shell\"]\n---\n\n")
```

with a derived list that reads the canonical agent's capability grants and emits only the Copilot-native tools corresponding to those grants. Read-only agents (`researcher`, `reviewer`, `evidence-verifier`) must not receive `edit` or `shell`.

---

## Acceptance Criteria (from issue #571)

1. Each emitted `.agent.md` `tools:` list reflects the source agent's capability.
2. Read-only agents do not include `edit`/`shell` (verified by test).
3. No regression for full-capability agents.

---

## Tasks

### Task 1 — Add capability-to-Copilot-tools translator

**File:** `packages/cli/internal/adapter/copilot.go`

After #569 lands, read the canonical capability field from the source frontmatter inside `copilotAgentMarkdownContent`. Parse it with the helper #569 provides (expected: `ExtractField(fm, "tools")` returning a YAML list, or a `readonly: true` boolean, per #569's chosen schema).

Map each canonical grant to its Copilot equivalent:

| Canonical grant | Copilot tool |
|---|---|
| `read` | `read` |
| `search` | `search` |
| `edit` | `edit` |
| `shell` | `shell` |
| `web`, `mcp`, `spawn` | (omit — no Copilot equivalent) |

Emit the derived list in JSON array style: `tools: ["read", "search"]` or `tools: ["read", "search", "edit", "shell"]`.

**Fallback behavior:** If the source has no capability field (pre-migration agent), fall back to the current full list `["read", "search", "edit", "shell"]` and log a warning. This preserves the existing behavior for unannotated agents while #569's migration propagates.

**Invariant:** The Copilot tool names are always lowercase. Validate the emitted names against the known set `{read, search, edit, shell}` and drop any unrecognized token.

### Task 2 — Update the existing test assertion

**File:** `packages/cli/internal/adapter/adapter_frontmatter_test.go`

`TestCopilotAgentMarkdownContent_BaselineNoTier` (line 162) currently asserts the blanket list. Update it to:

- Use a full-capability fixture (one that declares all grants or has no `readonly:` flag) and assert `tools: ["read", "search", "edit", "shell"]`.
- Add `TestCopilotAgentMarkdownContent_ReadOnly` with a read-only fixture and assert `tools: ["read", "search"]` and that `edit` and `shell` are absent.

`baselineAgentSource` (no `tools:` field) will trigger the fallback from Task 1, which is fine — assert the fallback emits the full list.

### Task 3 — Extend integration test to assert per-agent tools

**File:** `packages/cli/internal/adapter/adapter_frontmatter_test.go`

`TestCopilotAdapter_DefaultSevenBaselineAgentsOnly` (line 182) checks file presence and naming. Add assertions on the emitted `tools:` line for at least one read-only agent file (`reviewer.agent.md`) to verify the integration path (FS-based emission via `copyCopilotAgents`) respects the capability.

---

## Out of Scope

- `skillToCopilotAgentMarkdown` (`copilot.go:461`) and its `tools: ["*"]` — separate issue.
- `AgentSpecRaw` schema changes — #569's responsibility.
- Any other adapter.

---

## Implementation Order

1. ⛔ Wait for #569 to merge.
2. Implement Task 1 (translator in `copilotAgentMarkdownContent`).
3. Implement Task 2 (update unit test).
4. Implement Task 3 (integration test assertion).
5. Run `go test ./packages/cli/internal/adapter/...` — all three acceptance criteria must pass.
6. Open PR referencing #571; mark #571 as resolved.

---

Human Gate: APPROVED by rluisb at 2026-06-30T09:30:00-03:00
<!-- The human approver records approval by replacing the line above with:
     Human Gate: APPROVED — <approver> <date>
     AI-generated approvals are rejected by pre-commit hook and CI. -->
