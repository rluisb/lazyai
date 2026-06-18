# Knowledge Map

## Feature Specs

| # | Name | Status | Branch / PR |
|---|------|--------|-------------|
| 001 | AI techniques integration (W1.A) | ✅ Complete | merged |
| 002 | Simplification and restructure | ✅ Complete | merged |
| 003 | Post-install automation and integrations | ✅ Complete | merged |
| 004 | Go migration | ✅ Complete | merged |
| 005 | Setup flow fixes | ✅ Complete | merged |
| 006 | Housekeeping, memory, and bootstrap | ✅ Complete | merged |
| 007 | Wizard step-by-step UX | ✅ Complete | `feature/go-migration` (f6733ed) |
| 008 | CLI tool structure parity | ✅ Complete | `feature/go-migration` (0747e50) |
| 009 | Compile-time scope awareness & artifact parity | ✅ Complete | `feature/go-migration` (1dca890) |
| 010 | Wizard selection UI + Codex drive-cli + CLAUDE.md hybrid fill | ✅ Complete | `feature/go-migration` (1ee3e9f) |
| 011 | OpenCode deep setup (config, frontmatter, MCP merge, commands/modes, validation, plugins) | ✅ Complete | `feature/go-migration` (556db86) |
| 012 | Claude Code deep setup (global/project/workspace agents, rules, commands, output-styles) | ✅ Complete | `feature/go-migration` |
| 013 | GitHub Copilot deep setup (agents, instructions, chatmodes, MCP, global probe gating) | ✅ Complete | `feature/go-migration` (0b26dd1) |
| 014 | Copilot global MCP compile (`~/.copilot/mcp-config.json` + VS Code mcp.json split) | ✅ Complete | `feature/go-migration` |
| 015 | Claude Code `--local-secrets` flag routing MCP to `.claude/settings.local.json` | ✅ Complete | `feature/go-migration` |
| 016 | `ai-setup build-plugin` — generate Claude Code plugin from library (agents + skills + commands + output styles) | ✅ Complete | `feature/go-migration` |
| 017 | Gemini deep setup — `library/gemini/` restructure + `ai-setup build-gemini-extension` generator + LookPath validation | ✅ Complete | `feature/go-migration` |
| 018 | Codex deep setup — `--skip-git-repo-check` validation fix + `library/codex/` AGENTS.override template + `codex mcp list` post-install summary | ✅ Complete | `feature/go-migration` |
| 019 | Orchestrator Go runtime + ai-setup integration — Go binary `ai-setup-orchestrator` replaces `npx -y @ai-setup/orchestrator`; release assets/download/cache support implemented; A2A deferred/opt-in | ✅ Complete | `feature/orchestrator-a2a-rewrite` |
| 020 | Go/TS setup parity audit and alignment | ✅ Complete | archived top-level spec |
| 021 | Parity verification and gap report | ✅ Complete | archived top-level spec |
| 022 | Speckit workflow alignment | ✅ Complete | archived top-level spec; workspace-root follow-up noted in tasks |
| 023 | Repository cleanup — local hazard cleanup, legacy package audit, and spec hygiene proposal | ✅ Complete | `feature/repo-cleanup` |
| 024 | LazyAI Go-only packages — repo identity `github.com/rluisb/lazyai`, packages `cli`/`orchestrator`/`diffviewer`, binaries `lazyai-*`, npm/npx removed | ✅ Final verification | `feature/lazyai-go-only-plan` |

## Standards

| Document | Path | Status | Description |
|----------|------|--------|-------------|
| Reversa Confidence Scale | `specs/standards/confidence-scale.md` | Active | 🟢🟡🔴 confidence scale for `/populate` and AI-inferred content — meaning, evidence rules, write behavior, and human-escalation criteria |

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
| OpenCode config unified on `opencode.jsonc`; MCP compile preserves user servers via deep-merge | — | Prevents clobbering user-authored `mcp.servers` on re-run; managed entries win on key collision |
| OpenCode CLI used only for validation (`opencode debug *`) and plugin install, not file-writing | — | CLI is interactive-only for most operations; direct-write gives deterministic output |

## Packages Reference

