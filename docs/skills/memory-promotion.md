---
name: memory-promotion
description: Use at task closeout to propose durable ai-memory or documentation updates without writing silently, especially when caveman summaries, diagnoses, triage, or issue extraction reveal reusable knowledge.
---

# Memory Promotion

## When to Use

Use this skill at the end of a task when the work revealed a reusable convention, gotcha, operational fact, root cause, template, or project decision.

## Rule

Never write durable memory silently. Ask for approval with target, reason, and draft.

## Promotion Criteria

Propose promotion only when the lesson is likely to matter again:

- repeated setup or repair step;
- project-specific convention not already documented;
- sharp edge that caused debugging time;
- architecture or workflow decision with future consequences;
- validation command that should become default for similar work;
- caveman summary that contains a stable reusable insight.

Do not promote one-off task details, transient status, or information already present in docs.

## Caveman Bridge

Use `canonical/caveman-ai-memory.md`.

- Caveman may compress the working thread.
- Memory must preserve enough context to be valuable later.
- Promote the reusable 10%, not the whole compressed summary.
- Include source, evidence, decision, scope, and expiry/removal condition when relevant.

## Proposal Format

Before writing anything, ask:

```text
Should I promote this to memory?
```

Include:

```md
Target: <.ai-memory.toml | ai-memory entry | docs/path.md | canonical/path.md>
Classification: <rule | template | trap | pattern>
Reason: <why this will recur>
Source: <where this came from>
Draft:
- <fact with enough context to reuse>
```

## Human Gate

Only write after explicit approval.

For ai-memory-backed project knowledge, use the configured memory tool or `.ai-memory.toml` only as directed by project docs. For documentation, update the nearest existing doc instead of creating a new file unless a new durable topic is needed.

## Constraints

- Never write memory silently.
- Never use a handoff or caveman summary as permanent memory without extracting context-rich facts.
- Never promote secrets, credentials, or machine-local paths unless the user explicitly asks and the target is private.
