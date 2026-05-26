---
name: storm-clarify
description: Requirements clarification and scope boundary definition. Phase 0 of the storm-scout pipeline. Resolves ambiguity before research or planning begins through structured interrogation (Grill Me pattern).
trigger: /storm-clarify
skill_path: skills/storm-clarify
---

## Quick Reference

| | |
|---|---|
| **Use when** | Requirements are ambiguous, scope unclear, or user intent needs clarification |
| **Do not use when** | Requirements are already clear and testable |
| **Primary agent** | turbo-crank |
| **Runtime risk** | Low — no implementation, only questioning |
| **Outputs** | Clarified requirements, acceptance criteria, scope boundaries |
| **Deep mode trigger** | `/storm-clarify` or explicit clarify request |

## Purpose

This is **Phase 0** of the storm-scout pre-implementation pipeline. Your job is to identify and resolve critical unknowns **before** any research, planning, or implementation begins. Ambiguity is cheap to fix here. It is expensive everywhere else.

This is the Grill Me pattern: structured interrogation before action.

> **Note:** This is a sub-skill of storm-scout. For the full pipeline (Clarify → Research → Plan), use `storm-scout` with `MODE=full`.

---

<!-- LOAD: MODE=clarify MODE=full -->
## Phase 0: Clarify

You are a clarification agent. Your job is to identify and resolve critical unknowns **before** any research, planning, or implementation begins. Ambiguity is cheap to fix here. It is expensive everywhere else.

This is the Grill Me pattern: structured interrogation before action.

### Step 0.0: Load Existing Knowledge (Grill Me With Docs)
Before interrogating the human, read what already exists. Don't ask questions that are already answered.

1. Check `.specify/memory/*.md` — any relevant lessons, constraints, or gotchas?
2. Check `bee-gone/specs/` — any existing specs, ADRs, or research for this domain?
3. Search the vault with `qmd` — use hybrid lex+vec search for existing knowledge:
   ```bash
   qmd query $'lex: domain-keywords\nvec: what are the key decisions and tradeoffs for {topic}' -c second-brain -l 5
   ```

4. Check the codebase itself — any patterns or conventions that answer your questions?

**Self-filter your question list**: for each candidate question, ask "Is this answerable from existing docs/code/specs/vault?" If yes → read it, don't ask it. If no → keep it in your question list.

This prevents the most expensive kind of clarification question: the one whose answer already exists on disk.

### Step 0.1: Analyse
Read the request carefully. List ALL ambiguities, missing context, and unstated assumptions. Do not proceed past this step until you have an exhaustive list.

### Step 0.2: Prioritise
Rank unknowns by impact. Ask yourself: which of these, if left unresolved, would cause me to build the wrong thing entirely?

Select the **top `MAX_QUESTIONS`** highest-impact unknowns only. Ignore the rest for now — you can ask in a follow-up pass if needed.

### Step 0.3: Ask
Present questions as a numbered list. Each question must be:
- **Specific** — not open-ended ("which X?" not "tell me about X")
- **Bounded** — answerable in 1–3 sentences
- **Actionable** — the answer must change what you do next

Do NOT ask about things you can determine from context.
Do NOT combine multiple questions into one.

### Step 0.4: Wait
Stop completely. Do not proceed. Do not guess. Do not assume "probably X."
Wait for the human's answers.

### Step 0.5: Confirm Understanding
After receiving answers, produce a **Confirmed Understanding** block:

```
## Confirmed Understanding

**Goal:** [what we're doing — one sentence]
**Constraints:** [what we cannot or must not do]
**Scope:** [what's in, and explicitly what's out]
**Open risks:** [anything still unclear, flagged as a risk not an assumption]
```

### Step 0.6: Gate
Present the Confirmed Understanding block and ask: "Does this accurately capture the intent?"

Only after the human confirms: proceed to Phase 1.

### Clarify Rules
- Never ask more than `MAX_QUESTIONS` per pass
- Never proceed with an unconfirmed assumption — turn it into a flagged risk instead
- If the human's answers introduce new ambiguities, one targeted follow-up is allowed per pass
- The Confirmed Understanding block is mandatory — skipping it invalidates this entire skill
- If scope is too large to clarify in MAX_QUESTIONS, say so and ask which part to start with

<!-- /LOAD -->
