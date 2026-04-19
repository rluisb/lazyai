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
| 010 | Wizard selection UI + Codex drive-cli + CLAUDE.md hybrid fill | ✅ Complete | `feature/go-migration` (1ee3e9f) |

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
| `internal/scaffold/root.go#fillClaudeMdPlaceholders` | Hybrid template-placeholder substitution (mechanical auto-infer + subjective fill-in markers) |

## Pending / Follow-up

- [x] ~~Codex `config.toml` `[mcp_servers.*]` enrichment via `CompileMCP`~~ — spec 009
- [x] ~~`AGENTS.override.md` for Codex global install~~ — spec 008 follow-up
- [x] ~~`--drive-cli` flag for Gemini~~ — spec 008 follow-up
- [x] ~~Claude Code `--drive-cli`~~ — spec 009
- [x] ~~Codex `--drive-cli`~~ — spec 010
- [x] ~~Compile-time scope awareness~~ — spec 009
- [x] ~~Gemini custom commands + Copilot chatmodes~~ — spec 009
- [x] ~~Wizard UI for commands/chatmodes selection~~ — spec 010
- [x] ~~CLAUDE.md hybrid placeholder fill (mechanical + org/team)~~ — spec 010
- [x] ~~Spec-dir convention reconciliation~~ — spec 008 follow-up
- [x] ~~Store persistence for Commands/ChatModes~~ — spec 009 follow-up patch
- [ ] Snapshot tests for library assets + compiled output (deferred in spec 009)
- [ ] `--drive-cli` for OpenCode (interactive-only upstream) / Copilot (flag surface unverified)
