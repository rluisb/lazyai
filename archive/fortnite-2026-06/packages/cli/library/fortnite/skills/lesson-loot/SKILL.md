---
name: lesson-loot
description: Self-improvement through human corrections. Detects when a human corrects agent output, extracts the pattern, stores it as a durable lesson in the-vault, and applies it to future work. Every correction becomes permanent knowledge.
trigger: /lesson-loot
triggers:
  - "learn from this"
  - "remember this"
  - "don't do that again"
  - "lesson learned"
  - "capture this correction"
  - "save this lesson"
  - "note this for future"
skill_path: skills/lesson-loot
---

## Quick Reference

| | |
|---|---|
| **Use when** | Capturing human corrections, self-improvement |
| **Do not use when** | Initial implementation, routine tasks |
| **Primary agent** | All agents |
| **Runtime risk** | Low — lesson storage |
| **Outputs** | Durable lessons in the-vault, pattern extractions |
| **Validation** | Pattern applicability, future application |
| **Deep mode trigger** | `/lesson-loot` or human correction received |

# Lesson Loot — Self-Improvement Through Corrections

**Tagline**: Every correction is loot. Never lose a lesson again.

## Purpose

When a human corrects agent output — "no, do it this way", "that's wrong, use X", "don't ever do Y" — that correction is valuable knowledge. Without lesson-loot, it's lost when the session ends. With lesson-loot, every correction becomes a durable lesson stored in the-vault and loaded by storm-scout in future sessions.

**Use when:**
- Human says "no", "wrong", "instead", "don't", "correct way is"
- Human provides explicit guidance that contradicts agent output
- Human shares a constraint, gotcha, or pattern that should persist
- Any workflow phase produces output that gets corrected

## How It Works

```
Human corrects agent output
  ↓
lesson-loot detects correction
  ↓
Extracts: what was wrong, what's right, why
  ↓
Stores as durable lesson in the-vault
  ↓
Logs to truth-chain ledger
  ↓
Future sessions load via storm-scout Phase 0
```

## Correction Detection

lesson-loot auto-triggers when it detects these patterns in human messages:

| Pattern | Example | Confidence |
|---------|---------|------------|
| "no, ..." | "no, use JWT instead of sessions" | HIGH |
| "don't ..." | "don't hardcode the API key" | HIGH |
| "wrong, ..." | "wrong, the auth middleware loads before routes" | HIGH |
| "instead, ..." | "instead, use the existing UserService" | HIGH |
| "correct way is ..." | "correct way is to pass the context, not create a new one" | HIGH |
| "should be ..." | "this should be a POST, not a GET" | MEDIUM |
| "actually, ..." | "actually, we already have a helper for that" | MEDIUM |
| "never ..." | "never call the database directly from a controller" | HIGH |
| "always ..." | "always validate the token before processing" | HIGH |

## Lesson Extraction

When a correction is detected, extract:

### Required Fields
| Field | Purpose | Example |
|-------|---------|---------|
| **Context** | What were you doing? | "Implementing auth middleware in fedora" |
| **Mistake** | What was wrong? | "Used sessions instead of JWT tokens" |
| **Correction** | What's the right way? | "Use JWT with refresh token rotation" |
| **Why** | Why is this the right way? | "Stateless auth required for horizontal scaling" |
| **Source** | Where did this come from? | "Human correction during implement phase" |

### Optional Fields
| Field | Purpose | Example |
|-------|---------|---------|
| **Tags** | Categorization | "auth, jwt, middleware, fedora" |
| **Severity** | How important? | "critical" (would break prod), "high", "normal", "low" |
| **Scope** | Where does this apply? | "fedora auth module", "all Ruby services" |
| **Anti-pattern** | Named failure mode | "Stateful Auth in Stateless Service" |

## Lesson Storage

Lessons are stored in the-vault at `.specify/memory/`:

```
.specify/memory/
  fedora-auth-jwt-over-sessions.md
  school-plan-null-guard-required.md
  mono-frontend-api-timeout-pattern.md
```

Each lesson file follows the-vault template:

```markdown
# [Title — kebab-case, descriptive]

**Context:** [What were you doing when you learned this?]
**Mistake:** [What was wrong?]
**Correction:** [What's the right way?]
**Why:** [Why is this the right way?]
**Source:** [Human correction during {phase} phase]
**Tags:** [comma-separated]
**Severity:** [critical|high|normal|low]
**Scope:** [where this applies]
**Date:** [ISO timestamp]
```

## Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| **auto** | Auto-detect corrections, extract, store | Default — runs in background |
| **manual** | User explicitly invokes `/lesson-loot` | Explicit knowledge capture |
| **review** | Review recent corrections, deduplicate, refine | Periodic cleanup |

## Integration

### With the-vault
- lesson-loot writes lessons to `.specify/memory/`
- Uses the-vault's template format for consistency
- Tags enable cross-session discovery

### With storm-scout
- Phase 0 (Grill Me With Docs) loads relevant lessons before starting work
- Checks `.specify/memory/` for lessons matching current task tags
- Prevents repeating known mistakes

### With truth-chain
- Logs `lesson_extracted` entries with correction details
- Links to the-vault file path
- Enables audit of what was learned and when

### With all agents
- Auto-triggered when human corrects any agent's output
- Works across all workflow phases (clarify, research, plan, implement, verify)
- No agent-specific configuration needed

### With build-mode
- Pre-flight check: load relevant lessons before implementation
- Post-correction: auto-trigger lesson-loot on human feedback

### With zero-point
- Verification failures that get human correction → auto-trigger lesson-loot
- Drift violations that reveal missing invariants → lesson-loot captures the pattern

## Rules

1. **Never ignore a correction** — if a human says "no" or "don't", capture it
2. **Extract the pattern, not the instance** — "use JWT" not "use JWT in this specific endpoint"
3. **Tag aggressively** — more tags = better future discovery
4. **Deduplicate** — if a similar lesson exists, update it rather than creating a duplicate
5. **Respect severity** — critical corrections (would break prod) get highest priority
6. **Log everything** — every extraction goes to truth-chain ledger
7. **Auto-apply** — future sessions automatically load relevant lessons

## Output

When lesson-loot captures a correction, it reports:

```
🎒 Lesson Looted!
   Context: Implementing auth middleware in fedora
   Mistake: Used sessions instead of JWT tokens
   Correction: Use JWT with refresh token rotation
   Stored: .specify/memory/fedora-auth-jwt-over-sessions.md
   Tags: auth, jwt, middleware, fedora
   Severity: critical
```

## Anti-Patterns

| Pattern | Why Wrong | Fix |
|---------|-----------|-----|
| "I'll remember this" | You won't — sessions end, context is lost | Always capture to the-vault |
| "This is too specific" | Specific lessons are the most valuable | Tag with scope, let discovery filter |
| "It was just a small correction" | Small corrections compound into big patterns | Capture everything, deduplicate later |
| "The agent will figure it out" | Agents don't learn across sessions without storage | lesson-loot bridges the gap |
