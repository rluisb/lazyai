# Plan — 011: OpenCode Deep Setup

**Date:** 2026-04-19
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Work type:** Feature (full RPI)
**Research:** [research.md](./research.md)

---

## 1. Goal

Deliver 100% structural conformance between ai-setup's opencode install output and opencode's canonical layout at all three scopes (project, workspace, global). File-writes remain the primary mechanism; the `opencode` CLI is used for **validation** and **plugin installation** only.

## 2. Non-Goals (Out of Scope)

- Themes directory (`themes/`) — not currently needed.
- Refactoring other adapters (Claude, Gemini, Codex, Copilot).
- Replacing `StripFrontmatterAndInjectModel` globally; we add an opencode-specific path and leave other adapters untouched.
- CI-side validation with `opencode` binary (local dev only in spec 011; a CI follow-up may be added later).

## 3. Phased Implementation

Each phase leaves `go test ./... -count=1` green and the scaffold end-to-end working. Commit per phase.

### Phase 1 — Config foundation: `.jsonc` unification + instructions resolution

**Scope:** Consolidate install-time and compile-time config writes onto `opencode.jsonc`. Ensure `instructions` key resolves correctly at each scope.

**Changes:**
- `internal/adapter/opencode.go`:
  - Target `opencode.jsonc` (not `.json`) as the default config path.
  - On install, if `opencode.json` pre-exists: back it up (`.bak` sidecar), read its contents, write them to `opencode.jsonc`, delete `.json`. One-time migration.
  - `instructions` key: set to `["AGENTS.md"]` for project/workspace (relative to `.opencode/` → resolves to `.opencode/AGENTS.md` which we already install). For global, set to `["AGENTS.md"]` (resolves to `~/.config/opencode/AGENTS.md`).
  - Install `AGENTS.md` inside `.opencode/` root (already done via `InstallToolContextFiles` with `ToolDir: ocDir` — verify the root AGENTS.md is emitted).
- `internal/adapter/mcp_compiler.go#compileOpenCodeMCP`:
  - Change config target from hardcoded `opencode.jsonc` (already `.jsonc` — confirm) to a shared constant/helper.
- `internal/configmerge/`:
  - Ensure `MergeJSONFile` accepts `.jsonc` path and preserves JSONC comments on read (already via `jsonc.ReadJSONCFile`). Verify write path also supports JSONC comment preservation; if not, document that comments survive reads but are regenerated on writes.

**Tests (new / updated):**
- `internal/adapter/opencode_test.go`: install with pre-existing `opencode.json` → asserts `.jsonc` exists, `.json` removed, `.bak` sidecar present with original contents.
- `internal/adapter/scope_test.go`: no change (paths unchanged).
- `internal/adapter/mcp_compiler_test.go`: scope-parity test asserts `opencode.jsonc` at every scope (no `.json` ever).
- `internal/adapter/opencode_test.go`: assert `instructions: ["AGENTS.md"]` key in final config at all 3 scopes.

**Definition of done:** After install at any scope, exactly one opencode config file exists, named `opencode.jsonc`, at the correct root, with a valid `instructions` array pointing at the installed AGENTS.md.

---

### Phase 2 — Schema correctness: opencode frontmatter emitter + MCP deep-merge

**Scope:** Installed agents conform to opencode's frontmatter schema. MCP compile preserves user-authored servers.

**Changes:**
- `internal/adapter/opencode_frontmatter.go` (new):
  - `BuildOpenCodeAgentFrontmatter(src []byte, opts OpenCodeAgentOpts) []byte` — emits:
    ```yaml
    ---
    description: <from source or opts>
    mode: primary | subagent | all
    tools: { bash: true, read: true, ... }  # or comma string per opencode schema
    model: <provider/model>
    permission: { edit: ask, bash: ask }    # optional
    ---
    ```
  - Called from `opencode.go` Install in place of the shared `StripFrontmatterAndInjectModel` for opencode only.
- `internal/adapter/mcp_compiler.go#compileOpenCodeMCP`:
  - Replace `existingConfig["mcp"] = ocMcp` with a per-server merge:
    - Read existing `mcp` map (empty if absent).
    - For each ai-setup-managed server in `ocMcp`, upsert (overwrite by name).
    - Leave user-authored servers (those **not** in `ocMcp`) untouched.
  - Add a "managed" marker strategy: tag our upserts via the server config itself (no extra key) — we detect "managed" by matching against the current catalog snapshot; anything else is user-authored.

