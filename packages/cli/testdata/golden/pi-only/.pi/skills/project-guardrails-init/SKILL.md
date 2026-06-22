---
name: project-guardrails-init
description: Use when onboarding into a project to discover existing stack, architecture, commands, and conventions before proposing rules or memory.
---

# Project Guardrails Init

## When to Use

Use this skill when pointing an agent at a new repository, a new subsystem, or a project that lacks clear agent-facing rules.

Use it before adding project rules, coding standards, hooks, or canonical docs.

## Purpose

Discover the project's existing engineering shape and propose guardrails that preserve it.

This skill produces a proposal. It does not write memory, canonical docs, hooks, or context files without explicit approval.

## Inspect

Read only what is needed to answer the current onboarding question:

1. Project overview: `README`, docs, existing agent/context files.
2. Stack and framework: package manifests, lockfiles, framework config, runtime config.
3. Validation: test, lint, typecheck, build, and CI commands.
4. Architecture: entry points, module boundaries, data flow, storage, and external integrations.
5. Conventions: naming, file organization, error handling, testing style, dependency style.
6. Generated or sensitive areas: generated code, migrations, secrets, vendored files, local-only config.

## Output

Return a compact project working agreement:

```md
## Project Working Agreement

### Stack and Framework
- <observed fact>

### Architecture
- <observed boundary or data-flow fact>

### Validation Commands
- <command and when to use it>

### Existing Patterns
- <pattern to preserve>

### Coding Standards
- <project-specific convention>

### Do Not
- <generated file, unsafe command, or rejected pattern>

### Candidate Memory
- <observed fact that should be remembered after approval>

### Candidate Canonical Rules
- <stable rule that may deserve canonical promotion after approval>

### Open Questions
- <unknown that affects correctness>
```

## Promotion Flow

Use this path for discovered engineering knowledge:

```text
observed pattern -> candidate memory -> confirmed memory -> canonical rule
```

- **Candidate memory:** useful project-local fact, gotcha, command, or convention.
- **Confirmed memory:** fact reused or explicitly approved after review.
- **Canonical rule:** stable, project-wide rule that should be visible to every agent.

Ask before each promotion step.

## Hooks

Do not create hooks during discovery.

Recommend hooks only for narrow mechanical checks after the project commands and standards are confirmed, such as generated context freshness or known validation command availability.

## Constraints

- Do not overwrite existing `AGENTS.md`, `CLAUDE.md`, rules, or docs silently.
- Do not impose generic SOLID, DDD, OOP, FP, or design-pattern doctrine over local patterns.
- Do not invent validation commands when manifests or docs disagree.
- Do not promote secrets, credentials, local machine paths, or transient branch status.
- Prefer a markdown proposal before any write.
