# LazyAI + vibe-lab — Schemas and Examples

This document gives concrete examples for the complete product spec.

---

## 1. Starter `.ai/` tree

```text
.ai/
  lazyai.json
  lock.json
  mcp.json
  agents/
    guide.md
    planner.md
    researcher.md
    implementer.md
    reviewer.md
    deployer.md
    responder.md
    evidence-verifier.md
  skills/
    pr-review/SKILL.md
    rpi-bugfix/SKILL.md
    handoff/SKILL.md
    context-compaction/SKILL.md
  rules/
    anti-slop.md
    human-gates.md
    trace-evidence.md
    security.md
  fragments/
    rpi-workflow.md
    evidence-format.md
    handoff-format.md
  prompts/
    plan-from-ticket.md
    create-adr.md
    bugfix-rca.md
  commands/
    review-pr.md
    rpi-bugfix.md
  hooks/
    protected-paths.yaml
    evidence-after-tool.yaml
    no-secrets.yaml
  templates/
    adr.md
    bugfix-rca.md
    spec.md
    eval-case.yaml
  standards/
    starter/testing.md
    starter/security.md
    starter/errors.md
  evals/
    cases/
    holdouts/
    rubrics/
```

---

## 2. Manifest example

```json
{
  "$schema": "https://lazyai.dev/schemas/lazyai.schema.json",
  "version": "0.1",
  "profile": "team",
  "targets": ["opencode", "claude", "copilot", "pi", "omp", "antigravity", "kiro"],
  "library": {
    "packs": ["vibe-lab/starter", "vibe-lab/security"]
  },
  "adapters": {
    "opencode": {
      "enabled": true,
      "emit": ["agents", "skills", "commands", "hooks", "mcp"],
      "compatibility": {
        "emitAgentsSkills": true
      }
    },
    "claude": {
      "enabled": true,
      "emitClaudeMd": true,
      "emitPlugin": false
    },
    "copilot": {
      "enabled": true,
      "emitPlugin": false,
      "emitAgentsSkills": true
    },
    "pi": {
      "enabled": true,
      "emitPiSystem": false,
      "emitAgentsSkills": true
    },
    "omp": {
      "enabled": true,
      "emitPlugin": true
    },
    "antigravity": {
      "enabled": true,
      "emitGeminiMd": true,
      "writeGlobalSettings": false
    },
    "kiro": {
      "enabled": true,
      "emitSpecs": true,
      "emitSteering": true
    }
  },
  "safety": {
    "requireDiffBeforeWrite": true,
    "allowGlobalWrites": false,
    "denyInlineSecrets": true,
    "warnIfNoSandbox": true,
    "generatedFileMode": "managed-region"
  }
}
```

---

## 3. MCP catalog example

```json
{
  "$schema": "https://lazyai.dev/schemas/mcp-catalog.schema.json",
  "version": "0.1",
  "servers": {
    "ai-memory": {
      "type": "local",
      "command": "ai-memory",
      "args": ["server"],
      "env": {
        "AI_MEMORY_DB": "${AI_MEMORY_DB:-.ai/runtime/memory.db}"
      },
      "tools": [
        "memory_query",
        "memory_recent",
        "memory_handoff_accept",
        "memory_write_page",
        "memory_auto_improve",
        "memory_consolidate"
      ],
      "preferredUse": "hybrid",
      "description": "Long-term project memory and handoff annotations."
    },
    "filesystem": {
      "type": "local",
      "command": "filesystem-mcp",
      "args": ["--root", "${LAZYAI_WORKSPACE_ROOT}"],
      "tools": ["read_file", "write_file", "list_directory", "search_files"],
      "preferredUse": "cli-first",
      "safety": {
        "writeTools": ["write_file"],
        "defaultPermission": "ask"
      }
    },
    "codegraph": {
      "type": "local",
      "command": "codegraph",
      "args": ["mcp"],
      "tools": ["codegraph_context", "codegraph_search", "codegraph_files"],
      "preferredUse": "hybrid"
    },
    "obsidian": {
      "type": "local",
      "command": "ob",
      "args": ["mcp"],
      "env": {
        "OBSIDIAN_VAULT": "${OBSIDIAN_VAULT}"
      },
      "tools": ["read_note", "write_note", "search_notes", "list_notes"],
      "preferredUse": "hybrid"
    }
  }
}
```

---

## 4. Canonical reviewer agent

```markdown
---
name: reviewer
type: agent
description: Reviews code changes for correctness, safety, maintainability, and evidence quality.
mode: subagent
modelClass: medium-reasoning
tools:
  read: allow
  search: allow
  edit: deny
  bash: ask
permissions:
  destructive: deny
  network: ask
evidence:
  required: true
handoff:
  required: true
---

# Reviewer

## Role

You are a code reviewer. You review changes, diffs, pull requests, or proposed patches. You do not implement changes unless explicitly reassigned.

## When to use

Use this agent when the user asks to review, audit, check, inspect, validate, or assess a change.

## When not to use

Do not use this agent to implement a feature, write broad new code, or perform deploy actions.

## Process

1. Identify changed files or relevant files.
2. Understand intent and risk.
3. Check correctness, security, maintainability, and project conventions.
4. Verify test/lint/build evidence if available.
5. Separate blocking issues from suggestions.
6. State uncertainty explicitly.

## Evidence required

- Files inspected.
- Commands run or reason commands were not run.
- Risks found.
- Suggested fixes.

## Output format

```text
Summary:
Blocking issues:
Non-blocking suggestions:
Evidence:
Risks:
Next action:
```
```

