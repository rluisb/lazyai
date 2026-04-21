# Spec 018: Codex Deep Setup — Research

**Date:** 2026-04-21
**Status:** Research — awaiting human gate
**Sources:**
- [Codex CLI features](https://developers.openai.com/codex/cli/features)
- [Codex CLI reference](https://developers.openai.com/codex/cli/reference)
- [AGENTS.md guide](https://developers.openai.com/codex/guides/agents-md)
- [Skills](https://developers.openai.com/codex/skills)
- [Custom prompts (DEPRECATED)](https://developers.openai.com/codex/custom-prompts)
- [Slash commands](https://developers.openai.com/codex/cli/slash-commands)

---

## §1 — What Codex CLI supports (2026)

### Core surfaces

| Surface | Path | Purpose |
|---|---|---|
| AGENTS.md | `<repo>/AGENTS.md` (hierarchical cwd → root), `AGENTS.override.md` | System instructions; override wins where present |
| Config | `~/.codex/config.toml`, `<project>/.codex/config.toml` | Model, approval, MCP servers, project-doc fallback names |
| **Skills** | `.agents/skills/<name>/SKILL.md` (repo), `~/.agents/skills/` (user), `/etc/codex/skills/` (admin) | Reusable instructions + optional `scripts/`, `references/`, `assets/`, `agents/openai.yaml`. **Replaces deprecated custom prompts.** |
| Custom prompts | `~/.codex/prompts/*.md` | **DEPRECATED** — use skills instead |
| MCP | `config.toml.[mcp_servers.*]` + `codex mcp add` | External tool integrations |

### Codex-specific behaviors

- **Two-root split**: config at `.codex/`, skills at `.agents/skills/` (unlike Claude's `.claude/` or OpenCode's `.opencode/`). ai-setup already handles this via `ResolveCodexRoots`.
- **Hierarchical AGENTS.md**: Codex concatenates AGENTS.md from repo root down; `AGENTS.override.md` shadows base at its scope.
- **`codex exec`** is the non-interactive entrypoint: `codex exec --skip-git-repo-check "<prompt>"`.
- **No custom slash commands**: built-ins only (`/model`, `/plan`, `/permissions`, `/agent`, `/mcp`, `/compact`, `/diff`, …). Unlike Claude/Gemini/Copilot, users **cannot** register their own.
- **No extension/plugin bundle**: no marketplace or bundle format comparable to Claude plugins or Gemini extensions.

### `codex mcp add` CLI surface (already wired into ai-setup via `--drive-cli`)

```
codex mcp add <name> --env KEY=VALUE -- <command> [args…]
codex mcp add <name> --url <http-url> --bearer-token-env-var <VAR>
```

### `codex exec` CLI surface

```
codex exec [--skip-git-repo-check] [--json] [--ephemeral] "<prompt>"
```

`--skip-git-repo-check` is the flag we need for post-install validation — current implementation omits it, causing the probe to fail with "Not inside a trusted directory" even when it's pointed at a real temp workspace during tests.

### What Codex does NOT have
- **No user-defined slash commands** (only built-ins; skills are the extensibility surface)
- **No extension / plugin bundle format** (nothing to generate for a distribution surface analog to specs 016/017)
- **No modes / output-styles** comparable to Claude's `output-styles/` or OpenCode's `modes/`

---

## §2 — What ai-setup already has for Codex

From specs 008, 009, 010:

| Feature | Status | Where |
|---|---|---|
| `AGENTS.md` + `AGENTS.override.md` at all scopes | ✅ | `scaffold/root.go` + `library/root/AGENTS.*.template.md` |
| `.codex/config.toml` TOML deep-merge | ✅ | `adapter/codex.go` + `configmerge.MergeTOMLFile` |
| Split roots (`.codex/` for config, `.agents/skills/` for skills) | ✅ | `ResolveCodexRoots` |
| Skills copy as `<name>/SKILL.md` | ✅ | `CopyLibraryDirectory` with per-skill dir transform |
| Orchestrator as a skill (when enabled) | ✅ | `GetOrchestratorSkillContent` |
| `--drive-cli` via `codex mcp add` | ✅ | Spec 010 |
| MCP compile via `config.toml` with `[mcp_servers.*]` enrichment | ✅ | Spec 009 |
| Post-install validation (`codex exec`) | ⚠ **broken** | Missing `--skip-git-repo-check` flag — probe always fails |

---

## §3 — Actual gaps worth closing

### G1 — Broken validation (critical)

`CodexAdapter.RunHeadlessValidation` runs `codex exec "check .agents/ structure"` but omits `--skip-git-repo-check`, so every install in a non-repo directory (including tests) emits:

> `Not inside a trusted directory and --skip-git-repo-check was not specified.`

**Cost: one-line fix.** High value — closes a ship blocker that's silently degrading the install UX.

### G2 — No dedicated `library/codex/` directory

Compare to:
- `library/claudecode/` (spec 012) — commands, output-styles, rules
- `library/opencode/` (spec 011) — commands, modes, plugins
- `library/copilot/` (spec 013) — agents, instructions
- `library/gemini/` (spec 017) — commands

Codex has no per-tool home. Skills, AGENTS.md, and config.toml all come from generic locations. Adding `library/codex/` gives:
- **Parity** with the other four deep-setup specs
- **A home for Codex-specific skills** tuned to Codex's capabilities (`/plan` mode hint, Codex-friendly sandbox notes, `--skip-git-repo-check` idiom)
- **Codex-specific AGENTS.override.md starter** (ship a template that recipients can extend, mirroring `library/claudecode/rules/`)

### G3 — No bundle / extension generator

**Not applicable.** Codex has no plugin or extension format. The analog to specs 016 (Claude plugin) / 017 (Gemini extension) does not exist upstream. **This is a hard constraint, not a gap.**

### G4 — No Codex-specific skills

Codex's skill layout (`SKILL.md` in a directory with optional `scripts/`, `references/`, `assets/`, `agents/openai.yaml`) is richer than what the generic `library/skills/*.md` ships. A deep-setup spec could add a `library/codex/skills/` with skills that exercise the Codex-specific affordances (e.g., an `agents/openai.yaml` sidecar for UI config).

Low priority without concrete user demand — flag-only.

### G5 — Post-install enrichment beyond `codex exec`

Claude spec 012 added a post-install summary (`claude mcp list` + agent count). Codex offers:
- `codex mcp list` — list registered MCP servers (available via CLI)
- No agent-count equivalent (Codex has no subagent concept parallel to Claude's)

Shipping an optional `codex mcp list` summary after install would be a nice touch. Medium-low value.

---

## §4 — Options

| Option | Description | Effort |
|---|---|---|
| **A. Validation fix only** | Add `--skip-git-repo-check` to `RunHeadlessValidation`; nothing else | XS (~10 LOC + test) |
| **B. Validation fix + `library/codex/` per-tool dir** | A, plus create `library/codex/` with a Codex-specific AGENTS.override template and placeholder skill directory (empty but discovered) | S (~100 LOC + template content) |
| **C. B + post-install MCP summary** | B, plus `codex mcp list` summary after install (mirrors spec 012 Claude pattern) | M (~180 LOC) |
| **D. C + Codex-specific skills in `library/codex/skills/`** | C, plus ship 1-2 Codex-tuned skills exercising `agents/openai.yaml` sidecar | L (~300 LOC + skill content) |
| **E. Doc-only close** | Accept Codex is already functional; document why extension/plugin analog doesn't apply | XS |

**Recommendation: Option C.** It closes the ship blocker (G1), delivers visible parity (G2), and gives Codex the same post-install courtesy Claude already has (G5). Option D is speculative without a concrete Codex-skill use case. Option A is too thin given the parity discussion we just had.

Rejected:
- A alone leaves library asymmetry visible; the parity talk was the whole reason we started spec 017.
- D ships skill content without a concrete user ask.
- E ignores the broken validation and leaves the per-tool layout asymmetric.

---

## §5 — Decision interview

1. **Q1** — Pick Option A (fix only), B (fix + library dir), **C (fix + library dir + MCP summary)**, D (C + Codex skills), or E (doc-only)?
2. **Q2** — If B/C/D: content for `library/codex/AGENTS.override.template.md` — should it ship a raw template with `[YOUR_*]` placeholders (recipient-fills) or a prefilled Codex-idiom document (preset conventions like "use plan mode for long tasks", "ask before destructive ops")?
3. **Q3** — `codex mcp list` summary at install time: always run when `codex` binary is on PATH, or gate behind a flag?
4. **Q4** — If D: ship Codex skills as part of this spec or queue as a follow-up?
5. **Q5** — How should the library restructure handle the generic `library/skills/*.md`? Leave them shared across all tools (current), or move tool-specific variants under `library/codex/skills/` to mirror spec 017's restructure?

---

## §6 — Risks & constraints

| # | Risk | Mitigation |
|---|---|---|
| R1 | `--skip-git-repo-check` flag-name drift in future Codex releases | Pin the usage to a documented flag; log stderr output so a regression is visible, non-fatal |
| R2 | `codex mcp list` output format changes | Parse only loosely (line count or substring match); never fail install on a parse error |
| R3 | Library restructure breaks skills discovery at install time | Codex adapter already reads `library/skills/` for generic skills. If we add `library/codex/skills/`, resolver pattern from spec 017 applies (prefer tool-specific, fall back to generic). |
| R4 | AGENTS.override template collision with project-committed override | Write only if file doesn't exist; existing files untouched (current adapter pattern) |

---

## §7 — Out of scope

- Extension/plugin generator — **no upstream concept** for Codex; hard constraint.
- Custom slash commands — upstream doesn't support user-defined ones.
- Output styles / modes — no Codex equivalent.
- `~/.codex/prompts/` (deprecated surface) — skip; OpenAI recommends migrating to skills.
- Codex `agents/openai.yaml` sidecar — flag for future spec if we ship Codex-specific skills.
