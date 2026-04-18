# Knowledge Map

## Feature Specs

| # | Name | Status | Branch / PR |
|---|------|--------|-------------|
| 001 | Store and errors | ✅ Complete | merged |
| 002 | Simplification and restructure | ✅ Complete | merged |
| 003 | Post-install automation and integrations | ✅ Complete | merged |
| 004 | Go migration | ✅ Complete | merged |
| 005 | Setup flow fixes | ✅ Complete | merged |
| 006 | Housekeeping, memory, and bootstrap | ✅ Complete | merged |
| 007 | Wizard step-by-step UX | ✅ Complete | `feature/go-migration` (f6733ed) |
| 008 | CLI tool structure parity | ✅ Complete | `feature/go-migration` (0747e50) |
| 009 | Compile-time scope awareness & artifact parity | ✅ Complete | `feature/go-migration` (1dca890) |

## Key Architecture Decisions

| Decision | ADR | Rationale |
|----------|-----|-----------|
| Scope resolver as single source of truth for tool paths | — | Eliminates per-adapter `isGlobal` branching; easy to extend for new tools |
| Deep-merge with backup-on-first-touch for config files | — | Preserves user-authored keys across re-runs; one `.bak` sidecar only |
| Copilot unsupported at global scope | — | No upstream concept for global Copilot config |
| Codex split root: `.codex/` for config, `.agents/skills/` for skills | — | Matches upstream Codex CLI conventions |
| Workspace scope = project-shaped layout at user-selected dir | — | No tool-native workspace concept; direct-write is universal |
| `CompileContext` struct carries scope info to compile-time adapters | — | Breaks `CompileMCP(targetDir, records)` signature for all 5 adapters; clean internal migration |
| Claude Code × global compile skips `.mcp.json`; init's settings.json merge handles it | — | `.mcp.json` is a user-committed project-scope file; global mcpServers live in settings.json |

## Packages Reference

| Package | Purpose |
|---------|---------|
| `internal/adapter/scope.go` | `ResolveToolRoot`, `ResolveCodexRoots`, `IsScopeSupported`, `ErrScopeUnsupported` |
| `internal/adapter/mcp_compiler.go` | `CompileMCPForTool`, per-tool compile functions (scope-aware via `CompileContext`) |
| `internal/configmerge/` | `MergeJSONFile`, `MergeTOMLFile` — deep-merge with `.bak` |
| `internal/globalpaths/` | Home-dir roots for all tools; `ResolveCodexSkillsGlobalDir` |
| `internal/scaffold/root.go` | Single emitter for memory docs (`memoryDocDestPath`) |
| `library/commands/*.toml` | Gemini custom slash command templates |
| `library/chatmodes/*.chatmode.md` | Copilot chat mode templates |

## Pending / Follow-up

- [x] ~~Codex `config.toml` `[mcp_servers.*]` enrichment via `CompileMCP`~~ — done in spec 009
- [x] ~~`AGENTS.override.md` for Codex global install~~ — done in spec 008 follow-up
- [x] ~~`--drive-cli` flag for Gemini~~ — done in spec 008 follow-up
- [x] ~~Claude Code `--drive-cli`~~ — done in spec 009
- [x] ~~Compile-time scope awareness~~ — done in spec 009
- [x] ~~Gemini custom commands + Copilot chatmodes~~ — done in spec 009
- [x] ~~Spec-dir convention reconciliation~~ — done in spec 008 follow-up
- [ ] Snapshot tests for library assets + compiled output (deferred in spec 009)
- [ ] Wizard UI for commands/chatmodes selection (defer to spec 010)
- [ ] `--drive-cli` for OpenCode/Codex/Copilot (no viable upstream CLI yet)
