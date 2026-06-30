# Research: Issue #572 — OpenCode: canonical agents get no capability-derived permission/mode/tools

**Status:** Research complete — awaiting #569 merge before planning gates open.
**Part of:** Epic #568 (cross-CLI agent-tools alignment)
**Prerequisite:** #569 (canonical agent capability model)
**Date:** 2026-06-29

---

## 1. Problem Statement

OpenCode only applies per-agent `permission` for the three hardcoded built-in agents (`plan`, `build`, `explore`) in the root `opencode.json`. Canonical library agents (`researcher`, `reviewer`, `planner`, `implementer`, `deployer`, `guide`, `responder`, `evidence-verifier`) are emitted to `.opencode/agents/` with **description + managed marker only** — no `permission:`, `mode:`, or `tools:` gate is derived from capability. As a consequence, read-only canonical agents (researcher, reviewer, evidence-verifier) are **unrestricted** at runtime.

---

## 2. Evidence from Issue #572

Cited directly from the issue body:

- `packages/cli/internal/adapter/opencode.go:73-101` → hardcoded `plan`/`build`/`explore` permission blocks (3 built-in agents only; no canonical agents).
- `packages/cli/internal/adapter/opencode.go:134-141` + `agent_transform.go` `RewriteAgentForOpenCode` → canonical agents emit description only.
- `opencode_frontmatter.go` already supports a `Tools` gate map + `Permission` map (currently unused for canonical agents).
- Note: #550 already fixed bundled-mode deprecated `tools:` frontmatter — this issue is separate (canonical agents).

---

## 3. Code Walkthrough

### 3.1 `opencode.go` — Hardcoded permission in `opencode.json`

**File:** `packages/cli/internal/adapter/opencode.go:68-101`

```go
"permission": map[string]any{
    "skill": map[string]any{"*": "allow"},
},
"agent": map[string]any{
    "plan": map[string]any{
        "permission": map[string]any{
            "edit": "deny",
            "bash": "ask",
            "skill": map[string]any{"*": "allow"},
        },
    },
    "build": map[string]any{
        "permission": map[string]any{
            "edit": "ask",
            "bash": "ask",
            "skill": map[string]any{"*": "allow"},
        },
    },
    "explore": map[string]any{
        "permission": map[string]any{
            "edit": "deny",
            "bash": "deny",
            "skill": map[string]any{"*": "allow"},
        },
    },
},
```

This covers OpenCode's three built-in named modes (`plan`/`build`/`explore`). These are **OpenCode session modes**, not canonical library agent names. Canonical agents are emitted as individual `.md` files in `.opencode/agents/` — a completely separate surface.

**Install transform call** (`opencode.go:134-141`): the `CopyLibraryDirectory` transform invokes `RewriteAgentForOpenCode(content, ctx, "")`, passing an empty string for `mode` and ignoring any capability signal.

### 3.2 `agent_transform.go` — `RewriteAgentForOpenCode` ignores capability

**File:** `packages/cli/internal/adapter/agent_transform.go:134-151`

```go
func RewriteAgentForOpenCode(source []byte, ctx *AdapterContext, mode string) ([]byte, error) {
    _ = mode  // IGNORED
    _ = ctx   // IGNORED
    fm, body, err := frontmatter.ExtractFrontmatter(source)
    // ...
    return BuildOpenCodeAgentFrontmatter(cleaned, OpenCodeAgentOpts{
        Description:   description,
        ManagedMarker: managedAgentMarker("opencode", name),
    }), nil
}
```

