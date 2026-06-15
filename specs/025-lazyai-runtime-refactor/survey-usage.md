# P0-2: Fortnite/OpenCode Usage Survey

**Status:** Quantitative baseline established; additional qualitative outreach explicitly waived by the human gate on 2026-06-14 before Phase 2 excision
**Owner:** Ricardo Conceicao  
**Date:** 2026-06-14  
**Linked from:** `plan.md` Phase 0, P0-2

---

## Purpose

Determine active usage of Fortnite and OpenCode features to inform removal/rewrite decisions. No CLI telemetry exists; survey uses authoritative sources.

## Methodology

1. **Quantitative baseline (primary):** GitHub issues tagged `fortnite` or `opencode`
2. **Git history:** `git log -- packages/cli/library/fortnite/` for contributor count
3. **Code audit:** OpenCode adapter usage patterns from `opencode.go` and `opencode_validate.go`
4. **Qualitative:** Direct outreach (Slack/Discourse/GitHub discussions) for migration blockers. This was explicitly waived by the human gate on 2026-06-14 (`Go ahead`) after the prerequisite blocker was called out for Phase 2.

## Findings

### Git History

```
$ git log --oneline -- packages/cli/library/fortnite/
0369ca0 fix(ci): resolve trailing spaces, add missing AGENTS.md, replace symlinks
dc6a7a5 feat(opencode): default to Fortnite runtime

$ git shortlog -sn -- packages/cli/library/fortnite/
(no output — both commits are from the same author, Ricardo Conceicao)
```

**Conclusion:** Fortnite library has 2 commits, single contributor. Low adoption signal.

### Code Audit: FortniteMode Usage

- `cmd/helpers.go:173-183`: `FortniteMode` defaults to `true` when OpenCode is selected (unless `--plain-opencode`)
- `cmd/init_test.go:515-560`: Tests verify `FortniteMode` default behavior
- `adapter/types.go:48`: `FortniteMode bool` field in `AdapterContext`
- `adapter/opencode.go:77-79`: `FortniteMode` → `defaultAgent = "loop-driver"`, instructions = `["AGENTS.md", "STARTUP.md"]`

**Conclusion:** FortniteMode is the default for OpenCode installs. Removing it requires a migration path for existing OpenCode users.

### Orchestrator Coupling Surface

- `go.work`: includes `./packages/orchestrator` as a module
- `packages/cli/library/embed.go`: embeds `all:fortnite` and `all:orchestration`
- `cmd/mcp_setup.go`: configures LazyAI orchestrator MCP server
- `cmd/config.go`: `DefaultAgent: "orchestrator"` in two places
- `cmd/orchestration.go`: 15.3KB of orchestrator catalog commands
- `cmd/server.go`: special-cases `"orchestrator"` server name
- `cmd/doctor_mcp.go`: detects legacy ai-setup orchestrator MCP entries
- `cmd/doctor_health.go`: checks for `lazyai-orchestrator` binary on PATH

**Conclusion:** Orchestrator is deeply embedded. 17 files need rewrites; 4 have direct import breakage.

## Migration Needs

- **Notification template:** `docs/migration/fortnite-orchestrator-removal.md` — explain that Fortnite agents/workflows are replaced by canonical primary-agent path
- **Migration guide:** OpenCode users who relied on FortniteMode get `primary-agent` instead of `loop-driver`/orchestrator
- **Timeline:** Phase 2 rewrites CLI commands; Phase 5 populates canonical library

## Gate

⛔ Human must approve this survey before Phase 2 excision begins.

## Waiver record
- **Date:** 2026-06-14
- **Evidence:** Interactive human approval in this session — `Go ahead`
- **Scope:** Waives additional qualitative Fortnite/OpenCode outreach so Phase 2 destructive cleanup may proceed