| Package | Purpose |
|---------|---------|
| `packages/cli/internal/adapter/scope.go` | `ResolveToolRoot`, `ResolveCodexRoots`, `IsScopeSupported`, `ErrScopeUnsupported` |
| `packages/cli/internal/adapter/mcp_compiler.go` | `CompileMCPForTool`, per-tool compile functions (scope-aware via `CompileContext`) |
| `packages/cli/internal/configmerge/` | `MergeJSONFile`, `MergeTOMLFile` — deep-merge with `.bak` |
| `packages/cli/internal/globalpaths/` | Home-dir roots for all tools; `ResolveCodexSkillsGlobalDir` |
| `packages/cli/internal/scaffold/root.go` | Single emitter for memory docs (`memoryDocDestPath`) |
| `library/commands/*.toml` | Gemini custom slash command templates |
| `library/chatmodes/*.chatmode.md` | Copilot chat mode templates |
| `packages/cli/internal/scaffold/root.go#fillClaudeMdPlaceholders` | Hybrid template-placeholder substitution (mechanical auto-infer + subjective fill-in markers) |
| `packages/cli/internal/adapter/opencode_frontmatter.go` | `BuildOpenCodeAgentFrontmatter` — emits opencode-schema YAML frontmatter, drops incompatible source fields |
| `packages/cli/internal/adapter/opencode_validate.go` | `ValidateOpenCodeInstall` — post-install opencode debug checks via injectable `CmdRunner` |
| `library/opencode/commands/` | OpenCode slash command templates (review, test, commit) |
| `library/opencode/modes/` | OpenCode chat mode templates (plan, audit) |
| `library/opencode/plugins.json` | Curated list of installable plugin module names |
| `packages/cli/internal/adapter/claude_cli.go` | `ClaudeCLIRunner` interface, `LookupClaudeBinary()` — testable substrate for `claude` CLI invocations (spec 012) |
| `packages/cli/internal/adapter/copilot_cli.go` | `CopilotCLIRunner` interface, `LookupCopilotBinary()`, `CopilotHomePresent()` — Copilot probe helpers (spec 013) |
| `packages/cli/internal/adapter/mcp_compiler.go#toCopilotServerEntries` | Shared per-server translation for Copilot; callers `toCopilotVSCodeMcp` (uses `servers`) and `toCopilotCLIMcp` (uses `mcpServers`) split the two schema surfaces (spec 014) |
| `packages/cli/internal/adapter/mcp_compiler.go#compileCopilotCLIMcp` | Deep-merge emitter for `~/.copilot/mcp-config.json`; runs at every scope when probe passes (spec 014) |
| `packages/cli/internal/adapter/mcp_compiler.go#writeClaudeSettingsLocal` | Deep-merge emitter for `.claude/settings.local.json` when `--local-secrets` flag is set (spec 015) |
| `packages/cli/internal/scaffold/gitignore.go#CheckGitignoreGuidance` | Appends `.claude/settings.local.json` to existing `.gitignore` when `--local-secrets` is set; idempotent (spec 015) |
| `packages/cli/internal/plugin/plugin.go#Build` | Generates a Claude Code plugin directory from the library FS: manifest, agents (forbidden-field stripping), skills (flat → `<name>/SKILL.md`), commands, output styles (spec 016) |
| `packages/cli/cmd/build_plugin.go` | `lazyai-cli build-plugin --out <path> [--force]` cobra subcommand (spec 016) |
| `packages/cli/cmd/build_helpers.go#preflightOutDir` | Shared out-dir preflight logic reused by `build-plugin` and `build-gemini-extension` (spec 017) |
| `packages/cli/internal/library/embed.go#ResolveGeminiCommandsSubdir` | Resolves preferred `gemini/commands` with fallback to legacy top-level `commands/` for one release (spec 017) |
| `packages/cli/internal/geminiext/geminiext.go#Build` | Generates a Gemini CLI extension directory: `gemini-extension.json`, raw `GEMINI.md`, commands (with namespacing), static-only `mcpServers` (spec 017) |
| `packages/cli/cmd/build_gemini_extension.go` | `lazyai-cli build-gemini-extension --out <path> [--force]` cobra subcommand (spec 017) |
| `packages/cli/internal/library/embed.go#CodexAssetsDir` | Per-tool dir helper for `library/codex/`; `CodexAgentsOverrideTemplate` constant points to the starter template (spec 018) |
| `packages/cli/internal/adapter/codex.go#writeCodexAgentsOverride` | Copies `library/codex/AGENTS.override.template.md` into the config root on first install; never overwrites user-authored content (spec 018) |
| `packages/cli/internal/adapter/codex.go#displayCodexInstallSummary` | Post-install summary via `codex mcp list --json` with plaintext fallback; matches the Claude Code summary pattern from spec 012 (spec 018) |
| `packages/cli/internal/adapter/codex.go#codexExecValidationArgs` | Argv builder for `RunHeadlessValidation`; includes `--skip-git-repo-check` so the probe succeeds against non-repo workspaces (spec 018 fix) |
| `library/claudecode/commands/` | Claude Code slash command templates (review, test, commit) |
| `library/claudecode/output-styles/` | Claude Code output style templates (terse, explanatory) |

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
- [x] ~~OpenCode structural conformance (config, frontmatter, MCP, commands, modes)~~ — spec 011
- [x] ~~Post-install opencode debug validation~~ — spec 011
- [x] ~~OpenCode plugin install flow~~ — spec 011
- [ ] Snapshot tests for library assets + compiled output (deferred in spec 009)
- [ ] `--drive-cli` for OpenCode (interactive-only upstream — permanently deferred)
- [x] ~~`--drive-cli` for Copilot (renamed to global MCP compile; no scriptable `copilot mcp add` upstream)~~ — spec 014
- [ ] CI-side validation with opencode binary (deferred in spec 011)
- [x] ~~`claude mcp add-json` CLI-driven registration (deferred from spec 012; needs scope → flag mapping + fallback)~~ — spec 012 task 010
- [x] ~~Post-install verification summary via `claude mcp list` + `claude agents` (deferred from spec 012)~~ — spec 012 task 014
- [x] ~~`settings.local.json` coverage for Claude Code (deferred from spec 012; user secrets, local-only config)~~ — spec 015 (`--local-secrets` flag)
- [x] ~~Ship ai-setup as a Claude plugin manifest (deferred from spec 012; plugin schema version + capabilities)~~ — spec 016 (`ai-setup build-plugin` generator)

## Feature 001 W1.A — Completed Scope

| Item | Status | Notes |
|------|--------|-------|
| N8 Constitution Population | ✅ Done | |
| N4 Coverage Thresholds | ✅ Done | |
| N11 Standards-as-Code | ✅ Done | |
| Go targeted W1.A packages | ✅ Green | |
| TS typecheck (W1.A) | ✅ Green | |
| TS targeted W1.A tests | ✅ Green | 5 files / 76 tests |
| `git diff --check` | ✅ Clean | |
| S1 security fix (red-team re-review) | ✅ Approved | |
