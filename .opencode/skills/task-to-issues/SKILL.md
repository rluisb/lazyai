---
name: task-to-issues
description: Use when extracting actionable tasks from meeting notes, Slack threads, PR descriptions, specs, or other unstructured text. Converts ephemeral notes into tracked issues with context, acceptance criteria, deduplication, and learning capture.
---

# Task to Issues

## When to Use

Use this skill when:
- The user pastes notes and asks to create tickets or track them.
- A thread, PR, RFC, retro, or spec contains multiple action items.
- The user says something should not fall through the cracks.
- A PRD task list needs tracked implementation issues.

Do not use for a single obvious task or text with no concrete action items.

## Rule

An issue without context is a to-do that will be ignored. Every extracted task must include what to do, why it matters, and how to verify it.

## Workflow

1. Read the full source material.
2. Classify each item as task, decision, observation, or question.
3. Enrich each task with title, context, acceptance criteria, priority, assignee if known, and source.
4. Deduplicate against existing tracked work when a tracker is available.
5. Create or draft one issue per task unless tasks are tightly coupled.
6. Document source material to issue mapping and items not filed.

## Issue Template

```markdown
Title: <Imperative verb + object>

Context:
<1-2 sentences from the source material.>

Acceptance Criteria:
- [ ] <Observable done condition.>

Priority: <P0 | P1 | P2>
Assignee: <name | unknown>
Source: <link or note reference>
```

## Learning Capture

Use `canonical/learning-template.md` after completion when extraction reveals a reusable classification rule, template, trap, or workflow pattern.

Store raw findings in `specs/memory/sessions/learning-YYYY-MM-DD-<slug>.md`. Promote durable facts only through `memory-promotion`.

## Constraints

- Do not create vague issues without acceptance criteria.
- Do not batch unrelated tasks into one issue.
- Do not create duplicates; reference existing issues instead.
- Ask one focused clarification when source material is ambiguous enough to change issue scope.

## Verification Checklist

- [ ] Every task has title, context, and acceptance criteria.
- [ ] Duplicate search was performed or marked unavailable.
- [ ] P0 items are flagged.
- [ ] Source-to-issue mapping is documented.
- [ ] Reusable learnings were captured or intentionally skipped.

## Related Skills

- `issue-triage` — classify incoming issues after creation.
- `diagnose` — investigate confirmed bugs.
- `memory-promotion` — promote reusable learnings after approval.
