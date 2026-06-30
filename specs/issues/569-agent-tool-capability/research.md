# Issue 569 Research: Canonical Agent Tool Capability Model

Date: 2026-06-29
Issue: https://github.com/rluisb/lazyai/issues/569
Epic: https://github.com/rluisb/lazyai/issues/568
Foundation: PR #577 merged at `8c2dfbb56a2f95d4afa2a7fd69b4743861c2dd57`

## Scope

Add a machine-readable per-agent tool capability declaration to canonical LazyAI agents so adapters can translate that capability into target-native restrictions in follow-up issues #570-#573 and conditional issues #574-#575.

This research does not implement adapter-specific behavior.

## Observed issue facts

Issue #569 states:

- Canonical agents under `packages/cli/library/canonical/agents/*.md` have no machine-readable per-tool capability field.
- `researcher.md` is contradictory: `mode: all` while the description says read-only.
- `reviewer.md` is also described as read-only without a tool restriction.
- Acceptance requires canonical agents to carry a capability field, read-only roles to be marked consistently, and the vocabulary to match the matrix doc.

## Current canonical agents

Observed files:

```text
packages/cli/library/canonical/agents/evidence-verifier.md
packages/cli/library/canonical/agents/guide.md
packages/cli/library/canonical/agents/planner.md
packages/cli/library/canonical/agents/researcher.md
packages/cli/library/canonical/agents/responder.md
packages/cli/library/canonical/agents/reviewer.md
packages/cli/library/canonical/agents/deployer.md
packages/cli/library/canonical/agents/implementer.md
```

Read-only roles identified from descriptions/system prompts:

| Agent | Evidence | Current contradiction |
|---|---|---|
| `researcher` | Description: "read-only codebase explorer" | `mode: all` |
| `reviewer` | Description includes "Read-only" | `mode: all` |
| `evidence-verifier` | System prompt forbids fabricated/inferred evidence and only evaluates claims against sources | `mode: all` |

Full-capability roles:

```text
guide
planner
implementer
responder
deployer
```

## Current parser surface

`packages/cli/internal/frontmatter/agent_spec.go` defines `AgentSpecRaw` and `ParseAgentSpec` for tier/model routing metadata:

- `tier`
- `temperature`
- `thinking`
- `risk`
- optional `multimodal`

Canonical agents use generic LazyAI frontmatter (`role`, `mode`, `temperature`, `steps`, `skills`) and do not carry `tier`. Therefore capability parsing must not route canonical agents through `ParseAgentSpec`; doing so would error on missing `tier`.

The same package already owns generic frontmatter extraction helpers through `ExtractFrontmatter` / `ExtractField`, and avoids import cycles with adapter/model packages. This makes `packages/cli/internal/frontmatter` the right home for a target-neutral agent tool grant parser.

## Existing `tools:` usage

Repository search found `tools:` in tests/internal fixtures, but no canonical agent currently declares a `tools:` field. Notable existing fixture:

```go
// packages/cli/internal/frontmatter/agent_spec_test.go
// tools: memory qmd
```

That fixture belongs to `ParseAgentSpec` coverage and uses legacy MCP-ish tokens. #569 should avoid breaking `ParseAgentSpec`; the new capability parser can validate only when callers opt into it.

## Vocabulary decision from foundation matrix

Use target-neutral canonical tokens:

| Token | Meaning |
|---|---|
| `read` | File/content read |
| `edit` | File write/edit |
| `shell` | Shell/process execution |
| `search` | Search/grep/glob-style discovery |
| `web` | Web fetch/search |
| `mcp` | MCP tools |
| `spawn` | Subagent/delegation |

Recommended read-only grant:

```yaml
tools:
  - read
  - search
```

Recommended full grant:

```yaml
tools:
  - read
  - edit
  - shell
  - search
  - web
  - mcp
  - spawn
```

## Compatibility / defaulting

For backward compatibility, absent or empty `tools:` should parse as unrestricted (`nil` grants). This lets external/user agents without the new field preserve current adapter behavior while canonical library agents become explicit.

Unknown tokens should return an error from the new parser so adapter code does not silently emit unsupported restrictions.

## Naming collision risk

`packages/cli/internal/adapter/capabilities.go` already defines adapter-surface capability concepts (what a target emits: agents, skills, hooks, etc.). The per-agent tool capability model is different. Avoid names like `Capability`; prefer names such as:

- `AgentToolGrant`
- `ParseAgentToolGrants`

## Downstream dependencies

#569 blocks adapter issues:

- #570 Claude Code: translate canonical grants to Claude `tools:` / `disallowedTools` with capitalized built-ins.
- #571 Copilot: replace blanket `tools: ["read", "search", "edit", "shell"]` with grant-derived lists.
- #572 OpenCode: derive `permission`/`mode`/gate map from grants.
- #573 OMP: emit OMP-native frontmatter and restrict read-only agents.

Conditional downstream work:

- #574 Kiro: if runtime verification requires JSON agent profiles, translate grants into Kiro `tools` / `allowedTools`.
- #575 Antigravity: if subagent emission is chosen, translate grants into enable flags / permissions stance.

## Verification targets

Focused tests for #569 should cover:

1. Full token list parses in stable order.
2. Read-only token list parses as read/search.
3. Missing `tools:` returns `nil` / unrestricted.
4. Empty `tools:` returns `nil` / unrestricted.
5. Unknown token returns an error that names the offending token.
6. Existing `ParseAgentSpec` behavior remains unchanged for its legacy `tools: memory qmd` fixture.

Focused command after implementation:

```bash
cd packages/cli
go test ./internal/frontmatter
```

## Open questions for plan approval

1. Should read-only canonical agents remove `mode: all` entirely, or should it be replaced with a non-contradictory existing value if the repo already has one? Current evidence only supports removing the contradiction; no validated alternate mode value was observed.
2. Should `web` be included for read-only agents? Current issue language names researcher/reviewer as read-only code/evidence roles. The conservative grant is `read` + `search`; web can be added later if an agent prompt explicitly needs external fetch/search.
