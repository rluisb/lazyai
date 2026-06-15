---
name: reboot-van
description: "Root-cause investigation and disciplined fix loop. Two modes: diagnose (find what killed you — hypothesis-driven investigation) and iterate (respawn the fix — known root cause, one change per cycle). Never blind-guess. Never shotgun-fix."
trigger: /reboot-van
---

# Reboot Van


## Tool Selection

Use the right tool for each job. See skills/_tool-hierarchy.md for full decision tree.

| Task | Tool |
|------|------|
| Read known file | OpenCode Read |
| Find code by description | morph codebase_search |
| Symbol analysis | codegraph MCP |
| Vault search | qmd MCP |
| Architecture overview | graphify CLI |


## Purpose
Find what killed your code and bring it back. Two modes in one skill:

- **Diagnose mode** — root cause unknown. Hypothesis-driven investigation following the Disciplined Diagnose Loop.
- **Iterate mode** — root cause known. Compact fix loop: one targeted change per iteration with mandatory reflexion.

**The difference:** diagnose = unknown root cause, investigation mode. iterate = known root cause, fix mode. If you're unsure which mode you're in, you're in diagnose.

---

# Mode: Diagnose

Systematically investigate an unknown problem through hypothesis-driven, evidence-validated steps. Every action is linked to a hypothesis. No "let me try a few things."

## Tooling — Use `codebase_search` (WarpGrep)

Use `codebase_search` (WarpGrep) in **Step 4: Instrument** to locate code relevant to the hypothesis without manual file hunting.

Best for: "Find where this error is raised", "How is this value computed?", tracing call paths.
Not for: log file searches — use bash `grep` against captured output, reading known files — use OpenCode `Read`.

## Process

### Step 1: Build the Feedback Loop
Make the failure **observable and reproducible** before doing anything else.

- Can you see the failure clearly? (logs, error output, test failure, network trace)
- Is the failure deterministic? If not — why? (flaky test? race condition? environment?)
- Do NOT proceed until you have a reliable signal.

If you cannot observe the failure, your first task is to instrument visibility — not to fix.

### Step 2: Reproduce
Confirm you can trigger the failure on demand.

- Write the minimal reproduction case
- If you cannot reproduce it: stop — investigate **why** you can't reproduce before attempting any fix
- A fix without reproduction is guesswork

### Step 3: Hypothesise
Write down your top 3 hypotheses, ranked by likelihood. **No fixing yet.**

```
Hypothesis 1 (most likely): [statement of cause]
  — Evidence for: [why you believe this]
  — Evidence against: [what would refute it]

Hypothesis 2: [statement of cause]
  — Evidence for: ...
  — Evidence against: ...

Hypothesis 3: [statement of cause]
  — Evidence for: ...
  — Evidence against: ...
```

### Step 4: Instrument (one hypothesis at a time)
Target Hypothesis 1 only. Add the **minimum instrumentation** needed to collect evidence:
- What would confirm this hypothesis?
- What would refute it?
- Add only that. Nothing else.

### Step 5: Evaluate Evidence
Collect evidence. Assess:
- **Confirmed** → proceed to Step 6
- **Refuted** → update hypothesis list, move to next, return to Step 4
- **Inconclusive** → add more targeted instrumentation, stay in Step 4

### Step 6: Fix
Only after a hypothesis is confirmed by evidence:
- Make the **minimal targeted fix**
- One change at a time
- State: "Fixing [confirmed hypothesis] by [specific change]"
- Use `edit_file` for the fix — one call, minimal diff

### Step 7: Regression Test
Write (or identify) a test that would have caught this issue.

- If a test already existed and didn't catch it: understand why before moving on
- The goal is: this bug cannot sneak back undetected

## Diagnose Escalation
After **5 hypothesis cycles** with no confirmed root cause: STOP.

Escalate to human with:
- Full investigation log (what you tested, what evidence you got)
- Current best hypothesis and why it's unconfirmed
- What you would need to confirm or refute it

Never silently loop past 5 cycles.

---

# Mode: Iterate

Efficiently fix known failing tests or known issues through a disciplined loop. Every iteration is exactly one targeted change followed by verification.

## Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| MAX_ITERATIONS | 5 | Maximum fix attempts before escalating |

Override at invocation: include `MAX_ITERATIONS=N` in your prompt.

## Before Starting: Confirm Root Cause
1. Read the **full** error output — not just the last line
2. Understand exactly what is failing and why
3. Ask yourself: "Do I actually know the root cause?"

If the answer is no → **switch to Diagnose mode**. Iterating on an unknown root cause wastes cycles.

## Each Iteration

**Read:**
- Full error message (scroll up — the real cause is often not the last line)
- The failing test or assertion
- The code path that leads to the failure

**Plan:**
State in one sentence what you will change and why:
> "I am changing [X] because [it is the confirmed cause of Y]."

**Make one change:**
- One targeted change per iteration
- No refactoring alongside the fix
- No opportunistic improvements
- No "while I'm here" changes
- Use `edit_file` — one call, minimal diff

**Verify:**
- Run the quality gates
- Did the targeted failure resolve?
- Did any new failures appear?

**Reflexion (mandatory after each iteration):**
```
Iteration N:
- What I changed: [specific change]
- Why: [link to root cause]
- Result: [resolved / new error / same error]
- Next: [why I think next change will work, OR escalate reason]
```

### Assess After Each Iteration
- **Pass:** done ✓
- **New failures appeared:** understand them before planning next iteration — don't just fix symptoms
- **Same error:** your change didn't address the root cause → reconsider root cause; switch to Diagnose mode
- **Progress stalling after 2 iterations with same error:** stop, switch to Diagnose mode

## Escalation Triggers

**Escalate immediately (before MAX_ITERATIONS) if:**
- Same error persists after 2 different targeted fixes → root cause was wrong
- Each fix reveals a deeper issue → systemic problem, not a patch job

**Escalate after MAX_ITERATIONS:**
Report to human with:
- Full reflexion log (every iteration: what, why, result)
- Current error state
- Your best hypothesis for why it persists
- What you would need to confirm or refute that hypothesis

Never silently loop past MAX_ITERATIONS.

---

## Code Editing & Exploration — Use morph-mcp

For all fix steps in both modes, use `edit_file` (morph-mcp Fast Apply):

- Provide only the changing section with `// ... existing code ...` markers
- One `edit_file` call per change — keep it minimal and targeted
- Speed: 10,500+ tok/s with 98% accuracy — low overhead for rapid fix cycles

---

## Mode Selection Guide

| Situation | Mode | Why |
|-----------|------|-----|
| "Something is broken, I don't know why" | Diagnose | Root cause unknown |
| "Tests are failing after my change" | Iterate | Root cause is your change |
| "CI is red, error is clear" | Iterate | Error tells you exactly what's wrong |
| "Production is down, symptoms are vague" | Diagnose | Need to build feedback loop first |
| "Fixed it 3 times, still failing" | Diagnose | Your root cause assumption was wrong |

---

## Rules (both modes)
- Never fix without a confirmed hypothesis (diagnose) or confirmed root cause (iterate)
- Never make multiple changes simultaneously — one change per cycle
- Never skip reproduction — if you can't reproduce, investigate why first
- Every action must be linked to "I am testing whether [hypothesis]" or "I am fixing [cause]"
- No "let me just try X" — X must be a hypothesis test or a targeted fix, not a guess
- Read the full error before every cycle — not just the last line
- Reflexion step is mandatory in iterate mode — it prevents zombie iterations
- Never exceed cycle limits without escalating
- Progress is evidence; absence of progress is a signal to stop and switch modes