**Tests:**
- `internal/adapter/opencode_frontmatter_test.go` (new): table test for frontmatter output — description/mode/tools/model/permission combinations, verify YAML parses and keys match opencode schema.
- `internal/adapter/mcp_compiler_test.go`:
  - Test: pre-existing user-authored MCP server survives a second `CompileMCP` call.
  - Test: ai-setup-managed server can be toggled enabled→disabled via recatalog without dropping user entries.

**Definition of done:** A freshly scaffolded agent's frontmatter parses as a valid opencode agent definition. A second `ai-setup compile` preserves any MCP server the user added by hand.

---

### Phase 3 — Structural parity: Commands & Modes

**Scope:** Populate `.opencode/commands/` and `.opencode/modes/` from new library sources. Wizard adds opencode-specific selection steps (reusing spec 010 pattern).

**Changes:**
- `library/opencode/commands/` (new):
  - 2–4 starter commands (e.g., `review.md`, `test.md`, `commit.md`) with opencode frontmatter:
    ```yaml
    ---
    description: <short>
    agent: <optional>
    model: <optional>
    ---
    ```
- `library/opencode/modes/` (new):
  - 1–2 starter modes (e.g., `plan.md`, `audit.md`) with opencode mode frontmatter.
- `internal/library/embed.go` (or equivalent): extend embedded FS to include `library/opencode/`.
- `internal/adapter/opencode.go`:
  - New `CopyLibraryDirectory` calls for commands and modes, selection keys `"opencodeCommands"` and `"opencodeModes"`.
  - Destination: `<root>/commands/<name>.md` and `<root>/modes/<name>.md`.
- `tui/wizard/`:
  - Add two new selection steps (opencode commands, opencode modes) — behind the "custom" preset only (match spec 010 pattern).
- `internal/db/` store:
  - Persist `OpenCodeCommands` and `OpenCodeModes` selection slices (pattern follows `Commands` / `ChatModes` from spec 009 follow-up).

**Tests:**
- `internal/adapter/opencode_test.go`: assert commands and modes installed at every scope.
- `internal/library/integration_test.go`: assert `library/opencode/commands/` and `.../modes/` are embedded.
- `internal/db/store_test.go`: roundtrip for new selection fields.
- `tui/wizard/` tests if present: selection persistence.

**Definition of done:** `opencode debug config` (or direct `.opencode/` inspection) shows scaffolded commands and modes present; wizard lets the user opt in/out per-asset.

---

### Phase 4 — Post-install validation (opt-in, gated)

**Scope:** If `opencode` is on `PATH`, run validation probes after install and surface mismatches as warnings (non-fatal).

**Changes:**
- `internal/adapter/opencode_validate.go` (new):
  - `ValidateOpenCodeInstall(ctx) []ValidationWarning` — no-ops and returns nil if `exec.LookPath("opencode")` fails.
  - Runs `opencode debug config --json` (or `opencode debug config` and best-effort parses), asserts our expected mcp/agents/skills/commands/modes are visible.
  - Runs `opencode debug agent <name>` for each installed agent; captures parse errors.
  - Returns structured warnings (tool, item, reason).
- `internal/scaffold/scaffold.go`: after adapter install loop, if opencode was installed, call `ValidateOpenCodeInstall` and log warnings via the existing user-facing logger.

**Tests:**
- `internal/adapter/opencode_validate_test.go`: unit test with a fake exec runner producing canned stdout; asserts warning extraction logic.
- Skip / noop test when `opencode` binary is absent.

**Definition of done:** On a machine with opencode installed, running `ai-setup init` with opencode selected prints "✓ opencode config validated (N items)" or itemized warnings. On a machine without opencode, no change in output.

---

### Phase 5 — Plugin install flow (optional, gated)

**Scope:** Wizard offers a plugins selection; for each selected plugin, ai-setup shells out to `opencode plugin <module>` with `-g` at global scope, inside target dir otherwise.

