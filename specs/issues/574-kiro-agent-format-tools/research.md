# Research: Kiro Agent Format — JSON vs .md Reconciliation

**Issue:** #574 (part of #568)
**Date:** 2026-06-29
**Status:** research-complete / blocked on runtime verification
**Researcher:** agent

---

## Problem Statement

Two source-of-truth artifacts disagree on Kiro's custom agent format:

| Source | Claim | Date |
|---|---|---|
| `docs/ai-cli-tools/tool-systems/kiro.md` | Custom agents are **JSON** at `.kiro/agents/<name>.json` with `tools`/`allowedTools` fields | 2026-06-29 (verified from kiro.dev official docs) |
| `specs/030-kiro-cli-v3-output-gaps/spec.md:35` | "Canonical agent frontmatter is already valid Kiro v3 (unknown YAML keys tolerated); **no agent transform is needed**; kiro.go copies agents verbatim" | 2026-06-24 (verified flag in spec A-003, HIGH confidence) |
| `kiro.go:10-13` (inline comment) | "Kiro CLI v3 discovers custom agent profiles from `.kiro/agents/<name>.md`" | current adapter code |

The adapter currently emits `.md` (see `kiro.go:39-44`):

```go
if err := copyCanonicalDefaultAgent(ctx,
    filepath.Join(kiroDir, "agents", defaultAgentID+".md"),
    nil,
); err != nil {
    return nil, err
}
```

---

## Evidence Gathered

### From `docs/ai-cli-tools/tool-systems/kiro.md`

- **Status:** `verified`, `verified_on: 2026-06-29`, `provenance: official-docs (kiro.dev) via research subagent`
- Custom agents section says explicitly: "**JSON files.**" with table pointing to `.kiro/agents/<name>.json`
- Schema documented with typed fields: `name`, `description`, `tools` (array, whitelist), `allowedTools` (array, auto-approved), `resources`, `prompt`, `model`, `mcpServers`, `includeMcpJson`, `hooks`
- Source URLs: https://kiro.dev/docs/cli/custom-agents/ and https://kiro.dev/docs/cli/custom-agents/creating/
- Creation via `/agent create my-agent` or `kiro-cli agent create my-agent` — both create `.json` files

### From `specs/030-kiro-cli-v3-output-gaps/spec.md`

- Line 35 (`A-003`): "Canonical agent frontmatter needs no Kiro transform — HIGH (verified)."
- The term "tolerated" was the conclusion of a `ResearchAgentFrontmatter` subagent finding
- No URL to the specific kiro.dev doc that confirmed `.md` tolerance was recorded in spec 030
- The 030 spec was authored 2026-06-24 — five days before the kiro.md tool-systems doc (2026-06-29)
- 030's scope was **hooks gaps**, not agent format; agent format was a side-finding

### Contradiction Root Cause

Spec 030's A-003 assertion predates the full tool-systems research pass. The `kiro.md` doc is the more recent, more directly sourced finding (explicit official doc URL to `kiro.dev/docs/cli/custom-agents/`). The spec 030 "tolerated" finding has no backing URL recorded and was peripheral to that spec's main task (hooks).

### Runtime Verification Status

**Blocked.** Attempted `kiro-cli agent validate` to probe whether a `.md` file is discovered as a valid agent:

```
$ kiro-cli agent validate /tmp/probe-agent.md
error: You are not logged in, please log in with kiro-cli login

$ kiro-cli agent validate /tmp/probe-agent.json
error: You are not logged in, please log in with kiro-cli login
```

Both `.md` and `.json` probe files return the same auth error. The validator requires an active Kiro auth session; no unauthenticated discovery test is possible via CLI. Runtime verification remains blocked until a Kiro login token is available.

### Official Docs Signal

The official doc at https://kiro.dev/docs/cli/custom-agents/ describes only the JSON schema and `.json` extension. No mention of `.md` agent files in the custom-agents docs. The `.md` format is used by **steering files** and **skills** (`SKILL.md`) — distinct surfaces. The naming overlap (agents dir + `.md`) likely caused the misidentification in spec 030.

---

## Key Facts

1. **`kiro.md` is the higher-confidence source**: directly sourced from official docs with explicit URL, marked `verified 2026-06-29`.
2. **Spec 030 A-003 is lower-confidence**: no backing URL, peripheral finding, five days older.
3. **Current `kiro.go` emits `.md`**: line 40 writes `defaultAgentID+".md"` — this may be silently ignored by Kiro CLI if `.json` is required.
4. **Runtime verification blocked**: `kiro-cli agent validate` requires auth; no unauthenticated format probe is possible.
5. **`tools`/`allowedTools` mapping depends on #569**: the canonical capability model (machine-readable per-agent tool list) is #569's output; it must land before any adapter can populate these JSON fields.
6. **`kiro.go:39-44` is the only agent emission point** — a single call to `copyCanonicalDefaultAgent` with a `.md` suffix.
7. **No multi-agent emission today**: only the `defaultAgentID` agent is emitted; custom agents from the library are not iterated.

---

## Unknowns

| Unknown | Blocking? | How to resolve |
|---|---|---|
| Whether Kiro CLI discovers `.kiro/agents/<name>.md` alongside `.json` | Yes — determines reconciliation path | Kiro CLI login + manual test, or official docs clarification |
| Whether `.json` files require all fields or have required-vs-optional split | No — schema documented in `kiro.md` | `kiro.md` schema table sufficient for plan |
| `tools`/`allowedTools` canonical values for each agent | Yes for JSON-with-capability emit | Blocked on #569 |

---

## Sources

- `packages/cli/internal/adapter/kiro.go:10-13,39-44` — current adapter code
- `docs/ai-cli-tools/tool-systems/kiro.md` — verified official-docs research (2026-06-29)
- `specs/030-kiro-cli-v3-output-gaps/spec.md:35,A-003,A-124` — prior design decision
- `specs/031-cross-cli-agent-tools-alignment/research.md:37` — cross-CLI research noting the contradiction
- Observed runtime: `kiro-cli agent validate` → `error: You are not logged in, please log in with kiro-cli login`
- https://kiro.dev/docs/cli/custom-agents/
- https://kiro.dev/docs/cli/custom-agents/creating/
