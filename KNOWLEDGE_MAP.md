# Project Knowledge Map

> Navigable index of all project documentation.  
> The AI reads this for orientation. Update when creating new work items or ADRs.  
> **Speckit layout:** specs go in `.specify/`; legacy `specs/` archived.

---

## Terminology

Accepted domain terms live here as lightweight Vocabulary source of truth. New or ambiguous terms should be proposed through clarify/HITL before agents adopt them.

## Architecture Decisions

| ADR | Decision | Status |
|-----|----------|--------|
| [022] Speckit workflow alignment | Option B: Adapt & Enhance — absorbed speckit into ai-setup library, not a runtime dependency | Accepted |
| [017] Gemini command layout | Gemini TOML commands in `library/gemini/commands/` with legacy fallback | Accepted |

## Active Features

| Spec | Name | Status |
|------|------|--------|
| [022](.specify/features/022-speckit-workflow-alignment/) | Speckit Workflow Alignment | Done |
| workspace-root | Workspace Root Wizard | Done |

## Rules & Standards

| Type | Files | Purpose |
|------|-------|---------|
| Rules | `.claude/rules/*.md`, `.opencode/rules/*.md` | Prescriptive — WHAT to do |
| Standards | `specs/standards/*.md` (legacy), `.specify/memory/` (new) | Descriptive — HOW we do it |
| Constitution | `.specify/memory/constitution.md` | Governing contract for all workflows |

## Key Modules

| Path | Responsibility |
|------|---------------|
| `packages/cli/` | Go CLI — canonical implementation (source of truth) |
| `packages/cli/library/` | Library: agents, skills, templates, commands, MCP catalog |
| `packages/cli/internal/adapter/` | Per-tool adapters (Claude, OpenCode, Gemini, Copilot, Codex) |
| `packages/cli/internal/scaffold/` | Scaffold pipeline (specs, constitution, MCP, templates, infra) |
| `packages/orchestrator/` | Go MCP server for multi-agent orchestration (optional, user opt-in) |
| `packages/diffviewer/` | Go diff review utility shipped as `lazyai-diffviewer` |
