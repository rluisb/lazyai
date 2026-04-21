# Plan — Spec 010: Wizard UX for Commands/ChatModes, Codex DriveCLI, CLAUDE.md Hybrid Fill

**Date:** 2026-04-18
**Depends on:** Spec 009 merged
**Locked decisions:**
- Wizard UI for commands/chatmodes shown only when preset = custom
- `--drive-cli` added for **Codex only** (OpenCode is interactive-only; Copilot docs unverified)
- CLAUDE.md fill: **Option D (hybrid)** — auto-infer mechanical fields from scaffold context; one optional wizard step prompts for `org` + `team`; everything else stays `<!-- fill-in -->`

---

## Context

Three loose ends from the spec 009 research remain:

1. **Commands/ChatModes are auto-selected** by preset — `custom` preset has no UI to pick which to install.
2. **Codex CLI has a flag-driven `codex mcp add`** that also handles OAuth discovery — `ai-setup --drive-cli` should use it (currently only Gemini + Claude Code benefit).
3. **Scaffold-generated CLAUDE.md** ships with 10+ `[YOUR_*]` placeholders — `project_name` and `description` already substitute, but `org`, `team`, tech-stack, and commands remain as literal placeholders for users to fill manually. Auto-inferring the mechanical subset plus a brief optional "fill it now?" step for `org`/`team` eliminates the awkward placeholder-filled first commit.

---

## Acceptance Criteria

| # | AC | Verified by |
|---|---|---|
| AC-1 | When preset = custom, wizard phase 2 shows a "Commands" step (multi-select) and a "Chatmodes" step (multi-select). Both skipped when preset ≠ custom. | `phase2_test.go` |
| AC-2 | `Phase2Result.Commands` + `Phase2Result.ChatModes` flow through `buildScaffoldContext` into `ScaffoldContext.Commands/ChatModes`, overriding the preset-default `ALL_*` slices. | `phase2_test.go` + `helpers_test.go` |
| AC-3 | When `ctx.DriveCLI=true` and `codex` binary is on PATH, Codex adapter calls `codex mcp add <name> [--env K=V] -- <cmd> <args...>` for each enabled server. Falls back silently to direct-write TOML on any failure. | `codex_drivecli_test.go` (stub binary + missing-binary cases) |
| AC-4 | `library/root/CLAUDE.template.md` (and `AGENTS.template.md`) have mechanical fields (`PROJECT_NAME`, `PROJECT_DESCRIPTION`, `TECH_STACK`, `DEV/TEST/BUILD_COMMAND`) replaced at scaffold time based on `ScaffoldContext` (project name, primary language, framework). | `scaffold/root_test.go` extension |
| AC-5 | A new phase 1 sub-step "Project identity (optional)" asks for `Organization` and `Team`. Skippable (enter to skip → stays as `<!-- fill-in -->`). Non-interactive mode never prompts. | `phase1_test.go` |
| AC-6 | Subjective fields (`[YOUR_RULE_*]`, `[YOUR_ARCHITECTURE_NOTES]`, etc.) become `<!-- fill-in -->` markers in the generated CLAUDE.md, not raw `[YOUR_*]`. | same as AC-4 |
| AC-7 | No regressions in existing specs 007/008/009 tests. | `go test ./... -count=1` |

---

## Approach

### Wave A — Wizard commands/chatmodes selection (preset = custom only)

**Types** — extend `phase2.go`:
```go
type Phase2Result struct {
    Preset    types.PresetLevel
    Features  *types.FeatureFlags
    GitConv   *types.GitConventions
    Commands  []types.CommandId   // populated only when Preset == Custom
    ChatModes []types.ChatModeId  // populated only when Preset == Custom
}
```

**Steps** — after the features step (custom only), add two conditional steps:
1. "Custom commands" multi-select populated from `types.ALL_COMMANDS`
2. "Chatmodes" multi-select populated from `types.ALL_CHATMODES`

