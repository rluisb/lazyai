# Research: Cross-CLI Agent-Tools Alignment

**Epic:** #568
**Date:** 2026-06-29
**Status:** research

## Problem

LazyAI compiles canonical agents (`packages/cli/library/canonical/agents/*.md`) into
seven tool-native targets. Canonical agents carry no machine-readable per-tool
capability, so each adapter improvises its agent-tools emission:

- Claude Code (`RewriteAgentForClaudeCode`) emits name + description only — no `tools`/`disallowedTools`.
- OpenCode (`RewriteAgentForOpenCode`) emits description + managed marker only — no permission/mode.
- Copilot (`copilotAgentMarkdownContent`, `copilot.go:322`) hardcodes a blanket `tools: ["read","search","edit","shell"]`.
- OMP (`omp.go:48-58`) copies canonical agents verbatim; LazyAI-only fields leak.
- Kiro (`kiro.go:39-44`) copies `.md` verbatim; format (JSON vs `.md`) is contested.
- Antigravity (`antigravity.go`) emits no agent files; capability expressed via skills only.
- Pi: no upstream per-agent tool mechanism — legitimately exempt.

## Affected Surface

- `packages/cli/library/canonical/agents/*.md` (capability source of truth)
- `packages/cli/internal/adapter/agent_transform.go` (Claude/OpenCode/Codex/Copilot helpers)
- `packages/cli/internal/adapter/copilot.go`, `opencode.go`, `opencode_frontmatter.go`, `omp.go`, `kiro.go`, `antigravity.go`
- `docs/ai-cli-tools/tool-systems/` reference docs + `agent-tools-matrix.md`

## Patterns

- Adapters parse generic frontmatter via the `frontmatter` package; a shared capability
  parser must live where every adapter can import it without cycles.
- `capabilities.go` already defines a `Capability` type for *adapter surfaces* — the new
  per-agent tool capability MUST use a distinct name to avoid confusion.

## Unknowns / Risks

- Kiro agent format contradiction: `kiro.md` (JSON) vs `specs/030-kiro-cli-v3-output-gaps/spec.md:35` (`.md` tolerated). Requires verification (#574).
- Antigravity stance is a product decision (#575); scope expanded to emit tool capability across every surface that supports tool usage (subagents, skills, workflows, hooks, commands).
- `agent_transform.go` is a single-file collision point for #570 + #572.

## Reference

The cross-CLI compatibility matrix and per-target tool-system reference docs
(this PR) are the authoritative mapping from canonical capability to each
target's native per-agent tool model.
