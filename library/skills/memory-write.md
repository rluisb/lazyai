---
name: memory-write
description: Capture decisions and knowledge for future sessions.
trigger: /memory-write
phase: close
---

# Memory Write

## Purpose

Capture knowledge atomically as it is generated — during and after implementation. One concept per file. Max 30 lines. Never batch, never delay.

## When Active

Triggered during implementation and at task completion. Runs BEFORE lessons-learned skill.

## Memory Directory Structure

All memory artifacts live under `specs/memory/` (gitignored except index files):

```
specs/memory/
├── decisions/       # Why we chose X over Y
├── docs/            # Domain knowledge
├── tech/            # Patterns, gotchas, workarounds
└── tasks/
    └── questions/   # Non-obvious answers to questions
```

## Triggers and Actions

| Event | Action | Target |
|-------|--------|--------|
| Decision made | Write decision file | `specs/memory/decisions/YYYY-MM-DD-slug.md` |
| Domain knowledge | Write domain doc | `specs/memory/docs/topic.md` |
| Pattern or gotcha | Write tech pattern | `specs/memory/tech/slug.md` |
| Question answered (non-obvious) | Write answer file | `specs/memory/tasks/questions/TICKET-slug.md` |
| Path abandoned | Append journal line | Task file journal section |
| Task complete | Update task status | Task file: status=done + files list |

## YAGNI Gate

Before every write, ALL THREE must be true:

1. A future agent would change its behavior based on this information
2. It is not obvious from reading the code or existing docs alone
3. It is specific and actionable — not generic engineering advice

If any check fails → do not write. Silence is correct behavior.

## File Size Rule

- Maximum **30 lines** per file
- One concept per file
- If a file would exceed 30 lines → split into two files first
- Never update a file to cover a second concept — create a new file instead

## Journal Format

Append-only section in task files:

```
[HH:MM] event-type: description (≤15 words)
```

Examples:
```
[09:14] decided: exclude fully_paid_plan from recurring count query
[09:31] tried: use Sale.subscription scope — abandoned: excludes in-progress
[09:45] discovered: cancellable_by_student? already returns false for fully_paid
```

## Anti-Patterns

- Batching multiple decisions into one write
- Writing generic advice ("always check scopes")
- Duplicating what's already in code comments or docs
- Skipping the YAGNI gate
- Writing memory files after the task is fully closed (use lessons-learned instead)

## Skill Chain Position

```
research → plan → implement → iterate → memory-write → lessons-learned → task closed
```

Memory-write runs during and after implementation. Lessons-learned runs ONLY after memory-write completes.

## Integration
- **Primary agent:** Closer (knowledge capture)
- **Triggered by:** `/memory-write` command and task completion events
- **Depends on:** Decisions, task journal entries, and non-obvious implementation context
- **Feeds into:** `lessons-learned` and durable memory for future sessions
