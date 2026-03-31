# Lessons Learned

## Purpose

Extract non-obvious, reusable learnings from a completed task. Not a summary. Not a retrospective. Learnings only.

Maximum 3 per task. Zero is valid and often correct.

## When Active

Terminal skill — runs LAST in the skill chain, always AFTER memory-write completes.

## Skill Chain

```
research → plan → implement → iterate → memory-write → lessons-learned → task closed
```

**CRITICAL**: Never run before memory-write completes.

## Execution Flow

1. Read task journal and notes from the completed task
2. Identify candidates: surprises, abandoned paths, late discoveries
3. Apply YAGNI filter to each candidate
4. Keep maximum 3 highest-value lessons (zero is valid)
5. Write each lesson to `docs/memory/lessons/` directory

## YAGNI Filter

A lesson passes ONLY if ALL THREE are true:

1. **Future value**: A future agent would change its approach based on this lesson
2. **Non-obvious**: It is not obvious from reading the code, docs, or ticket alone
3. **Actionable**: It is specific and actionable — not generic advice like "write tests first"

## Output Format

Write each lesson as a file in `docs/memory/lessons/`:

```
docs/memory/lessons/YYYY-MM-DD-slug.md
```

Each file contains:
```markdown
# Lesson: {title with ticket ID}

**Trigger**: What caused the problem or surprise
**Impact**: What went wrong or was unexpected
**Prevention**: How to prevent this in the future

Keywords: {comma-separated tags}
```

Maximum 10 lines per lesson file.

## Anti-Patterns

- Writing generic lessons: "always check existing scopes", "read tests before coding"
- Batching two learnings into one file
- Writing a lesson that's already captured in decisions or docs
- Running this skill before memory-write has completed
- Padding with low-value lessons to reach 3 — zero is correct behavior

## Integration

- **Documenter agent**: Primary executor of this skill at task end
- **Builder agent**: Can trigger during implementation if a significant surprise occurs
- **Self-Improvement Protocol**: Lessons feed the memory promotion path (lesson → repeated 2x → standard/rule → delete lesson)
