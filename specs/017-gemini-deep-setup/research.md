# Spec 017: Gemini Deep Setup — Research

**Date:** 2026-04-20
**Status:** Research — awaiting human gate
**Sources:**
- [Gemini CLI documentation](https://geminicli.com/docs/)
- [Custom commands](https://geminicli.com/docs/cli/custom-commands/)
- [Extension schema reference](https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/reference.md)
- [Writing extensions](https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/writing-extensions.md)

---

## §1 — What Gemini CLI supports (2026)

### Core surfaces

| Surface | Path | Purpose |
|---|---|---|
| Context file | `<project>/GEMINI.md`, `~/.gemini/GEMINI.md`, nested subdir `GEMINI.md` | System instructions (hierarchical, most specific wins) |
| Settings | `<project>/.gemini/settings.json`, `~/.gemini/settings.json` | Model config, approval mode, `context.fileName`, MCP servers |
| Custom commands | `<project>/.gemini/commands/**/*.toml`, `~/.gemini/commands/` | User-invoked slash commands (`/commit`, `/gcs:sync`, …) |
| Extensions | `~/.gemini/extensions/<name>/` | Bundled manifest that can ship context + commands + MCP + settings + themes |
| MCP | `settings.json.mcpServers` + CLI `gemini mcp add` | External tool integrations |

### `gemini-extension.json` schema

```json
{
  "name": "string",                    // required
  "version": "string",                 // required
  "description": "string",
  "mcpServers": {                      // bundled MCP servers
    "<name>": {
      "command": "...",
      "args": ["..."],
      "cwd": "..."
    }
  },
  "contextFileName": "GEMINI.md",      // default: loads the extension's GEMINI.md
  "excludeTools": ["tool_name"],       // disable built-in tools
  "migratedTo": "...",                 // forward-ref to replacement
  "plan": { "directory": "..." },
  "settings": [                        // user-config prompts at install time
    {
      "name": "API_KEY",
      "description": "Your API key",
      "envVar": "API_KEY",
      "sensitive": true
    }
  ],
  "themes": [ ... ]                    // custom colour themes
}
```

Extension directory layout:

```
~/.gemini/extensions/<name>/
├── gemini-extension.json
├── GEMINI.md              # referenced by contextFileName
└── commands/              # optional custom commands
    ├── commit.toml
    └── nested/
        └── sync.toml     # → /nested:sync
```

### What Gemini does NOT have
- **No agents/subagents concept** — unlike Claude Code. The adapter's current skip is correct.
- **No `gemini --version` non-interactive subcommand** — only the interactive `/about` slash command. Limits post-install validation to binary-presence checks.
- **No `gemini init` shell subcommand** — `/init` is interactive-only.

---

## §2 — What ai-setup already has for Gemini

From specs 008 and 009 (incremental work, no dedicated deep-setup spec):

| Feature | Status | Where |
|---|---|---|
| `GEMINI.md` root emission at all scopes | ✅ | `scaffold/root.go` + `library/root/GEMINI.template.md` (314 lines, parity with CLAUDE.md) |
| Placeholder fill (`[YOUR_ORG]`, `[YOUR_TEAM]`, etc.) | ✅ | `fillClaudeMdPlaceholders` handles CLAUDE/AGENTS/GEMINI uniformly |
| `.gemini/settings.json` with `context.fileName`, model, approval mode | ✅ | `adapter/gemini.go:29-60` |
| `.gemini/skills/<name>/SKILL.md` from canonical `library/skills/` | ✅ | Reuses generic skills library |
| Custom slash commands from `library/commands/*.toml` | ✅ | `adapter/gemini.go:87-99` |
| `--drive-cli` via `gemini mcp add` | ✅ | `installGeminiMCPViaCLI` |
| Orchestrator skill generation | ✅ | `GetOrchestratorSkillContent` |
| MCP compile via settings.json (deep-merge) | ✅ | `compileGeminiMCP` |

**So most of the "deep" surface is already covered.** Gemini is not the weakest tool in ai-setup — it's actually reasonably well wired.

---

## §3 — Actual gaps worth closing

### G1 — No dedicated `library/gemini/` directory

Compare to:
- `library/claudecode/` — commands, output-styles, rules (spec 012)
- `library/opencode/` — commands, modes, plugins (spec 011)
- `library/copilot/` — agents, instructions (spec 013)

Gemini commands live at the generic top-level `library/commands/`. Moving them to `library/gemini/commands/` gives:
- **Parity** with the other three deep-setup specs
- **Room to ship** Gemini-specific assets (extension manifests, themes, playbooks) without polluting the generic library root
- **Clearer per-tool surface** for future contributors

### G2 — No extension bundle generator

Analog to spec 016's `ai-setup build-plugin`: a command that generates a `gemini-extension.json` + `GEMINI.md` + `commands/` tree ready to be `gemini extensions link`-ed or published.

Value:
- New **distribution channel** — users can `gemini extensions install` without running `ai-setup init`
- Matches the Claude plugin story (spec 016) for cross-tool parity
- Reuses the same library content, no duplication

### G3 — No post-install validation

Upstream `gemini` CLI has no non-interactive `--version` or `doctor` subcommand as of my research cutoff. The best we can do:
- `exec.LookPath("gemini")` to verify binary presence
- Parse `~/.gemini/settings.json` back after write and assert it still validates

Low-value but cheap to ship.

### G4 — Template polish (optional, out of scope for spec 017)

`library/root/GEMINI.template.md` could be trimmed / Gemini-tuned, but it's already functional and mirrors CLAUDE.md. Leave alone unless a concrete complaint surfaces.

---

## §4 — Options

| Option | Description | Effort |
|---|---|---|
| **A. Library restructure only** | Move `library/commands/*.toml` → `library/gemini/commands/`; update adapter path; add empty `library/gemini/` as a home for future assets | S (~80 LOC) |
| **B. Extension bundle generator only** | Add `ai-setup build-gemini-extension --out <path>` command. Emits `gemini-extension.json`, copies `GEMINI.md` template, bundles commands. Keep `library/commands/` as-is. | M (~350 LOC) |
| **C. A + B together** | Full parity: per-tool library dir + extension generator. Most coherent story. | L (~450 LOC) |
| **D. Minimal: doc-only close** | Document that Gemini is "done" and close the gap note. Accept the library naming asymmetry. | XS (~0 LOC) |
| **E. Codex first, Gemini later** | Skip Gemini for now; do spec 018 (Codex) instead where the gaps are larger. | — |

**Recommendation: C (library restructure + extension generator).** Delivers visible parity with specs 011/012/013 and ships a new distribution channel matching spec 016's plugin story. Effort is reasonable (~450 LOC) and the code pattern is already proven from spec 016.

Rejected:
- A alone is cosmetic without the extension generator.
- B alone leaves the library asymmetry.
- D defers work the user already asked for.
- E is a judgment call about priority — we should finish what we start.

---

## §5 — Decision interview

1. **Q1** — Pick Option A (restructure only), **B (generator only)**, **C (both)**, D (doc-only), or E (Codex first)?
2. **Q2** — If C: does the extension generator reuse the same `dist/` output tree as spec 016 (e.g., `dist/gemini-extension/`) or write elsewhere?
3. **Q3** — Extension manifest fields: minimum viable (`name`, `version`, `description`, `contextFileName: "GEMINI.md"`) or also ship `mcpServers` from the canonical catalog (same static-only rule as spec 014)?
4. **Q4** — Should the extension's `GEMINI.md` be the **placeholder-filled** version (org/team substituted) or the **raw template** (for the receiver to fill in themselves)?
5. **Q5** — Keep `library/commands/*.toml` as a fallback during the move, or delete after restructure?
6. **Q6** — Extension `name`: `"ai-setup"` (same as plugin) or `"ai-setup-gemini"` (disambiguate)?
7. **Q7** — Add a minimal headless validation (`exec.LookPath("gemini")` + non-fatal warning), or leave `CanRunHeadless() = false`?

---

## §6 — Risks & constraints

| # | Risk | Mitigation |
|---|---|---|
| R1 | Moving `library/commands/` breaks existing installs that rely on the old path | Keep a fallback read path (`library/gemini/commands/` preferred; fall back to `library/commands/` for one release). Delete fallback in a follow-up. |
| R2 | Gemini extension install flow changes upstream | Research was done in 2026; pin to a `gemini` version in the generator's docstring. Extension format is documented, unlikely to break fast. |
| R3 | MCP servers with `${VAR}` placeholders can't ship in extension settings without user-config prompts | Either (a) skip placeholder-bearing entries as in spec 014, or (b) map `${VAR}` → extension `settings` array with `envVar` matching. Option (b) is cleaner; keep as a stretch goal. |
| R4 | Tests that currently assume `library/commands/` needs updating | Low risk; the path is centralized in the Gemini adapter. |

---

## §7 — Out of scope

- Codex deep setup — separate spec (018) if approved.
- Gemini theme ship in the extension (Gemini extensions support `themes`). Skipped until a use case appears.
- Auto-publishing the extension to a registry (Gemini has no official marketplace yet as of research cutoff).
- Gemini `/init` slash command replication — already covered by `ai-setup init` writing GEMINI.md directly.
