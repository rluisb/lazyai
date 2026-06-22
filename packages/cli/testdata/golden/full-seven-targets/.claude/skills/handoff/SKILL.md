---
name: handoff
description: Use when a session is ending, context needs to be preserved for a future session, or when transferring work between agents. Generates a structured handoff document with open questions, next steps, and context summary.
---

# Handoff

## When to Use

Use this skill when:
- A session is ending and work is incomplete
- You need to preserve context for a future session (yours or another agents)
- Transferring work between agents in a multi-agent workflow
- The user says "pick this up later" or "continue tomorrow"
- A long-running task needs to be checkpointed

Do not use for trivial completions (single file edit, quick fix). Do not use when the work is fully done — a simple summary in chat is sufficient.

## Rule

A handoff is a single markdown document stored in the durable memory system. It must answer three questions for the next agent: What was done? What is blocked? What should happen next?

## Workflow

1. **Assess handoff necessity**
   - Is there uncommitted work? → Handoff needed
   - Are there pending decisions? → Handoff needed
   - Are there open branches/PRs? → Handoff needed
   - Was this a single quick fix? → Skip handoff, summarize in chat

2. **Gather context**
   - List files modified (with git status)
   - Identify open questions or decisions
   - Note which tools were used
   - Capture the current working directory

3. **Write the handoff document** with this exact structure:

```markdown
# Handoff: [Brief Title]

**Date:** [ISO 8601]
**From:** [agent/session identifier]
**Status:** [in_progress | blocked | pending_review]

## Work Completed

- [ ] [Specific accomplishment with file path]
- [ ] [Another accomplishment]

## Open Questions

1. [Question that needs answering before next step]
2. [Another question]

## Next Steps

1. [ ] [Concrete next task with acceptance criteria]
2. [ ] [Another task]

## Context for Next Agent

- **Branch:** [current git branch]
- **Uncommitted changes:** [yes/no + brief description]
- **Blocked by:** [whats blocking progress, or "none"]
- **Key files:** [list of files the next agent must read first]

## Technical Notes

- [Tool/dependency version if relevant]
- [Configuration that changed]
- [Known issues or gotchas]
```

4. **Store the handoff**
   - If `ai-memory` or similar durable store is available: store there with tag `handoff-[date]`
   - If no durable store: save to `specs/memory/handoffs/[YYYY-MM-DD]-[title].md`
   - If the repo has a specific handoff location per project conventions, use that

5. **Notify**
   - If another agent is expected to pick up: mention the handoff location
   - If this is for your own future session: confirm storage location
   - If work is truly blocked: mark as blocked and explain why

## Constraints

- One handoff per session. Do not create multiple handoffs for the same session.
- Handoffs must be actionable. "Continue work" is not enough — specify exactly what "continue" means.
- Include file paths. The next agent must know which files to read without guessing.
- State blockers honestly. If you dont know why something failed, say so.
- No speculation in "Work Completed" — only facts about what was actually done.

## Verification Checklist

- [ ] Handoff document has all five sections (Work Completed, Open Questions, Next Steps, Context, Technical Notes)
- [ ] Status is one of: `in_progress`, `blocked`, `pending_review`
- [ ] Date is ISO 8601 format
- [ ] At least one Next Step has a concrete acceptance criterion
- [ ] Key files list includes paths relative to repo root
- [ ] Handoff is stored in durable location (not just chat history)

## Related Skills

- `memory-promotion` — For elevating ephemeral notes to durable knowledge
- `skill-authoring` — For turning handoff patterns into reusable skills