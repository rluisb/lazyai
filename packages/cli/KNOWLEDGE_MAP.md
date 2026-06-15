# Project Knowledge Map

> Navigable index of all project documentation.
> The AI reads this for orientation. Update when creating new work items or ADRs.

---

## Architecture Decisions

| ADR | Decision | Feature | Status |
|-----|----------|---------|--------|
| [specs/adrs/001-title.md] | [what was decided] | [NNN-feature — or "Cross-cutting"] | Accepted |

## Active Features

| ID | Name | Status | ADRs |
|----|------|--------|------|
| [specs/features/001-name/] | [description] | [Research/Plan/Implement/Done] | [NNN, NNN — or "None"] |

## Active Bugfixes

| ID | Name | Status |
|----|------|--------|
| [specs/bugfixes/001-name/] | [description] | [Research/Fix/Done] |

## Active Refactors

| ID | Name | Status | ADRs |
|----|------|--------|------|
| [specs/refactors/001-name/] | [description] | [Research/Plan/Implement/Done] | [NNN] |

## Tech Debt

| ID | Name | Priority | Status |
|----|------|----------|--------|
| [specs/tech-debt/001-name/] | [description] | [Low/Med/High/Critical] | [Identified/Planned/In Progress/Resolved] |

## Rules & Standards

| Type | Files | Purpose |
|------|-------|---------|
| Rules | specs/rules/*.md | Prescriptive — WHAT to do |
| Standards | specs/standards/*.md | Descriptive — HOW we do it |

## Key Modules

<!-- Map your codebase's main modules for quick AI orientation. -->

| Path | Responsibility | Owner |
|------|---------------|-------|
| [src/module-a/] | [what it does] | [team/person] |
| [src/module-b/] | [what it does] | [team/person] |
| [src/shared/] | Shared utilities (read-only for agents) | [team] |


## Terminology

### Accepted domain terms

| Term | Meaning | Source of truth |
|------|---------|-----------------|
| setup-core | The default lazyai-cli command set (init, compile, update, doctor, add, build-plugin, etc.) | `specs/adrs/005-core-vs-optional-modules.md` |
| runtime-adjacent module | Optional command families (session, message, ledger, memory, auth, cost, metrics, notify, secret, backup, restore-runtime-db, git) | `specs/adrs/005-core-vs-optional-modules.md` |
| artifact type | A category of generated asset: agent, skill, command, prompt, template, rule, infra, specs-dir | `packages/cli/internal/validation/validation.go` |
| library | Embedded reference content used by lazyai-cli to populate target repos | `packages/cli/library/` |
| curation manifest | YAML manifest of every embedded library asset with provenance metadata | `packages/cli/library/manifests/curation.yaml` |

### Vocabulary source of truth

Runtime must not introduce a dedicated terminology lookup subsystem: the source of truth stays in this Markdown section, with derived enums (e.g. `packages/cli/internal/types/types.go`) hand-mapped from the table above. If a lookup feels necessary, extend the table instead.
