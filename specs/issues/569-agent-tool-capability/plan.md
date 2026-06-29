# Issue 569 Plan: Canonical Agent Tool Capability Model

Date: 2026-06-29
Issue: https://github.com/rluisb/lazyai/issues/569
Research: `specs/issues/569-agent-tool-capability/research.md`
Status: Draft — pending RPI Plan human approval

## Purpose

Introduce a target-neutral, machine-readable `tools:` capability field on canonical LazyAI agents and a reusable Go parser that follow-up adapters can translate into native tool restrictions.

This plan intentionally stops before adapter-specific behavior. Issues #570-#573 consume this model after #569 merges.

## Planned implementation

### 1. Canonical vocabulary

Define the canonical per-agent tool grant vocabulary as:

```text
read
edit
shell
search
web
mcp
spawn
```

Semantics:

| Token | Meaning |
|---|---|
| `read` | File/content read |
| `edit` | File write/edit |
| `shell` | Shell/process execution |
| `search` | Search/grep/glob-style discovery |
| `web` | Web fetch/search |
| `mcp` | MCP tools |
| `spawn` | Subagent/delegation |

### 2. Canonical agent metadata

Update all canonical agents under `packages/cli/library/canonical/agents/*.md` to carry explicit `tools:`.

Read-only agents:

```yaml
tools:
  - read
  - search
```

Apply to:

```text
researcher
reviewer
evidence-verifier
```

Full-capability agents:

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

Apply to:

```text
guide
planner
implementer
responder
deployer
```

Remove the contradictory `mode: all` from read-only agents unless a repository-supported non-contradictory replacement is found before editing.

### 3. Parser API

Add a frontmatter-package parser, preferably in `packages/cli/internal/frontmatter/agent_tools.go` or adjacent to `agent_spec.go`:

```go
type AgentToolGrant string

const (
    AgentToolRead   AgentToolGrant = "read"
    AgentToolEdit   AgentToolGrant = "edit"
    AgentToolShell  AgentToolGrant = "shell"
    AgentToolSearch AgentToolGrant = "search"
    AgentToolWeb    AgentToolGrant = "web"
    AgentToolMCP    AgentToolGrant = "mcp"
    AgentToolSpawn  AgentToolGrant = "spawn"
)

func ParseAgentToolGrants(source []byte) ([]AgentToolGrant, error)
```

Behavior:

- Missing frontmatter: return `nil, nil` if no `tools:` can be found, preserving unrestricted legacy behavior.
- Missing `tools:`: return `nil, nil`.
- Empty `tools:`: return `nil, nil`.
- YAML sequence: parse and validate tokens.
- Scalar forms: accept comma/space separated strings if existing extraction/coercion patterns make this straightforward; otherwise document and test YAML-list support as the canonical form.
- Unknown token: return an error naming the token.
- Preserve input order so downstream adapters can emit deterministic target-native lists.

Do not extend or reuse adapter-surface `Capability` names; avoid collision with `packages/cli/internal/adapter/capabilities.go`.

### 4. Tests

Add focused tests under `packages/cli/internal/frontmatter` covering:

- Full canonical list parses.
- Read-only list parses.
- Missing `tools:` returns nil unrestricted.
- Empty `tools:` returns nil unrestricted.
- Unknown token returns an error containing the token.
- Existing `ParseAgentSpec` behavior remains unchanged, including its legacy `tools: memory qmd` fixture.

### 5. Documentation alignment

Review `docs/ai-cli-tools/tool-systems/agent-tools-matrix.md` and update only if the canonical vocabulary wording needs alignment after the parser/constants land. Do not change adapter gap statuses in #569 except for a row/section directly describing the canonical model.

## Planned files

```text
packages/cli/library/canonical/agents/*.md
packages/cli/internal/frontmatter/agent_tools.go        # new, if chosen
packages/cli/internal/frontmatter/agent_tools_test.go   # new, if chosen
docs/ai-cli-tools/tool-systems/agent-tools-matrix.md    # optional vocabulary alignment only
```

## Non-goals

- Do not implement Claude, Copilot, OpenCode, OMP, Kiro, or Antigravity translation in this issue.
- Do not regenerate all adapter golden fixtures unless canonical metadata changes force a focused fixture update.
- Do not rename existing adapter-surface capability types.
- Do not add speculative tokens beyond the seven-token vocabulary.
- Do not introduce fallback aliases that hide invalid canonical metadata.

## Verification

Focused verification command:

```bash
cd packages/cli
go test ./internal/frontmatter
```

If docs are changed, also run:

```bash
DISABLE_MKDOCS_2_WARNING=true mkdocs build --strict
```

Project-wide gates and adapter-specific tests are deferred to PR CI and follow-up adapter issues unless #569 changes unexpectedly affect those areas.

## RPI gate

This plan is ready for human approval before implementation.

<!-- The human approver records approval here. Do NOT let an AI author this line. -->
Human Gate: PENDING