Both `ctx` (which will carry the capability model once #569 delivers) and `mode` are explicitly discarded. Only `description` and `ManagedMarker` are populated. The `Tools`, `Permission`, and `Mode` fields of `OpenCodeAgentOpts` are never set.

### 3.3 `opencode_frontmatter.go` — Infrastructure already present

**File:** `packages/cli/internal/adapter/opencode_frontmatter.go:29-66`

`OpenCodeAgentOpts` already contains all needed fields:

| Field | Type | OpenCode frontmatter key | Status |
|---|---|---|---|
| `Mode` | `string` | `mode:` (primary/subagent/all) | Unused for canonical agents |
| `Tools` | `map[string]bool` | `tools:\n  bash: true` | Unused for canonical agents |
| `Permission` | `map[string]string` | `permission:\n  edit: deny` | Unused for canonical agents |

The comment on `Tools` explicitly guards the MCP-name leakage concern (line 54-57):
> "Note: these are OpenCode's BUILT-IN tool gates (write/edit/bash), NOT MCP server names. If nil or empty, the key is omitted and opencode enables all tools by default. MCP servers belong in `.mcp.json`, not here. (#199 Bug 1)"

`BuildOpenCodeAgentFrontmatter` correctly emits all three fields when non-zero.

### 3.4 Existing test guards the MCP non-leakage invariant

**File:** `opencode_frontmatter_test.go:82-102`

`TestBuildOpenCodeAgentFrontmatter_DropsSourceExtraKeys` uses source `tools: filesystem ripgrep memory` (MCP server names) and asserts the key is **absent** from emitted output. This test already enforces the constraint that MCP server names must never become the `tools:` gate map.

### 3.5 Current test FORBIDS `mode:` in agent output

**File:** `opencode_adapter_test.go:99-101`

```go
if _, ok := fm["mode"]; ok {
    t.Errorf("%s: mode key should not be emitted in baseline OpenCode output", e.Name())
}
```

`TestOpenCodeAdapter_Install_FromFS` currently asserts that `mode` is NOT emitted. This test guards the current baseline (description-only). Implementing capability-derived `mode:` will break this assertion and require updating the test to reflect the new expected behavior per agent.

---

## 4. Canonical Agents — Current State

All 8 canonical agents currently declare `mode: all` in their source frontmatter. This is stripped by `RewriteAgentForOpenCode` (source frontmatter is not forwarded). The mismatch between source `mode: all` and the agent descriptions creates the issue #569 contradiction:

| Agent | Source `mode:` | Description claim | Correct restriction |
|---|---|---|---|
| `researcher` | `all` | "read-only codebase explorer" | edit:deny, bash:deny |
| `reviewer` | `all` | "Read-only" | edit:deny, bash:deny |
| `evidence-verifier` | `all` | verify claims against source (read) | edit:deny, bash:deny |
| `planner` | `all` | produces plans (markdown writes) | edit:ask, bash:deny |
| `implementer` | `all` | builds from specs, writes code | edit:ask/allow, bash:ask |
| `deployer` | `all` | infrastructure and CI/CD ops | edit:ask, bash:ask |
| `guide` | `all` | front-door, general purpose | no restriction (default) |
| `responder` | `all` | SRE, incident response | edit:ask, bash:ask |

Note: the exact permission values for non-read-only agents depend on the capability vocabulary #569 will define. Only the read-only set (researcher/reviewer/evidence-verifier) has unambiguous expected values: `{ edit: deny, bash: deny }`.

---

## 5. Dependency: Issue #569

Issue #572 is **blocked on #569** for implementation. #569 delivers:

1. A machine-readable capability field on canonical agents (e.g., `readonly: true` or a `tools:` allowlist).
2. Corrected read-only annotations for researcher, reviewer, evidence-verifier (fixing the `mode: all` contradiction).
3. A documented capability vocabulary that all per-adapter children (#570–#575) translate from.

#572 must not encode speculative capability mappings before #569 defines the vocabulary — doing so would embed assumptions the canonical model does not yet guarantee.

---

## 6. Constraint: No MCP Server Names in Gate Map

OpenCode's `tools:` frontmatter gate (`map[string]bool`) is a **built-in tool allowlist** using OpenCode-native names (`bash`, `edit`, `write`, `read`, etc.). It is completely separate from MCP server identifiers (e.g., `filesystem`, `ripgrep`, `memory`). The existing comment in `OpenCodeAgentOpts.Tools`, the existing test `TestBuildOpenCodeAgentFrontmatter_DropsSourceExtraKeys`, and issue #199 all enforce this separation. #572 must preserve this invariant: capability-derived permission/tools must only reference OpenCode built-in names, never MCP server names.

---

## 7. OpenCode Schema Reference

Relevant OpenCode per-agent frontmatter fields (from `opencode_frontmatter.go` and OpenCode schema):

| Field | Values | Purpose |
|---|---|---|
| `mode` | `primary` / `subagent` / `all` | when agent is available (primary loop, subagent call, or both) |
| `permission.edit` | `allow` / `ask` / `deny` | file-edit gate |
| `permission.bash` | `allow` / `ask` / `deny` | shell-execution gate |
| `permission.write` | `allow` / `ask` / `deny` | file-write gate |
| `tools.<name>` | `true` / `false` | explicit built-in tool on/off (OpenCode names only) |

The existing hardcoded config (`opencode.go:73-101`) uses the `permission` pattern (not `tools`) — the implementation for canonical agents should use the same pattern for consistency with FR-013 (`permission` over deprecated `tools`/`maxSteps`).

---

## 8. Related Tests and What They Guard

| Test | File | What it guards |
|---|---|---|
| `TestOpenCodeAdapter_DefaultConfigIncludesSkillSurface` | `opencode_adapter_test.go:480-588` | plan/build/explore hardcoded permissions; FR-013 no deprecated tools/maxSteps |
| `TestOpenCodeAdapter_Install_FromFS` | `opencode_adapter_test.go:17-103` | `mode` must NOT be emitted in baseline output (will conflict with #572) |
| `TestBuildOpenCodeAgentFrontmatter_DropsSourceExtraKeys` | `opencode_frontmatter_test.go:82-102` | MCP server names must not appear in tools gate |
| `TestBuildOpenCodeAgentFrontmatter_ExplicitOptsOverrideSource` | `opencode_frontmatter_test.go:37-80` | Permission + Tools opts correctly emitted |
| `TestBuildOpenCodeAgentFrontmatter_DeterministicOrder` | `opencode_frontmatter_test.go:104-124` | tools/permission keys sorted; output byte-stable |

---

## 9. Summary of Gaps

1. `RewriteAgentForOpenCode` discards both `ctx` and `mode` — no capability signal consumed.
2. `OpenCodeAgentOpts` is populated with only `Description` and `ManagedMarker` for canonical agents.
3. Read-only canonical agents (researcher, reviewer, evidence-verifier) receive no `permission` restriction.
4. The existing `TestOpenCodeAdapter_Install_FromFS` test asserts `mode` is absent — will conflict with capability-derived `mode:` emission and must be updated.
5. The source `tools:` field (MCP server names) is already correctly stripped by `BuildOpenCodeAgentFrontmatter` — this guard must be preserved.
6. All of the above is safely deferrable to post-#569 because the capability vocabulary is not yet defined.