---

## 5. Canonical skill example

```markdown
---
name: pr-review
description: Review a pull request, staged diff, or patch for correctness, risk, and verification evidence.
license: MIT
compatibility: lazyai,opencode,claude,copilot,pi,omp,antigravity
metadata:
  workflow: review
  owner: vibe-lab
---

# PR Review Skill

## What this skill does

This skill guides a structured review of a PR, diff, or staged change.

## Trigger examples

Use this skill when the user says:

- "review this PR"
- "check this diff"
- "audit staged changes"
- "look for risks before merge"

## Non-trigger examples

Do not use this skill when the user asks to:

- implement a new feature;
- brainstorm architecture;
- write documentation only;
- deploy a release.

## Procedure

1. Locate the diff or changed files.
2. Identify the intent of the change.
3. Inspect code paths affected by the change.
4. Check correctness, tests, security, maintainability, and style.
5. Run or request deterministic verification.
6. Produce blocking and non-blocking findings.

## Evidence required

- Files inspected.
- Commands run.
- Test/lint/build result or reason not run.
- Risk areas.

## Output format

```text
Review summary:
Blocking findings:
Non-blocking findings:
Verification evidence:
Residual risks:
```
```

---

## 6. Hook policy example

```yaml
---
name: protected-paths
type: hook
description: Require approval before editing protected files.
events:
  - before_tool
match:
  tools:
    - write
    - edit
    - apply_patch
paths:
  deny:
    - ".git/**"
    - ".env"
    - ".env.*"
    - "**/*secret*"
    - ".ai/lock.json"
action:
  type: require_approval
  message: "Protected path. Confirm before modifying."
severity: high
---
```

---

## 7. Generated OpenCode agent example

```markdown
---
description: Reviews code changes for correctness, safety, maintainability, and evidence quality.
mode: subagent
permission:
  read: allow
  grep: allow
  glob: allow
  list: allow
  edit: deny
  bash: ask
steps: 12
---

# Reviewer

You are a code reviewer...
```

---

## 8. Generated Claude skill example

```text
.claude/skills/pr-review/SKILL.md
```

```markdown
---
name: pr-review
description: Review a pull request, staged diff, or patch for correctness, risk, and verification evidence.
---

# PR Review Skill

...
```

---

## 9. Generated Copilot plugin example

```text
lazyai-vibelab-copilot-plugin/
  plugin.json
  agents/reviewer.agent.md
  skills/pr-review/SKILL.md
  hooks.json
  .mcp.json
```

`plugin.json`:

```json
{
  "name": "lazyai-vibelab",
  "version": "0.1.0",
  "description": "vibe-lab agents, skills, hooks, and MCP configuration compiled by LazyAI."
}
```

---

## 10. Generated Kiro steering example

```markdown
---
inclusion: always
---

# vibe-lab engineering rules

Use research → plan → implement → verify. Important transitions require human approval and evidence.
```

Path-scoped:

```markdown
---
inclusion: fileMatch
fileMatchPattern: "src/**/*.rs"
---

# Rust project rules

Run `cargo fmt`, `cargo clippy`, and relevant tests before marking work complete.
```

---

## 11. Eval case example

```yaml
id: skill-pr-review-trigger-001
title: PR review skill triggers on staged diff review request
tags:
  - skills
  - trigger-accuracy
  - review
input:
  user: "Review my staged changes before I commit."
expected:
  shouldUseSkill: pr-review
  shouldNotUseSkills:
    - deploy
    - rpi-bugfix
  evidenceRequired:
    - changed files inspected
    - verification command or reason not run
rubric:
  - name: correct_skill_trigger
    weight: 0.4
  - name: evidence_quality
    weight: 0.4
  - name: no_unrequested_edits
    weight: 0.2
holdout: false
```

---

## 12. Harness change report example

```markdown
# Harness change report

## Problem

Reviewer agent missed migration-file risks in 4 of 12 review traces.

## Evidence

- Trace IDs: local-2026-06-18-001, local-2026-06-19-003, local-2026-06-20-002, local-2026-06-20-009
- Failure type: `missing_context`, `review_quality`

## Proposed change

Update `pr-review` skill trigger/procedure to require checking migrations, schema changes, and data backfill risk when diff includes database files.

## Single change rule

Only the skill procedure changes. No agent prompt, hook, or MCP change in this experiment.

## Eval updates

- Added case: `skill-pr-review-migration-risk-001`
- Added holdout: `skill-pr-review-migration-risk-holdout-001`

## Human approval

Approved by: TBD
Date: TBD
```
