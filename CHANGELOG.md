# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0] - 2026-05-12

### Added
- **Orchestrator Dashboard**: Embedded dashboard MVP with global event streams, run details, budget cards, and Catppuccin theming.
- **Design System**: Comprehensive refactor of the CLI to use the new Catppuccin-based design system for logs, errors, and interactive forms.
- **Automated Initialization**: Simplified install wizard and automated AI tool initialization after scaffolding.
- **New Skills**: Adopted `diagnose` and `improve-codebase-architecture` skills.
- **Open Source Readiness**: Auto-generate `llm.txt` from curated docs on every MkDocs build and added community infrastructure.

### Changed
- **Breaking:** Project renamed from `ai-setup` to `LazyAI`. Binary names, Go module paths, and GitHub org all updated.
- Binary `ai-setup` → `lazyai-cli`, `ai-setup-orchestrator` → `lazyai-orchestrator`, `diffviewer` → `lazyai-diffviewer`.
- Go module path changed from `github.com/ricardoborges-teachable/ai-setup` to `github.com/rluisb/lazyai`.
- Installation is Go-only via `go install`. npm/npx distribution has been removed.
- All Teachable/org references removed from source, specs, and documentation.
- **Copilot**: Bumped `claude-sonnet-4.5` to `4.6` across agent sources.
- **Security**: Hardened RPI human gates to prevent auto-mode bypasses.

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