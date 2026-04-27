<rule>
  <scope>auto</scope>
  <globs>specs/memory/**</globs>
  <description>Memory files — patterns and learnings captured by agents for future sessions</description>
</rule>

# Memory Rules

## What Goes Here
Patterns, gotchas, and learnings discovered by agents during work that should persist across sessions.

## Rules
- Memory files are suggestions, not commands. They inform, they don't enforce.
- To enforce something: move it to specs/rules/ via PR.
- Agents MAY write memory files. Humans MUST review them.
- Keep each file under 50 lines. One topic per file.
- Delete memory files that have been promoted to rules or standards.

## Multi-Session Handoffs

Use a dedicated handoff directory for cross-session continuity:

- Path: `specs/memory/handoffs/`
- Naming: `YYYY-MM-DD-[topic].md`
- Read the latest handoff at session start before planning

Handoff file should capture:
1. Current objective and status (done/in-progress/blocked)
2. Decisions made and rationale
3. Open assumptions/questions
4. Next 1–2 concrete actions
5. Risks/watchouts for the next session

## Examples of Memory Content
- "The payments module has a 5-second timeout on webhook processing — discovered during bugfix 004"
- "When adding a new entity to the billing module, also update the billing.module.ts providers array"
- "The CI pipeline fails silently if .env.test is missing — always check"

## Memory Is Not
- Not a replacement for rules (rules are enforced, memory is advisory)
- Not a replacement for standards (standards show patterns, memory shows gotchas)
- Not a changelog (use git for that)

## Promotion Path

Memory files are temporary. They either get promoted or deleted:

```
Memory note written
    │
    ├── Pattern is repeated 2+ times → promote to specs/standards/ → delete memory
    ├── Rule is needed to prevent issue → promote to specs/rules/ → delete memory
    ├── Info is now outdated → delete memory
    └── Still advisory after 30 days → keep or re-evaluate
```

When promoting: note in the standard/rule that it originated from a memory insight.

## Self-Improvement
- When a memory note is promoted to a rule or standard → delete the memory note
- When a memory note is older than 30 days → re-evaluate: still relevant? promote or delete
- When a new category of memory notes emerges → consider if a new standard category is needed
