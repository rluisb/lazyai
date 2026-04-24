# Spec 019 — OpenCode-Only Consolidation

## Status
In Progress

## Goal

Remove all non-OpenCode providers (Claude Code, Gemini, Copilot, Codex, PI) from both
the TypeScript and Go implementations. Consolidate both runtimes so they implement the
exact same install steps, scope handling, and file layout for OpenCode.

## Motivation

- The project is pivoting to support only OpenCode as the target AI coding tool.
- TypeScript has the correct conceptual flows for global/workspace/project; Go has
  richer implementations (commands, modes, plugins) but a config-placement bug.
- Maintaining five adapters creates drift and test surface with no benefit.

## Scope

| Area | Action |
|------|--------|
| `internal/adapter/claudecode*.go` | Remove |
| `internal/adapter/codex*.go` | Remove |
| `internal/adapter/copilot*.go` | Remove |
| `internal/adapter/gemini*.go` | Remove |
| `internal/adapter/registry.go` | Keep only OpenCode |
| `internal/adapter/scope.go` | Simplify (remove non-OpenCode cases) |
| `internal/adapter/mcp_compiler.go` | Remove non-OpenCode tool branches |
| `internal/globalpaths/globalpaths.go` | Remove non-OpenCode tool paths |
| `internal/types/types.go` | Remove non-OpenCode ToolIds; keep OpenCode-only constants |
| `src/adapters/claude-code.ts` | Remove |
| `src/adapters/codex.ts` | Remove |
| `src/adapters/copilot.ts` | Remove |
| `src/adapters/gemini.ts` | Remove |
| `src/adapters/pi.ts` | Remove |
| `src/adapters/registry.ts` | Keep only OpenCode |
| `src/types.ts` | ToolId = `'opencode'` only |
| `src/wizard/planner.ts` | Remove non-OpenCode entries from ADAPTER_PATHS |
| `src/wizard/phase1-context.ts` | Remove tool selection (opencode only) |
| `library/claudecode/`, `library/copilot/`, `library/gemini/`, `library/codex/` | Remove |
| `src/adapters/opencode.ts` | Add modes dir, commands/modes copy, plugin install, `opencode.jsonc` |
| `internal/adapter/opencode.go` | Fix config placement (project root, not `.opencode/`) |

## Alignment: TS vs Go — OpenCode Install Steps (final state)

Both runtimes must execute these steps in order:

### project / workspace scope

1. Resolve `ocDir = <targetDir>/.opencode`
2. Create dirs: `ocDir/agents`, `ocDir/skills`, `ocDir/commands`, `ocDir/modes`
3. Write `<targetDir>/opencode.jsonc` (project root) if absent — default config
4. Copy agents from `library/agents/` → `ocDir/agents/*.md` (with opencode frontmatter transform)
5. Copy orchestrator agent if enabled → `ocDir/agents/orchestrator.md` (mode=primary)
6. Copy skills from `library/skills/` → `ocDir/skills/<name>/SKILL.md`
7. Copy commands from `library/opencode/commands/` → `ocDir/commands/*.md`
8. Copy modes from `library/opencode/modes/` → `ocDir/modes/*.md`
9. Install context files: `AGENTS.md` in `ocDir/`, `ocDir/agents/`, `ocDir/skills/`
10. Install plugins via `opencode plugin <module>` (if selections present)

### global scope

1. Resolve `ocDir = ~/.config/opencode`
2. Create dirs: `ocDir/agents`, `ocDir/skills`, `ocDir/commands`, `ocDir/modes`
3. Write `ocDir/opencode.jsonc` (global config at `~/.config/opencode/opencode.jsonc`) if absent
4. Steps 4–10 same as above, rooted at `ocDir`

## Implementation Phases

### Phase 1 — Fix Go config placement bug
- In `internal/adapter/opencode.go`: for project/workspace scope, write config at
  `<targetDir>/opencode.jsonc` (project root), not `ocDir/opencode.jsonc`.
- For global scope: `ocDir = ~/.config/opencode`, so `ocDir/opencode.jsonc` is correct.

### Phase 2 — Align TS OpenCode adapter
- Add `modes` dir creation
- Add commands copy from `library/opencode/commands/`
- Add modes copy from `library/opencode/modes/`
- Change config filename from `opencode.json` to `opencode.jsonc`
- Add plugin install via `opencode` binary

### Phase 3 — Remove non-OpenCode Go adapters
- Delete: `claudecode*.go`, `codex*.go`, `copilot*.go`, `gemini*.go` (and test files)
- Update `registry.go`, `scope.go`, `globalpaths.go`, `types.go`, `mcp_compiler.go`

### Phase 4 — Remove non-OpenCode TS adapters
- Delete: `claude-code.ts`, `codex.ts`, `copilot.ts`, `gemini.ts`, `pi.ts`
- Update `registry.ts`, `types.ts`, `wizard/planner.ts`, `wizard/phase1-context.ts`

### Phase 5 — Remove non-OpenCode library assets
- Delete: `library/claudecode/`, `library/copilot/`, `library/gemini/`, `library/codex/`

### Phase 6 — Verify
- `go build ./...` passes
- `go test ./... -count=1` passes
- `npm run build` or `tsc --noEmit` passes

## Acceptance Criteria

- [ ] `go build ./...` succeeds with no references to removed tools
- [ ] `go test ./... -count=1` passes for all remaining tests
- [ ] TS compiles clean (`tsc --noEmit` or equivalent)
- [ ] Manual: `go run . init` for project scope installs into `.opencode/`
- [ ] Manual: `go run . init --scope global` installs into `~/.config/opencode/`
- [ ] Config file at `opencode.jsonc` at project root (not inside `.opencode/`)
- [ ] Commands in `ocDir/commands/`, modes in `ocDir/modes/`
