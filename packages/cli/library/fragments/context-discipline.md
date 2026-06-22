<context-discipline>

### Context Discipline

Minimize context to reduce cost and improve accuracy.

**Before reading files**:
1. State what you need to know and why.
2. Read the minimum files to answer that question.
3. Stop reading when you have enough context.

**Context budget**:
- Read max 10 files before making a change.
- Prefer focused snippets over full files.
- If you have read 15+ files without a plan, STOP and ask the user.

**Session hygiene**:
- One task per session — start fresh for new tasks.
- If context grows large, summarize progress and offer to continue in a new session.
- Do not re-read files already discussed in this session.

**Priority order** (read in this order, stop when sufficient):
1. Task description and acceptance criteria
2. Files being modified + their tests
3. Type definitions and interfaces
4. Related standards/rules in specs/
5. ADRs only if an architectural decision is needed

**Anti-patterns**:
- ❌ Reading many files "just in case"
- ❌ Re-reading unchanged files from earlier in the session
- ❌ Loading full files when a function signature is enough


**See also**:
- [context-compaction](context-compaction.md) — what to keep after reading
- [Context discipline concept](../../../../docs/concepts/context-discipline.md) — the three practices of context management
</context-discipline>
