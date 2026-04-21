# Spec 015: Claude Code `settings.local.json` — Research

**Date:** 2026-04-20
**Status:** Research — awaiting human gate

## §1 — The deferred item

From `specs/KNOWLEDGE_MAP.md` Pending/Follow-up:
> `settings.local.json` coverage for Claude Code (deferred from spec 012; user secrets, local-only config)

Explicitly deferred as **Q5 in spec 012** on 2026-04-19:
> "Do not write or stub in 012. Add a line in Pending/Follow-up."

## §2 — Upstream reality check

From https://code.claude.com/docs/en/settings:

### Scope precedence (highest → lowest)
1. Managed (IT/MDM) — can't be overridden
2. CLI arguments
3. **Local** — `.claude/settings.local.json` (gitignored)
4. Project — `.claude/settings.json` (committed)
5. User — `~/.claude/settings.json`

### Key properties of `settings.local.json`
| Property | Value |
|---|---|
| Path | `<repo>/.claude/settings.local.json` |
| Gitignore | **Auto-added by Claude CLI** when the file is created |
| Scope | Per-user, per-repo ("You, in this repository only") |
| Schema | Same as `settings.json` |
| Managed by | User (manually) or `claude` CLI (`/config` → Local tab) |
| Use cases | Machine-specific permissions, dev-only hooks, secrets, experimental config |

### What ai-setup currently does for Claude Code
- Writes `.claude/settings.json` via `configmerge.MergeJSONFile` (deep-merge, backup-on-first-touch)
- At global scope, writes `~/.claude/settings.json`
- `.mcp.json` handled separately (project-scope only, committed)
- **Never touches `settings.local.json`**

## §3 — Use case analysis

The knowledge map says "user secrets, local-only config". Three concrete user-facing needs:

| Need | Today | With settings.local.json support |
|---|---|---|
| N1. Commit MCP server catalog without leaking API keys | User puts `GITHUB_PAT` literal in `.mcp.json`; risks commit | Split: non-secret server metadata in `.mcp.json`; secret env in `settings.local.json` |
| N2. Per-developer permission overrides | User edits `.claude/settings.json`, must avoid committing | Edit `settings.local.json`; gitignore handles it |
| N3. Experimental hooks/tool allowlist | Same risk | Same mitigation |

**Finding:** The primary demand is **N1** — safely separating secret env vars from committed MCP config. N2/N3 are bonus.

## §4 — Current gap in ai-setup

Today, if a user wants to keep `GITHUB_PAT` out of git:
1. They manually create `.claude/settings.local.json` with `{ "env": { "GITHUB_PAT": "..." } }`.
2. ai-setup does not know about this file.
3. On next `ai-setup init`, `.mcp.json` might get overwritten with a fresh copy that re-introduces an env-var placeholder.

No direct ai-setup damage (it doesn't touch `settings.local.json`), but there is a **missing affordance**: users have no guided way to route secrets into settings.local.json.

## §5 — Options

| Option | Description | Pros | Cons |
|---|---|---|---|
| **A. Docs-only** | Add docs describing how users should manually populate settings.local.json | Zero risk, zero new code | No automation; doesn't address N1 mechanically |
| **B. Opt-in flag: `--local-secrets`** | New flag; when passed, MCP compile emits **all** server env vars to settings.local.json, leaves server metadata in `.mcp.json` | Automated secret split; user-controlled | Meaningful code; must update MCP compile path |
| **C. Per-server opt-in via catalog** | Add `local: true` marker in `library/mcp/catalog.json`; those servers' env vars route to settings.local.json | Fine-grained; lets library authors mark secret-bearing servers | More complex; still need compile path changes |
| **D. Always-split** | By default, ship any env var whose value matches `${VAR}` placeholder pattern into settings.local.json as `{env: {VAR: ""}}` stub | No flag; automatic | Surprising; may overwrite user's existing local config |

**Recommendation: B** as the MVP, with a follow-up path to C if needed. B provides the deterministic behavior the user asked for, keeps the default path unchanged (low regression risk), and slots cleanly into the existing `--drive-cli`-style flag vocabulary.

## §6 — Decision interview (need human answers)

1. **Q1** — Primary use case: secret env routing (N1) only, or also N2/N3?
2. **Q2** — Pick Option B (opt-in flag), C (catalog marker), or something else?
3. **Q3** — Scope of `--local-secrets`: project scope only, or also workspace/global (note: `~/.claude/settings.json` already exists for global user scope)?
4. **Q4** — When `--local-secrets` is set, what exactly moves to settings.local.json:
   - **Just env vars** with placeholder values (`${FOO}`), or
   - **Entire servers** flagged as secret-bearing, or
   - **All servers**?
5. **Q5** — Ensure `.claude/settings.local.json` is listed in `.gitignore` when ai-setup writes it? (Claude CLI does this automatically on its own writes; we'd only need to do it if ai-setup creates the file first.)
6. **Q6** — Merge strategy on re-run: deep-merge with backup-on-first-touch (consistent with settings.json), or last-write-wins?

## §7 — Risks & constraints

- **R1:** Overwriting user's existing settings.local.json. Must deep-merge; never truncate.
- **R2:** Managed-settings can override our writes; we must not **assume** our values take effect.
- **R3:** If `.claude/settings.local.json` and `.mcp.json` both list the same server, precedence rules determine effective config. Document the split clearly.
- **R4:** Claude CLI auto-ignores settings.local.json. If ai-setup writes first, gitignore must be updated by ai-setup (not Claude CLI) to maintain the invariant.

## §8 — Out of scope

- Claude Code plugin marketplace integration (parked per spec 012 follow-up)
- MDM / managed-settings.json handling
- CLAUDE.local.md (per-user project memory) — analogous but separate file, not under settings
