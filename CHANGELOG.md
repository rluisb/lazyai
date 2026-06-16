# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Changed
- **Exact vibe-lab baseline parity for default agent/tool surfaces.** LazyAI now
  emits the same seven baseline agents as `/Users/ricardo/code/vibe-lab` across
  Claude Code (`.claude/agents`), OpenCode (`.opencode/agents`), GitHub Copilot
  (`.github/agents`), and compatible `bin/` maintainer commands. The Go adapter
  layer and `bin/inject` now emit the same managed-marker contract and frontmatter
  shape as the baseline.

### Added
- `bin/inject.original` baseline artifact, copied as non-executable from vibe-lab.
- Regression tests for baseline-style agents without LazyAI `tier` metadata and
  for a default Copilot install producing exactly seven `.agent.md` files with no
  `.agent.yaml` leakage.

### Fixed
- OpenCode default config is now root `opencode.json` copied from the vibe-lab
  baseline; `.opencode/opencode.jsonc` was removed. LazyAI-only MCP/runtime extras
  remain isolated in `.opencode/lazyai.mcp.jsonc`.
- OpenCode hook plugin surface restored to baseline name
  `.opencode/plugins/vibe-lab-hooks.js` / `VibeLabHooks`.
- `ValidateAgentResolutions` now tolerates the missing `tier` field for canonical
  baseline agents while still reporting malformed frontmatter.
- Removed a stray `XXXX CONFIG_PATH` debug print from OpenCode adapter output.
### Added (Plan C parity follow-up)
- `packages/cli/library/skills/evidence-verifier.md` added as a setup-library skill; the canonical `evidence-verifier` agent was already emitted by adapters and the duplicate name is avoided by keeping only the canonical agent.
- `packages/cli/library/skills/issue-triage.md` and `packages/cli/library/skills/task-to-issues.md` are now curated and tracked in `curation.yaml` as adapter-support skills.
- `HookGenerator` already registered in `packages/cli/internal/generator/registry.go` for `lazyai-cli create hook <name>`; covered by `TestHookGeneratorRegisteredInRegistry`.

## [1.1.3] - 2026-05-13

### Fixed
- `lazyai-cli --version` now resolves Go module build metadata when installed via `go install ...@vX.Y.Z`, while preserving CI release ldflags precedence and local `0.0.0-dev` behavior.

## [1.1.2] - 2026-05-12

### Added
- **Orchestrator Dashboard**: Embedded dashboard MVP with global event streams, run details, budget cards, and Catppuccin theming.
- **Design System**: Comprehensive refactor of the CLI to use the new Catppuccin-based design system for logs, errors, and interactive forms.
- **Automated Initialization**: Simplified install wizard and automated AI tool initialization after scaffolding.
- **Doctor Diagnostics**: Added a diagnostic to detect and help remove stale `ai-setup-orchestrator` MCP entries.
- **New Skills**: Adopted `diagnose` and `improve-codebase-architecture` skills from mattpocock/skills.
- **Auth Command**: Added `auth list` command.
- **Atlassian MCP**: Switched Atlassian scaffold to remote authv2 endpoint.
- **Open Source Readiness**: Auto-generate `llm.txt` from curated docs on every MkDocs build and added community infrastructure (DCO, issue templates).

### Changed
- **Copilot**: Bumped `claude-sonnet-4.5` to `4.6` across agent sources.
- **Security**: Hardened RPI human gates to prevent auto-mode bypasses.
- Renamed `--non-interactive` flag to `--no-interactive`.
- Added structured logging across Go packages.
- Migrated orchestrator MCP tool schemas to typed schemas for reliability.

### Fixed
- Fixed OpenCode frontmatter + Copilot skill tier resolver.
- Fixed workspace artifact routing to correct roots.
- Fixed Claude Code agent description frontmatter emission on install.
- Fixed contract validator false positives.

## [1.0.0] - 2026-05-04

### Changed
- **Breaking:** Project renamed from `ai-setup` to `LazyAI`. Binary names, Go module paths, and GitHub org all updated.
- Binary `ai-setup` → `lazyai-cli`, `ai-setup-orchestrator` → `lazyai-orchestrator`, `diffviewer` → `lazyai-diffviewer`.
- Go module path changed from `github.com/ricardoborges-teachable/ai-setup` to `github.com/rluisb/lazyai`.
- Installation is Go-only via `go install`. npm/npx distribution has been removed.
- All Teachable/org references removed from source, specs, and documentation.

### Migration
- See [Migration from ai-setup to LazyAI](docs/migration/ai-setup-to-lazyai.md) for the full checklist.
- Local state files (`.ai-setup.json`, `.ai-setup.db`, `.ai-setup.toml`, `.ai-setup-backup/`) are **not** automatically renamed. They continue to work as-is.

## [0.2.0] - 2026-04-01

### Added
- Introduced the new Migration Engine for importing existing AI assistant setups into `ai-setup` without starting from scratch.
- Added `ai-setup import` / `ai-setup migrate` flows to detect and migrate existing configurations from OpenCode, Claude Code, Pi, Gemini CLI, and GitHub Copilot.
- Added preview support so migrations can be reviewed before any files are written.
- Added merge strategies for different migration styles: `smart`, `preserve`, `replace`, and `append`.
- Added backup-aware migration execution so existing files can be preserved before replacement.
- Added drift checking support for migration-managed setups via `ai-setup doctor --migration-check`.
- Added extensible parser discovery to support built-in, local, global, and npm-based migration parsers.

### Notes
- This release focuses on helping teams adopt `ai-setup` incrementally by importing their current AI tooling conventions and customizations.
- GitHub release creation should remain draft-only for this version unless the npm publish workflow is intentionally used.

## [0.1.0] - 2026-03-01

### Added
- Initial release with `ai-setup init` command and interactive 8-phase wizard for tool-agnostic `.ai/` setup.
- Added canonical source → compile model: `ai-setup compile` re-generates tool-native directories from `.ai/`.
- Added support for `project`, `workspace`, and `global` setup scopes.
- Added adapters for **OpenCode**, **Claude Code**, **Codex (OpenAI)**, **GitHub Copilot**, **Gemini CLI**, and **Pi**.
- Added `ai-setup doctor` for drift detection against manifest hashes.
- Added `ai-setup status`, `ai-setup add <tool>`, `ai-setup update`, `ai-setup eject`, and `ai-setup create` commands.
- Added lowdb-backed `.ai-setup.json` manifest with Zod schemas, structured error types, and migration support.
- Added workspace scope with planning-repo-only setup and referenced-repo scanning for type detection (Rails, Next.js, Go, etc.).
- Added global scope compiling to `~/.config/opencode/` and `~/.claude/` for OpenCode and Claude Code.
- Added reusable agent guidance, skills scaffolding, and constitution documents for supported tools.
