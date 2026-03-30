# GitHub Copilot Instructions

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

## Rules

<!-- GitHub Copilot loads .github/copilot-instructions.md for repository-wide instructions -->
<!-- Use .github/instructions/*.instructions.md for path-specific rules with YAML frontmatter -->
<!-- Example: .github/instructions/typescript.instructions.md with applyTo: "**/*.ts" -->

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

<!-- GitHub Copilot supports .github/prompts/*.prompt.md for reusable prompt files -->
<!-- Use .github/instructions/ for path-specific instructions with YAML applyTo frontmatter -->
<!-- AGENTS.md files in directories provide agent-specific instructions -->

## Self-Improvement Protocol

After completing a task:
1. Update documentation if any interfaces or behaviors changed
2. Add lessons learned to `docs/memory/`
3. [YOUR_SELF_IMPROVEMENT_STEP]
