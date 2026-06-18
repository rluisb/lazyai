---
name: slurp-juice
description: Mandatory session checkpoint and handoff protocol. Keeps your state regenerating through the storm so you never lose progress. Loss of execution state is elimination.
trigger: /slurp-juice
skill_path: skills/slurp-juice
scripts:
  - name: rtk-watch.sh
    description: Continuous context monitor for session state
    path: scripts/rtk-watch.sh
  - name: session-reset.sh
    description: Pre-session health check and state reset
    path: scripts/session-reset.sh
  - name: knowledge-inject.sh
    description: Proactive vault context injection at session start
    path: ../../scripts/knowledge-inject.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | Session checkpoint, handoff, resume |
| **Do not use when** | Durable project memory (use the-vault) |
| **Primary agent** | all agents |
| **Runtime risk** | Low — state preservation |
| **Outputs** | Checkpoint files, handoff state, resume prompts |
| **Validation** | State completeness, restore success |
| **Deep mode trigger** | `/slurp-juice` or session transition |

# Slurp Juice

## Purpose
Create compact, lossless checkpoints and handoffs that preserve task state across context resets, session pauses, and agent handoffs. Like slurp juice in Fortnite — it keeps your shields up through the storm.

**This skill is mandatory, not optional.** See trigger points below.

Compact wording is good. Loss of execution state is elimination.

---

## Scripts

This skill owns the following scripts:

| Script | Purpose |
|--------|---------|
| `rtk-watch.sh` | Continuous context monitor — watches for checkpoint triggers |
| `session-reset.sh` | Pre-session health check — validates state before resuming |
| `knowledge-inject.sh` | Proactive vault context injection — loads vault knowledge before work begins |

Run from skill directory: `./scripts/rtk-watch.sh` or `./scripts/session-reset.sh`

---

## Mandatory Trigger Points

You MUST produce an RTK output at each of these points — do not wait to be asked:

| Trigger | Action | When |
|---------|--------|------|
| **Start** | `/task-context` | Before doing anything on a non-trivial task |
| **Transition** | `/rtk-check` | After research, plan, implement, or verify completes; or when a blocker appears |
| **End** | `/rtk-handoff` | Before pausing work or when session is running long |

If unsure whether a task is "non-trivial": checkpoint anyway. A checkpoint you didn't need costs 30 seconds. A missing checkpoint that causes rework costs hours.

### In-flight Drift Check (recommended)

Between mandatory trigger points, perform a lightweight 3-question drift self-check at natural breakpoints (~every 10 tool calls, before writing files, at phase boundaries):

1. Am I still aligned with the spec/task/done-condition?
2. Have I drifted into scope creep or speculation?
3. Should I checkpoint now (per slurp-juice triggers)?

---

## Non-negotiable Fields

Never omit these in any checkpoint or handoff:

| Field | Rule |
|-------|------|
| Objective / task | required |
| Current state | required |
| Decisions made + why | required |
| Exact error text | quote or "none" |
| Next concrete action | required |
| Resume prompt | required |

Write `unknown` for any field you cannot determine. Never invent missing facts.

## Output Contract

Every checkpoint and handoff must preserve in this order:

1. **Goal** — what are we trying to accomplish?
2. **Current state** — where are we right now?
3. **Decisions + why** — what was decided and the rationale
4. **Errors / blockers** — exact error text, nothing paraphrased
5. **Next action** — the single most important next step
6. **Resume prompt** — exact phrasing to hand to a fresh agent to resume

## Restore Usefulness Metric

After resuming from a checkpoint or handoff, evaluate whether the checkpoint actually helped:

| Metric | How to measure |
|--------|----------------|
| **Time to resume** | Minutes from session start to first productive tool call |
| **Context gaps** | Number of follow-up questions needed to recover missing state |
| **Rework avoided** | Did the checkpoint prevent re-running prior commands? |

If a checkpoint is loaded but the resuming agent still asks "what were we doing?", the checkpoint failed. Simplify and sharpen the resume prompt.

---

## Compression Rules

- Keep exact paths, commands, IDs, errors, branch/worktree references
- Prefer bullets over prose
- Remove narrative fluff — preserve causal context
- If something happened for a reason, keep the reason, not just the fact

---

## Safety

- Read-only summaries only — no mutations, no fixes in a checkpoint
- No invented file lists
- No invented approvals
- No speculative fixes

---

## What this skill does NOT do

Writing durable project lessons (non-obvious constraints, architectural decisions, recurring gotchas) to `.specify/memory/` is handled by the **`the-vault` skill** — that is long-term project knowledge, not session state. Load `the-vault` at the end of a task when something worth preserving was discovered.