**Changes:**
- `tui/wizard/`: new "OpenCode plugins" multi-select step, only shown if (a) opencode is selected and (b) binary on PATH. Populated from a hardcoded curated list in `library/opencode/plugins.json` (new).
- `internal/adapter/opencode.go`: post-install step — iterate selected plugins, run `opencode plugin <module>` with appropriate `-g` flag and working directory. Capture stdout/stderr; surface failures as non-fatal warnings.
- `internal/db/` store: persist `OpenCodePlugins` selection.

**Tests:**
- Unit: plugin exec is mocked via a `PluginRunner` interface — tests verify correct flags (`-g` at global, not at project/workspace; correct `cwd`).
- Skip if opencode absent.

**Definition of done:** User can pick plugins in the wizard; ai-setup installs them via the opencode CLI at the correct scope.

---

## 4. Task Breakdown

| # | Task | Phase | Est. |
|---|------|-------|------|
| 001 | Unify opencode config on `.jsonc` + migration | 1 | S |
| 002 | Verify/fix `instructions` key resolution at all 3 scopes | 1 | S |
| 003 | New opencode agent frontmatter emitter | 2 | M |
| 004 | MCP per-server deep-merge on compile | 2 | M |
| 005 | Add `library/opencode/commands/` and `modes/` assets | 3 | S |
| 006 | Adapter wiring for commands + modes | 3 | M |
| 007 | Wizard selection steps + store persistence for commands/modes | 3 | M |
| 008 | Post-install validation via `opencode debug *` | 4 | M |
| 009 | Plugin install flow via `opencode plugin <module>` | 5 | L |

Task files live in `specs/011-opencode-deep-setup/tasks/NNN-name.md` — created alongside this plan.

## 5. Acceptance Criteria

See [checklists/requirements.md](./checklists/requirements.md) — canonical list.

Highlights:
- **AC-1**: Install at each scope produces exactly one `opencode.jsonc` at the correct root.
- **AC-2**: `instructions` key resolves to an existing AGENTS.md at each scope.
- **AC-3**: Every installed agent has opencode-schema-valid frontmatter.
- **AC-4**: MCP compile preserves user-authored servers.
- **AC-5**: Commands and modes are installed at all 3 scopes and selectable in wizard.
- **AC-6**: With `opencode` on PATH, post-install validation runs and produces zero warnings on a clean install.
- **AC-7**: With `opencode` on PATH and plugins selected, plugins are installed at the correct scope.

## 6. Verification Protocol

Per CLAUDE.md, this is a **complex task** → 3 rounds of verification per phase completion:

1. **Round 1:** Requirements check against §5 AC list + `go test ./... -count=1` + `go vet ./...`.
2. **Round 2:** Edge cases — pre-existing `opencode.json`, hand-authored MCP entries, scope misroutes, wizard cancellation mid-step.
3. **Round 3:** Integration boundaries — scope-parity (all 3 × all assets), `opencode debug config` roundtrip on a dev box with opencode installed.

## 7. Risks & Unknowns

| Risk | Mitigation |
|---|---|
| `opencode` frontmatter schema drift (e.g., `tools` expected as map vs. comma string depending on version) | Phase 2: write schema emitter against v1.4.9 (local install); add compatibility note; Phase 4 validator will catch drift as it happens |
| Shell-out to `opencode plugin` introduces a hard runtime dep for Phase 5 | Gate strictly on `exec.LookPath`; skip step cleanly if absent |
| JSONC comment preservation on write is partial (comments round-trip on read but regenerate on write) | Document clearly; acceptable tradeoff; only impacts users who hand-edit comments |
| "Managed" vs "user-authored" MCP detection (Phase 2) has no explicit marker | Base the decision on the in-memory catalog at compile time: anything we would emit is ours, anything else stays. Documented limit: if user mirrors a managed name with a custom config, we will overwrite — surface as a future follow-up |

## 8. Rollback Plan

Every phase is a separate commit; each commit leaves the codebase working. If a phase breaks in verification, `git revert` the phase commit. No database migrations in this spec, so rollback is code-only.

## 9. Knowledge Map Updates (at completion)

- `specs/KNOWLEDGE_MAP.md`: add row for spec 011; update Packages Reference with new files.
- `CLAUDE.md` Codebase Map: add `library/opencode/` line if adopted.
- `specs/standards/testing/`: no new pattern (uses existing scope-parity + store roundtrip patterns).
- `specs/standards/coding/`: consider adding an "opencode frontmatter schema" note if reusable elsewhere.