Both default to all-selected (preserving today's behaviour for users who click through). Skip when preset ≠ custom (reuse the existing `previousPhase2Step/nextPhase2Step` conditional pattern at `phase2.go:479-496`).

**Consumer update** — `cmd/helpers.go:86-98`: when `result.Phase2.Commands` or `ChatModes` is non-nil AND preset == custom, use those values; otherwise fall back to `ALL_*` for non-minimal presets.

**Tests** — mirror `TestRunPhase2*` pattern in `phase2_test.go`.

### Wave B — Codex `--drive-cli`

Mirror the Gemini/Claude pattern in `internal/adapter/codex.go`:

```go
if ctx.DriveCLI {
    if ok := installCodexMCPViaCLI(ctx); ok {
        log.Println("[codex] MCP servers registered via CLI")
    }
}
```

`installCodexMCPViaCLI` — resolves `codex` binary, reads canonical `.ai/mcp.json`, iterates enabled servers, and for each calls:
```
codex mcp add <name> [--env K=V ...] -- <command> <arg1> <arg2> ...
```

Extra care — Codex's positional `COMMAND [ARGS...]` comes after a literal `--` separator (verified from source). Falls back silently to the existing direct-write TOML when binary missing or any exec fails.

**Test:** `codex_drivecli_test.go` with stub binary + missing-binary cases (mirror `claudecode_drivecli_test.go`).

### Wave C — CLAUDE.md hybrid fill

**C-1. Wizard step:** new optional "Project identity" question block in phase 1, after project name:
- `Organization` (text input, enter-to-skip)
- `Team` (text input, enter-to-skip)

In non-interactive mode, reads from flags (`--org`, `--team`) — empty by default.

**C-2. Extend `WizardConfig`:**
```go
CLIOrg  string
CLITeam string
```

**C-3. Extend `ScaffoldContext`** + `ScaffoldCompiledRootOptions` with `Organization`, `Team`, and mechanical derivations (`PrimaryLanguage` + `Framework` already exist).

**C-4. Template substitutions** in `scaffold/root.go`:

| Placeholder | Source | Fallback (if empty) |
|---|---|---|
| `[YOUR_PROJECT_NAME]` | `opts.ProjectName` | already handled |
| `[YOUR_PROJECT_DESCRIPTION]` | new `ScaffoldCompiledRootOptions.ProjectDescription` (optional), derived from primary language and scope | `"AI-assisted development project"` |
| `[YOUR_TECH_STACK]` | derived from `opts.PrimaryLanguage` + `opts.Framework` | `"<!-- fill-in: tech stack -->"` |
| `[YOUR_ORG]` | `opts.Organization` | `"<!-- fill-in: your org -->"` |
| `[YOUR_TEAM]` | `opts.Team` | `"<!-- fill-in: your team -->"` |
| `[YOUR_DEV_COMMAND]` etc. | table lookup per primary language (Go: `go run .`, Node: `npm run dev`, etc.) | `"<!-- fill-in: dev command -->"` |
| `[YOUR_RULE_*]`, `[YOUR_DO_NOT_*]`, `[YOUR_SESSION_CHECK]`, `[YOUR_ARCHITECTURE_NOTES]`, `[YOUR_CODE_STYLE]`, `[YOUR_NAMING_CONVENTIONS]`, `[YOUR_TESTING_STRATEGY]`, `[YOUR_GIT_WORKFLOW]` | **always** `"<!-- fill-in: <hint> -->"` | — |

New helper `fillClaudeMdPlaceholders(content, opts) string` in `scaffold/root.go`.

**C-5. Apply same substitutions to `AGENTS.template.md` and `GEMINI.template.md`** (they share the placeholder set).

---

## Critical Files

### Modified

| File | Change |
|---|---|
| `tui/wizard/phase1.go` | Add optional org/team step |
| `tui/wizard/phase2.go` | Add custom-only commands/chatmodes steps |
| `tui/wizard/wizard.go` | Extend `WizardConfig` with `CLIOrg`, `CLITeam`; extend `Phase2Result` |
| `cmd/init.go` | New `--org`, `--team` flags |
| `cmd/helpers.go` | Use Phase2Result.Commands/ChatModes when set; pass Org/Team into ScaffoldContext |
| `internal/scaffold/types.go` | Add `Organization`, `Team`, `ProjectDescription` |
| `internal/scaffold/root.go` | New `fillClaudeMdPlaceholders`; wire `Organization`/`Team`; new `ScaffoldCompiledRootOptions` fields |
| `internal/adapter/codex.go` | New `installCodexMCPViaCLI` + DriveCLI branch |
| `library/root/CLAUDE.template.md` | Codebase-map placeholders stay, but mechanical ones get filled at scaffold time (no template change needed if substitution logic handles it) |

### New

| File | Purpose |
|---|---|
| `internal/adapter/codex_drivecli_test.go` | Stub-binary codex test |
| `tui/wizard/phase1_identity_test.go` | Org/team prompt test (or inline in `phase1_test.go`) |
| `tui/wizard/phase2_customselect_test.go` | Commands/chatmodes selection test (or inline in `phase2_test.go`) |

### Not touched

- `types.SetupScope` enum
- `CompileMCP` interface (scope-aware already from 009)
- Library template files (substitution happens post-read, not via template rewrite)

---

## Implementation Waves

### Wave A — wizard selection (preset = custom)
1. Extend `Phase2Result` with `Commands`, `ChatModes`.
2. Add two conditional steps after features selection; mirror `appendPhase2BackOption` pattern.
3. Update `nextPhase2Step`/`previousPhase2Step`.
4. Extend `cmd/helpers.go` to prefer Phase2Result values when custom preset.
5. Test with `TestRunPhase2Custom_SelectsCommandsAndChatModes` + non-interactive pass-through test.

### Wave B — Codex `--drive-cli`
1. Add `installCodexMCPViaCLI` helper in `codex.go`.
2. Wire behind `ctx.DriveCLI` in `Install`.
3. Test with stub + missing-binary cases.

### Wave C — CLAUDE.md hybrid fill
1. Extend `WizardConfig` + `--org`/`--team` flags.
2. Extend `ScaffoldContext` + `ScaffoldCompiledRootOptions` with `Organization`, `Team`, `ProjectDescription`.
3. New `fillClaudeMdPlaceholders(content, opts)` helper in `scaffold/root.go`.
4. Wire helper into root compilation path (replaces scattered `strings.ReplaceAll` calls).
5. Add table test covering every placeholder.

### Dependency graph

```
A, B, C — all independent; can merge in any order.
Recommended sequence: A → B → C (easiest → highest blast radius).
```

---

## Verification

```bash
go vet ./...
go test ./internal/adapter/... ./tui/wizard/... ./internal/scaffold/... ./cmd/... -count=1 -v
go test ./... -count=1
```

**Manual smoke:**
1. `ai-setup init --scope project --preset custom` → interactive wizard shows Commands + Chatmodes steps after features.
2. `ai-setup init --scope project --preset minimal` → Commands + Chatmodes steps do NOT appear.
3. `ai-setup init --drive-cli --tools codex --non-interactive` with stub `codex` binary → stub receives `mcp add ...`.
4. `ai-setup init --org "Acme" --team "Platform" --non-interactive` → generated `CLAUDE.md` contains `**Organization:** Acme` + `**Team:** Platform`; no `<!-- fill-in -->` for those two fields.
5. `ai-setup init --non-interactive --primary-language Go` → `**Tech Stack:** Go` in CLAUDE.md, not `[YOUR_TECH_STACK]`.

---

## Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | Phase 2 step reordering regresses spec 007 | Medium | Extend conditional stepping, don't rewrite; tests catch ordering drift |
| R-2 | Codex `mcp add` flag format changes upstream | Low | Test with stub binary; real Codex invocation is a smoke-only check |
| R-3 | `<!-- fill-in -->` markers in HTML-rendering previews show up literally | Low | HTML comments are standard; acceptable |
| R-4 | Non-interactive default for commands/chatmodes silently "all" | Low | Preserves today's behaviour; matches agents/skills defaults |
| R-5 | Wizard phase 1 additional step increases setup friction | Medium | Hard-cap at 2 questions (org, team); both skippable with enter |
| R-6 | Tech-stack inference is wrong for mixed-language projects | Medium | Only derive when `PrimaryLanguage` set; otherwise leave `<!-- fill-in -->` |

---

## Out of Scope

- OpenCode `--drive-cli` (interactive-only `mcp add`; direct-write is strictly better)
- Copilot `--drive-cli` (CLI flag surface unverifiable)
- Interactive prompts for subjective CLAUDE.md fields beyond org/team
- Auto-inference of rules/architecture notes (not achievable from scaffold context)
- Rename of `types.ALL_COMMANDS` / `ALL_CHATMODES` presets

---

⛔ **HUMAN GATE — approve plan before implementing.**
