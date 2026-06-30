# Plan: Claude Code — agents drop tool restrictions (#570)

**Issue:** [#570](https://github.com/rluisb/lazyai/issues/570)
**Epic:** #568 (Cross-CLI Agent-Tools Alignment)
**Status:** DRAFT — awaiting human gate
**Blocker:** #569 must merge before implementation begins

---

## Scope

Translate canonical agent tool-capability grants into Claude Code `tools:`/`disallowedTools:`
frontmatter during the `RewriteAgentForClaudeCode` transform in `agent_transform.go`.
Read-only canonical agents (researcher, reviewer, evidence-verifier) must no longer emit
unrestricted Claude subagent files. Non-read-only agents must retain full tool access.

### In scope

- `packages/cli/internal/adapter/agent_transform.go` — extend `RewriteAgentForClaudeCode`
  to read the canonical capability field introduced by #569 and emit `tools:`/`disallowedTools:`
  with correctly capitalized Claude built-in names.
- `packages/cli/internal/adapter/claudecode_frontmatter_test.go` — add assertions that
  read-only agents emit restricted tool lists and that full-access agents are not over-restricted.

### Out of scope

- OpenCode adapter (`RewriteAgentForOpenCode`) — covered by #572 in the same file; must be
  coordinated but is a separate change.
- Copilot adapter — covered by #571.
- OMP adapter — covered by #573.
- Kiro, Antigravity, Pi adapters — covered by #574, #575, and the intentional no-op respectively.
- Changes to canonical agent source files — owned by #569.
- LazyAI tier/model selection — no change to `resolveCtxFor` or `agentSpecToModelsSpec`.

---

## Prerequisite: #569

This plan is **blocked** on #569 merging. #569 will:
1. Add a machine-readable capability field to canonical agent frontmatter
   (likely `tools:` allowlist and/or `readonly: true` flag).
2. Fix `researcher.md`, `reviewer.md`, and `evidence-verifier.md` to declare read-only
   consistently (currently `mode: all` despite read-only descriptions).
3. Export a Go parser (expected: `frontmatter.ParseAgentToolGrants` or equivalent) that adapters can call to obtain structured grants. Do not route this through the existing tier-oriented `ParseAgentSpec` unless #569 explicitly extends it safely.

All task-level decisions below that reference "the #569 capability field" must be reconciled
with #569's actual exported API before implementation starts.

---

## Tasks

### T1 — Read the #569 capability field in `RewriteAgentForClaudeCode`

After #569 merges, determine the exported parser / type for the canonical capability field.
Replace the `_ = ctx` discard in `RewriteAgentForClaudeCode` with a genuine spec parse:

```go
// Before (#570)
_ = ctx
fm, body, err := frontmatter.ExtractFrontmatter(source)

// After (sketch — exact API from #569)
grants, grantErr := frontmatter.ParseAgentToolGrants(source)
// nil grants => unrestricted legacy behavior; non-nil grants drive Claude restrictions.
```

If the #569 parser reports missing/empty capability as unrestricted, preserve the current behavior
(name + description only) so external or partial agents do not break.

### T2 — Build canonical → Claude tool name mapping

Define a `canonicalToClaudeTools(grants []string) (tools []string, disallowed []string)`
helper in `agent_transform.go` (or a sibling package if shared with #572).

Mapping table (pending #569 vocabulary):

| Canonical token | Claude `tools:` entry | Casing |
|---|---|---|
| `read` | `Read`, `Grep`, `Glob` | PascalCase |
| `search` | `Grep`, `Glob` | PascalCase |
| `edit` | `Edit`, `Write` | PascalCase |
| `shell` | `Bash` | PascalCase |
| `web` | `WebFetch`, `WebSearch` | PascalCase |
| `spawn` | `Agent` | PascalCase |
| `mcp` | omit here; MCP tool names are server-specific, not Claude built-in names | n/a |

For read-only agents, emit `disallowedTools: [Edit, Write, Bash]` (deny list). Do not use an explicit read allowlist in this issue; it is easier to get stale as Claude adds or renames read-only built-ins.

Decision: prefer `disallowedTools` for read-only agents to stay forward-compatible as
Claude adds new read-only built-ins that should remain accessible without a plan change.

### T3 — Emit `tools:`/`disallowedTools:` in the frontmatter builder

Extend `RewriteAgentForClaudeCode`'s `strings.Builder` block to conditionally write
`tools:` or `disallowedTools:` after `description:`, space-separated (spec 012 task 004
— no commas):

```go
if len(disallowed) > 0 {
    b.WriteString("disallowedTools: ")
    b.WriteString(strings.Join(disallowed, " "))
    b.WriteByte('\n')
}
if len(tools) > 0 {
    b.WriteString("tools: ")
    b.WriteString(strings.Join(tools, " "))
    b.WriteByte('\n')
}
```

Agents with no capability restriction (full access) must emit neither key — no regression
for full-access agents.

### T4 — Add / update tests in `claudecode_frontmatter_test.go`

Minimum new assertions:

1. **Read-only agents have disallowedTools**: for each of `researcher`, `reviewer`,
   `evidence-verifier`, assert the emitted frontmatter contains `disallowedTools:` with
   at minimum `Edit`, `Write`, `Bash` listed.

2. **Full-access agents are not restricted**: for `implementer`, `deployer`, assert
   neither `tools:` nor `disallowedTools:` restricts the agent (no over-deny of Bash/Edit).

3. **Casing is correct**: all tool names in emitted frontmatter match the PascalCase
   Claude built-in names exactly — no lowercase `edit`, `bash`, etc.

4. **Whitespace separation**: existing `validateAgentsSchemas` check (no comma-space) must
   continue to pass; the new output must use space separation.

Extend `validateAgentsSchemas` to check casing and required presence of `disallowedTools`
for known read-only agent files, or add a dedicated `TestRewriteAgentForClaudeCode_ToolGrants`
test that covers both read-only and full-access cases end-to-end.

### T5 — Coordinate merge with #572 (OpenCode, same file)

`agent_transform.go` hosts both `RewriteAgentForClaudeCode` (this issue) and
`RewriteAgentForOpenCode` (#572). Both issues extend the same file after #569 merges.

Steps:
1. Implement #570 first (or in a dedicated worktree branch).
2. Rebase #572 on top of the merged #570, or sequence merges to avoid conflicts.
3. Do **not** edit `RewriteAgentForOpenCode` as part of this PR.

---

## Acceptance criteria

(from issue #570)

- [ ] Read-only Claude agents (`researcher`, `reviewer`, `evidence-verifier`) cannot
  `Edit`/`Write`/`Bash` — verified by frontmatter test asserting `disallowedTools` presence.
- [ ] Non-read-only agents (`implementer`, `deployer`, etc.) retain full tool access —
  no `tools:`/`disallowedTools:` emitted when not warranted.
- [ ] Tool-name casing matches Claude's built-in names (PascalCase: `Read`, `Bash`, `Edit`, etc.).
- [ ] `mkdocs build --strict` continues to pass (docs not affected by this PR).
- [ ] Existing `claudecode_frontmatter_test.go` tests continue to pass unchanged.

---

## Verification

```bash
# Run only the affected test file
go test ./packages/cli/internal/adapter/... -run TestRewriteAgentForClaudeCode
go test ./packages/cli/internal/adapter/... -run TestClaudeCode
```

No project-wide build or lint run required during planning; full CI runs on the PR.

---

## Sequencing summary

```
#569 merges          ← prerequisite; do not start T1-T4 before this
  └─ T1: read capability field from #569 parser
  └─ T2: define canonical → Claude tool name mapping
  └─ T3: emit tools:/disallowedTools: in frontmatter builder
  └─ T4: add/update tests
  └─ T5: coordinate merge with #572 on agent_transform.go
```

---

<!-- The human approver records approval here. Do NOT let an AI author this line. -->

Human Gate: APPROVED by rluisb at 2026-06-30T09:30:00-03:00
