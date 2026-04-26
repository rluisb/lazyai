---
name: memory-write
description: Write persistent context and decisions for future sessions.
trigger: /memory-write
phase: post-task
---

# Memory Write Skill

## When to Write Memory
- A non-obvious gotcha was discovered during implementation
- A pattern decision was made that future sessions need to know
- A workaround was applied that should be documented
- A lesson was learned that prevents repeating a mistake

## Workflow
1. Identify the insight — one concept per memory file
2. Write to specs/memory/ — max 30 lines, include real file references
3. Categorize: decisions/ | patterns/ | handoffs/
4. Check for promotion — if this pattern repeats 2+ times, promote to specs/standards/

## YAGNI Gate
Before writing, ask: "Will this help a future session? Or am I just documenting for the sake of it?"
If the insight is obvious from the code itself, skip the memory file.

## Promotion Path
- Pattern repeated 2+ times → promote to specs/standards/ → delete memory
- Rule needed to prevent issue → promote to specs/rules/ → delete memory
- Still advisory after 30 days → re-evaluate: keep or delete

## Integration
- Agent: any (typically Builder or Documenter)
- Triggered: after completing a task or discovering a non-obvious insight
