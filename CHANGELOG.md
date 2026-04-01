# Changelog

All notable changes to this project will be documented in this file.

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
