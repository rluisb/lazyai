# Research: Issue #571 — Copilot per-agent tools instead of hardcoded blanket list

**Issue:** [#571](https://github.com/rluisb/lazyai/issues/571)
**Epic:** [#568](https://github.com/rluisb/lazyai/issues/568) — Cross-CLI agent-tools alignment
**Prerequisite:** [#569](https://github.com/rluisb/lazyai/issues/569) — Canonical agent tool-capability model
**Verified:** 2026-06-29

---

## Problem Statement

The Copilot adapter emits `.agent.md` files with a hardcoded `tools:` list for **every** agent, regardless of the agent's actual capability. The blanket list `["read", "search", "edit", "shell"]` over-permissions read-only agents (researcher, reviewer, evidence-verifier) with `edit` and `shell`. Skills separately hard-code `tools: ["*"]`, granting unrestricted tool access.

---

## Evidence

### 1. Hardcoded blanket list in `copilotAgentMarkdownContent`

**File:** `packages/cli/internal/adapter/copilot.go:322`

```go
b.WriteString("\ntools: [\"read\", \"search\", \"edit\", \"shell\"]\n---\n\n")
```

The function `copilotAgentMarkdownContent` (lines 300–331) reads `name` and `description` from source frontmatter but ignores all other fields. Every emitted `.agent.md` receives the identical `tools:` list irrespective of the source agent's `role:`, `mode:`, or any future `tools:` / `readonly:` field that #569 will introduce.

### 2. Three call sites share this single transform

`copyCopilotAgents` (lines 247–298) calls `copilotAgentMarkdownContent` via `CopyWithRecord` at three points:
- `copilot.go:254` — for `defaultAgentID` (the guide agent)
- `copilot.go:275` — for `ctx.LibraryFS`-based agents (embed path)
- `copilot.go:293` — for disk-based agents (`LibraryDir` path)

All three paths converge on the same transform, so any fix in `copilotAgentMarkdownContent` covers all emission paths.

### 3. Skills use `tools: ["*"]` — a separate gap

**File:** `packages/cli/internal/adapter/copilot.go:491–492`

```
tools:
  - "*"
```

`skillToCopilotAgentMarkdown` (lines 461–497) emits wildcard tool access for every skill-derived agent. This is a separate concern and is out of scope for #571 (see §Out of Scope below).

### 4. Canonical agents have no machine-readable capability field

**Files:** `packages/cli/library/canonical/agents/*.md`

All eight canonical agents (`deployer`, `evidence-verifier`, `guide`, `implementer`, `planner`, `researcher`, `responder`, `reviewer`) carry:

| Agent | `mode:` | Description says |
|---|---|---|
| researcher | `all` | "read-only codebase explorer" |
| reviewer | `all` | "Universal verifier. Read-only." |
| evidence-verifier | `all` | (verification role, no write intent) |
| deployer | `all` | full capability |
| guide | `all` | full capability |
| implementer | `all` | full capability |
| planner | `all` | full capability |
| responder | `all` | full capability |

None carry a `tools:` allowlist or `readonly:` flag. `AgentSpecRaw` (`packages/cli/internal/frontmatter/agent_spec.go`) currently parses `tier`, `temperature`, `thinking`, `risk`, `multimodal` — no capability/tools field. **This is exactly the gap #569 closes.**

### 5. Copilot's native per-agent tool vocabulary

Per `docs/ai-cli-tools/tool-systems/agent-tools-matrix.md:24`:

> Copilot — `tools:` list in `.agent.md` | lowercase set: `read`, `search`, `edit`, `shell` (+ `*`) | allowlist

The correct read-only emission (per the matrix, row "read-only role"):

```yaml
tools: ["read", "search"]
```

Full-capability emission:

```yaml
tools: ["read", "search", "edit", "shell"]
```

### 6. Existing test hard-asserts the blanket list

**File:** `packages/cli/internal/adapter/adapter_frontmatter_test.go:162–180`

```go
func TestCopilotAgentMarkdownContent_BaselineNoTier(t *testing.T) {
    source := baselineAgentSource("implementer", "Universal implementer.")
    out := copilotAgentMarkdownContent(source)
    // ...
    if !strings.Contains(string(out), `tools: ["read", "search", "edit", "shell"]`) {
        t.Errorf("missing expected tools array:\n%s", out)
    }
}
```

This test currently asserts the broken behavior. It must be updated to reflect the post-#569 derived list. `baselineAgentSource` (defined in `adapter_test_helpers_test.go:14`) emits name + description only (no `tools:`/`readonly:` field), so it will need to be exercised with both a read-only agent fixture and a full-capability fixture once the canonical schema lands.

### 7. Integration test coverage of agent emission

**File:** `packages/cli/internal/adapter/adapter_frontmatter_test.go:182–243` —  
`TestCopilotAdapter_DefaultSevenBaselineAgentsOnly` installs the full Copilot adapter against an in-memory FS and checks agent file presence/naming. It does not assert `tools:` content per agent, so it is not directly broken but will need a `tools:` assertion extension to become meaningful coverage.

---

## Capability mapping (Copilot)

Once #569 defines the canonical tool-grant vocabulary, the Copilot adapter must translate as follows:

| Canonical grant | Copilot tool name |
|---|---|
| `read` (file-read) | `read` |
| `search` | `search` |
| `edit` (file write/edit) | `edit` |
| `shell` | `shell` |
| `web` | (no Copilot equivalent — omit) |
| `mcp` | (resolved via VS Code/CLI MCP — omit from tools list) |
| `spawn` | (no per-agent spawn mechanism in Copilot — omit) |

Read-only agents receive grants `{read, search}` → emit `tools: ["read", "search"]`.  
Full-capability agents receive all grants → emit `tools: ["read", "search", "edit", "shell"]`.

---

## Out of Scope

- **Skill `tools: ["*"]`** (`skillToCopilotAgentMarkdown`, line 461) — separate behavioral gap; #571 does not touch skills.
- **`AgentSpecRaw` schema changes** — owned by #569; #571 consumes the result.
- **Any other adapter** — #570 (Claude Code), #572 (OpenCode), #573 (OMP) are parallel siblings.

---

## Conclusion

The bug is localized to `copilotAgentMarkdownContent` (one function, ~31 lines) and its single test assertion. The fix is straightforward once #569 delivers a parseable capability field: read the field in the transform, map it to the Copilot tool vocabulary, and emit the derived list. The test must be updated to assert per-capability behavior. The fix carries no risk to the skill emission path or any other adapter.
