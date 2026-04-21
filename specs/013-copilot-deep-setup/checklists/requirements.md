# Requirements Checklist — 013: Copilot Deep Setup

Verify after each phase. Every box must be checkable before spec is marked complete.

## Phase 1 — Library assets
- [ ] `library/copilot/agents/{planner,builder,scout,reviewer,orchestrator}.agent.yaml` exist
- [ ] `library/copilot/instructions/{typescript,go,tests}.instructions.md` exist
- [ ] All YAML parses; all instruction files have non-empty `applyTo`
- [ ] Schema unit test green (`internal/library/copilot_schema_test.go`)

## Phase 2 — Project/Workspace emission
- [ ] `.github/agents/*.agent.yaml` emitted at project + workspace scope for selected agents
- [ ] `.github/instructions/*.instructions.md` emitted at project + workspace scope
- [ ] Orchestrator gated on `EnableServers`
- [ ] Skills emit as `.agent.yaml` (not `.prompt.md`); previously-ai-setup-owned `skill.prompt.md` cleaned up on re-run
- [ ] `library/prompts/*` still emits as `.prompt.md` (unchanged)
- [ ] `.vscode/mcp.json` scaffolds `inputs` when `${VAR}` placeholders present; omits key otherwise; dedupes

## Phase 3 — Scope lift + probe
- [ ] `IsScopeSupported(ToolIdCopilot, SetupScopeGlobal) == true`
- [ ] `globalpaths.ResolveGlobalToolTargetDir(ToolIdCopilot, home) == <home>/.copilot`
- [ ] `LookupCopilotBinary` + `CopilotHomePresent` helpers exist and unit-tested
- [ ] Global-scope install no-ops with single warning when both probe signals absent

## Phase 4 — Global emitters
- [ ] `~/.copilot/agents/*.agent.yaml` emitted when probe passes; content byte-identical to `.github/agents/` for same library source
- [ ] `~/.copilot/copilot-instructions.md` emitted on first install; untouched on re-run
- [ ] `~/.copilot/mcp-config.json` written via deep-merge; user-authored keys preserved; managed servers updated
- [ ] `.bak` sidecar created exactly once

## Phase 5 — Validation
- [ ] `CanRunHeadless` returns true iff `copilot` on PATH
- [ ] Per-agent smoke with 5s timeout and injectable runner
- [ ] Non-zero exit produces warning only; install still reports success

## Phase 6 — Tests + docs
- [ ] Scope-parity matrix includes (copilot, global) — both probe-pass and probe-fail paths
- [ ] Frontmatter schema tests parse every emitted artifact
- [ ] MCP round-trip integration test covers user-scope deep-merge + inputs scaffold
- [ ] `specs/KNOWLEDGE_MAP.md` updated (spec row, decisions, packages, pending)
- [ ] `go test ./... -count=1` green
- [ ] `go vet ./...` clean

## Cross-cutting
- [ ] No changes to Claude, OpenCode, Gemini, or Codex adapter code paths
- [ ] `.chatmode.md` emission behavior unchanged
- [ ] `library/prompts/*.prompt.md` emission unchanged
- [ ] Workspace scope remains project-identical for Copilot
