<rule>
  <scope>auto</scope>
  <globs>**</globs>
  <description>Main context entry point for the tool directory and its role-based subdirectories</description>
</rule>

# Tool Directory Root

This file is the primary entry point for directory-level guidance. Use it to choose the minimal context required for the task at hand, then load deeper guidance only where needed.

## Directory Map

| Directory | Contains | Context File |
|-----------|----------|--------------|
| agents/ | Role-specialized agent definitions and handoff protocols | agents/AGENTS.md (adapter-installed from `agents-dir.md`) |
| skills/ \\| commands/ \\| prompts/ | Workflow skills and execution recipes (naming varies by adapter) | skills/CLAUDE.md or prompts/AGENTS.md (adapter-installed from `skills-dir.md`) |
| templates/ | Prompt and task templates for recurring work types | templates/CLAUDE.md or templates/AGENTS.md (adapter-installed from `templates-dir.md`) |

## Agent Pipeline Overview

Use this tool-agnostic pipeline as the default flow:

1. **Scout** researches the problem space and constraints.
2. **Planner** creates a sequenced execution plan with verification steps.
3. **Builder** implements approved changes in minimal, testable increments.
4. **Reviewer** validates correctness, quality, and requirement coverage.
5. **Documenter** captures outcomes, rationale, and reusable learnings.

**Red-Team** is available on demand for security, abuse-case, or adversarial reviews.

## Progressive Loading

Load ONLY the directory context required for your current phase.

- Planning task: load root + agents context; skip implementation-heavy files.
- Execution task: load root + skills context + relevant template.
- Documentation task: load root + templates context + documenter guidance.

Do not load every agent or workflow file by default. Overloading context increases ambiguity and slows decision quality.

## Rules Reference

Always-active constraints live in `docs/rules/`.

Core categories to consult:
- workflow
- testing
- security
- access
- review
- cost
- code-style

When in doubt, prefer explicit rule guidance over inferred behavior.

## Standards Reference

Project-specific implementation patterns live in `docs/standards/`.

Use standards to align with established conventions for architecture, code structure, testing style, and quality expectations.

## Memory Reference

Historical suggestions and contextual notes live in `docs/memory/`.

Consult memory to avoid repeating past mistakes, recover rationale, and accelerate similar tasks.

## Operating Principles

- Keep scope explicit before execution.
- Prefer small, reversible steps with clear verification.
- Escalate unknowns early instead of making hidden assumptions.
- End each task with captured learnings and updated references.

## Self-Improvement

When structure or process changes:
- Update the directory map table to reflect added, renamed, or removed paths.
- Refresh pipeline guidance if agent responsibilities evolve.
- Update references when new rule or standards categories are introduced.
- Ensure progressive loading examples match real team workflows.
