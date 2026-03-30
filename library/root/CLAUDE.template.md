# CLAUDE.md

**Project:** [YOUR_PROJECT_NAME]
**Organization:** [YOUR_ORG]
**Team:** [YOUR_TEAM]

---

## Project Overview

[YOUR_PROJECT_DESCRIPTION]

## Tech Stack

- [YOUR_TECH_STACK]

## Codebase Map

| Component | Responsibility | Path |
|-----------|---------------|------|
| [YOUR_COMPONENT_1] | [YOUR_RESPONSIBILITY_1] | [YOUR_PATH_1] |
| [YOUR_COMPONENT_2] | [YOUR_RESPONSIBILITY_2] | [YOUR_PATH_2] |

## Architecture & Patterns

[YOUR_ARCHITECTURE_NOTES]

## Conventions

- **Code Style:** [YOUR_CODE_STYLE]
- **Naming:** [YOUR_NAMING_CONVENTIONS]
- **Testing:** [YOUR_TESTING_STRATEGY]
- **Git:** [YOUR_GIT_WORKFLOW]

## Decision Tree

Before starting work, identify the task type and follow the appropriate guide:

| Task Type | Guide | Key Process |
|-----------|-------|-------------|
| Feature (new) | `@docs/features/AGENTS.md` | Research → PRD → TechSpec → Implement → Verify |
| Bugfix | `@docs/bugfixes/AGENTS.md` | Reproduce → Root-cause → Fix → Regression test |
| Refactor | `@docs/refactors/AGENTS.md` | ADR → Plan → Phased implementation |
| Tech Debt | `@docs/tech-debt/AGENTS.md` | Risk assessment → Prioritize → Incremental fix |
| Architecture Decision | `@docs/adrs/AGENTS.md` | Context → Options → Decision → Record |
| Standards/Rules | `@docs/standards/AGENTS.md` | Review existing → Propose → Document |
| Documentation | `@docs/AGENTS.md` | Structure → Write → Cross-reference |

> **Don't know where to start?** Read `@docs/AGENTS.md` first for the full documentation map.

## Rules

<!-- Path-specific rules live in .claude/rules/ -->
<!-- Each rule file uses YAML frontmatter with paths: glob patterns -->
<!-- Example: .claude/rules/typescript.md with paths: ["src/**/*.ts"] -->

- [YOUR_RULE_1]
- [YOUR_RULE_2]

## Do NOT

- Do not push directly to main — always use branches and PRs
- Do not modify generated files without updating the source template
- [YOUR_DO_NOT_1]
- [YOUR_DO_NOT_2]

## Workflow

1. **Branch:** Create a feature branch from main
2. **Research:** Explore the codebase and understand existing patterns
3. **Plan:** Create a task list with dependencies
4. **Implement:** Write tests first, then implementation
5. **Verify:** Run all quality checks before committing
6. **Review:** Open a PR for human review and merge

### Task Sizing

- **Small** (<20 lines changed): Direct implementation
- **Medium** (20-100 lines): Brief plan → implement → test
- **Large** (100+ lines): Research → plan → staged implementation with checkpoints

## Testing

- **Unit Tests:** [YOUR_UNIT_TESTING_STRATEGY]
- **Integration Tests:** [YOUR_INTEGRATION_TESTING_STRATEGY]
- **E2E Tests:** [YOUR_E2E_TESTING_STRATEGY]

## Key Commands

| Command | Purpose |
|---------|---------|
| [YOUR_DEV_COMMAND] | [YOUR_DEV_DESCRIPTION] |
| [YOUR_TEST_COMMAND] | [YOUR_TEST_DESCRIPTION] |
| [YOUR_BUILD_COMMAND] | [YOUR_BUILD_DESCRIPTION] |

## Session Start Checks

1. Read this file completely
2. Review recent git log for context
3. Check `docs/` for project documentation and standards
4. Verify you are on the correct branch
5. [YOUR_SESSION_CHECK]

## Recovery Procedures

- If tests fail: [YOUR_RECOVERY_PROCEDURE]
- If build breaks: [YOUR_RECOVERY_PROCEDURE]

## Memory & Context

<!-- Claude Code auto-manages memory in ~/.claude/projects/ -->
<!-- Reference project docs for persistent context -->

@docs/rules/
@docs/standards/
@docs/AGENTS.md
@docs/features/AGENTS.md
@docs/bugfixes/AGENTS.md
<!-- For other task types, reference the corresponding docs/*/AGENTS.md guide -->

## Self-Improvement Protocol

After completing a task:
1. Update documentation if any interfaces or behaviors changed
2. Add lessons learned to `docs/memory/`
3. [YOUR_SELF_IMPROVEMENT_STEP]

### Impact Assessment

Before proposing process changes, assess severity:

| Severity | Examples | Action |
|----------|----------|--------|
| Critical | Wrong architecture, security flaw | Stop and escalate immediately |
| High | Missing tests, broken conventions | Fix before continuing |
| Medium | Suboptimal pattern, minor inconsistency | Note and fix if time permits |
| Low | Style preference, minor naming | Record for future reference |
