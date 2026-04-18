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

## Key Architecture Decisions

| Decision | ADR | Rationale |
|----------|-----|-----------|
| Scope resolver as single source of truth for tool paths | — | Eliminates per-adapter `isGlobal` branching; easy to extend for new tools |
| Deep-merge with backup-on-first-touch for config files | — | Preserves user-authored keys across re-runs; one `.bak` sidecar only |
| Copilot unsupported at global scope | — | No upstream concept for global Copilot config |
| Codex split root: `.codex/` for config, `.agents/skills/` for skills | — | Matches upstream Codex CLI conventions |
| Workspace scope = project-shaped layout at user-selected dir | — | No tool-native workspace concept; direct-write is universal |

## Packages Reference

| Package | Purpose |
|---------|---------|
| `internal/adapter/scope.go` | `ResolveToolRoot`, `ResolveCodexRoots`, `IsScopeSupported`, `ErrScopeUnsupported` |
| `internal/configmerge/` | `MergeJSONFile`, `MergeTOMLFile` — deep-merge with `.bak` |
| `internal/globalpaths/` | Home-dir roots for all tools; `ResolveCodexSkillsGlobalDir` |
| `internal/scaffold/root.go` | Single emitter for memory docs (`memoryDocDestPath`) |

## Pending / Follow-up

- [ ] Codex `config.toml` `[mcp_servers.*]` enrichment via `CompileMCP`
- [ ] `AGENTS.override.md` for Codex global install
- [ ] `--drive-cli` flag for Gemini (call `gemini mcp add` when binary present)
- [ ] Spec-dir convention reconciliation (flat `specs/NNN-*` vs typed `specs/features/NNN-*`)
