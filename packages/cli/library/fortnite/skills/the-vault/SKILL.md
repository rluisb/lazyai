---
name: the-vault
description: Store your most valuable knowledge loot. Write durable project knowledge to .specify/memory/ — long-term institutional memory that survives sessions, agents, and people. Distinct from slurp-juice session checkpoints.
trigger: /the-vault
---

# The Vault

## Purpose
Persist lessons, constraints, and architectural decisions that future agents and humans should know about this codebase — without having to re-derive them from scratch. Like the Vault in Fortnite — store your most valuable loot where you can retrieve it later.

**This is project long-term memory.** It lives on disk and outlives any session.

**This is NOT session state.** For preserving task state within or between sessions, use `slurp-juice`.

**lesson-loot feeds the vault.** When a human corrects agent output, lesson-loot auto-extracts the pattern and stores it here as a durable lesson.

---

## Two types of memory — do not confuse them

| Type | Skill | Lives in | Audience | Lifecycle |
|------|-------|----------|----------|-----------|
| Session state | `slurp-juice` | Context window | Same agent resuming | Task duration |
| Project knowledge | `the-vault` | `.specify/memory/` | Any future agent | Indefinitely |

---

## What to save

Write a memory entry when you discover something that a future agent would benefit from knowing:

- **Non-obvious constraint** — a rule that isn't visible in the code or docs
- **Recurring gotcha** — something that has caused or would cause wasted cycles if forgotten
- **Enforced pattern** — a pattern the codebase requires that isn't explained anywhere
- **Architectural decision** — a choice made with specific rationale that future work must respect
- **Hard-won lesson** — something you learned the expensive way in this session

## What NOT to save

- Transient debug notes — session-specific noise with no lasting value
- Code-visible facts — if a future agent can read it in the code in 10 seconds, don't duplicate it
- Obvious things — what every competent developer would know
- Speculative concerns — "might be a problem if..." belongs in a checkpoint risk flag, not the vault

## Smart filter test

Before writing, ask: **"If a new agent opened this codebase cold in 6 months, would this entry save meaningful time?"**
- Clear yes → write it
- Unsure → write it, mark `[unverified]`
- Clear no → skip it

---

## Memory Format

```markdown
## [Short descriptive title — specific, not generic]

**Context:** [What were you doing when you learned this?]
**Finding:** [The lesson — specific and actionable, not vague]
**Why it matters:** [What would go wrong without this knowledge]
**Source:** [Where you learned it — file path, error, conversation, doc]
**Date:** YYYY-MM-DD
```

---

## Storage Path

**Default:** `.specify/memory/[kebab-case-title].md`

| Project type | Path |
|-------------|------|
| Speckit project | `.specify/memory/` |
| Non-speckit project | `.ai-memory/` |
| Unsure | Ask the human |

One file per lesson. Filename should be the kebab-case title: `school-id-scoping-required.md`, `fedora-auth-middleware-load-order.md`.

---

## When to write memories

- After a non-obvious constraint is discovered during research
- After a recurring error reveals something structural about the codebase
- After an architectural decision is made with rationale worth preserving
- At the **end of a complex multi-session task** — before closing, ask: "What should the next session know?"

---

## Rules

- Only save what you're confident is accurate — a wrong memory is worse than no memory
- Mark uncertain entries `[unverified]` rather than skipping them
- Keep entries short and actionable — if it needs a paragraph of context, write two entries
- Never duplicate what's clearly visible in the code
- When in doubt: save it, mark it, let a human validate
- The Vault is precious. Don't fill it with common loot.
