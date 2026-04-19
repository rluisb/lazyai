# Requirements Checklist — 012: Claude Code Deep Setup

Verification criteria the implementation must satisfy before the spec can be marked complete. Each item should be testable (unit, integration, or manual smoke).

## Layout conformance — project scope
- [ ] `<project>/.mcp.json` written via `claude mcp add(-json) --scope project` (not raw JSON write)
- [ ] `<project>/.claude/settings.json` exists and merges (not overwrites) user keys
- [ ] `<project>/.claude/agents/<name>.md` for each selected agent
- [ ] `<project>/.claude/skills/<name>/SKILL.md` for each selected skill
- [ ] `<project>/.claude/commands/` directory present (empty if no assets — directory itself created)
- [ ] `<project>/.claude/output-styles/` directory present
- [ ] `<project>/.claude/rules/<rule>.md` sourced from `library/rules/`, not hardcoded
- [ ] `<project>/.claude/CLAUDE.md` (project conventions) present
- [ ] `<project>/.claude/agents/CLAUDE.md` context file present and references the correct dir
- [ ] `<project>/.claude/skills/CLAUDE.md` context file present
- [ ] Root `<project>/CLAUDE.md` present with placeholders filled

## Layout conformance — workspace scope
- [ ] All project-scope items above, rooted at `<workspace>/`
- [ ] No leakage into the user's home dir during install

## Layout conformance — global / user scope
- [ ] `~/.claude/agents/<name>.md` (NOT flat at `~/.claude/<name>.md`) — primary defect from research
- [ ] `~/.claude/skills/<name>/SKILL.md`
- [ ] `~/.claude/commands/` directory present
- [ ] `~/.claude/output-styles/` directory present
- [ ] `~/.claude/rules/<rule>.md` sourced from `library/rules/`
- [ ] MCP servers added via `claude mcp add --scope user`, not by hand-merging `settings.json`
- [ ] No `~/.mcp.json` written (it's not a recognized location)
- [ ] `~/.claude/settings.json` deep-merged, no clobber of pre-existing keys (ConfigChange hooks etc.)
- [ ] Personal-conventions `~/.claude/CLAUDE.md` not stomped by tool-context content
- [ ] Orchestrator behavior at global scope is intentional and documented (either present or explicitly excluded with rationale)

## Frontmatter contracts
- [ ] Agent frontmatter `tools` field uses **whitespace-separated** values (Claude's documented form)
- [ ] Agent frontmatter contains `name` and `description` at minimum
- [ ] Skill `SKILL.md` frontmatter contains `name` (or relies on dir name) and `description`
- [ ] No `mode:` or other OpenCode-only keys leak into Claude agents

## CLI orchestration
- [ ] `claude` CLI presence is probed before use
- [ ] When CLI is missing, install falls back to direct-write and emits a warning
- [ ] `--drive-cli=false` (or env equivalent) forces direct-write path for testability
- [ ] `claude mcp list` post-check runs after install and is included in the install summary
- [ ] Failures from `claude mcp add` surface as errors, not silent fallbacks

## Tests
- [ ] Unit tests for frontmatter schema on every emitted agent/skill
- [ ] Unit tests for MCP server transformation (canonical → CLI args / JSON payload)
- [ ] Scope-parity test: project, workspace, global all install successfully and produce expected file tree
- [ ] Bug regression test: global `agents/` subdir is created and populated
- [ ] CLI integration test (with injectable `CmdRunner`) for `claude mcp add` invocation shape
- [ ] Round-trip test: write MCP via CLI, then `claude mcp list` reports the same servers

## Documentation
- [ ] `specs/KNOWLEDGE_MAP.md` updated with spec 012 status
- [ ] Any new package paths added to root `CLAUDE.md` codebase map
- [ ] ADR recorded if architectural decision is non-obvious (e.g. CLI-first MCP)

## Backwards compatibility
- [ ] Re-running `ai-setup init` over an existing install does not lose user-authored content (settings.json keys, MCP servers added by hand, custom agents)
- [ ] Existing global-scope installs with flat agents (the bug) are migrated or at least not broken silently
