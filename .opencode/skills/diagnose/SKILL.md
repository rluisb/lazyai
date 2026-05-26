---
name: diagnose
description: Rigorous 6-phase debugging: build feedback loop → reproduce → rank hypotheses → instrument → fix + regression test → cleanup + post-mortem. Use for hard/non-obvious bugs where guess-and-check fails.
trigger: /diagnose
phase: diagnose
techniques: [chain-of-thought, react]
output_schema:
  sections:
    - Phase 1: Feedback Loop Strategy
    - Phase 2: Reproduction Confirmation
    - Phase 3: Ranked Hypotheses
    - Phase 4: Instrumentation Log
    - Phase 5: Correct Seam + Regression Test
    - Phase 6: Memory Payload
consumes:
  - bug description or issue reference
  - codebase access
produces_for:
  - memory-write (Phase 6 post-mortem)
  - improve-codebase-architecture (optional handoff)
mcp_tools: [filesystem, ripgrep, qmd]
harness:
  feed_forward: [bug description]
  contract: [speckit-review]
  anti_slope: [no-guessing, no-silent-proceed, regression-test-mandatory]
workspace:
  scope: [project]
  reads: [buggy code, tests, relevant modules]
  writes: [memory payload output as phase artifact]
---
## Quick Reference

| | |
|---|---|
| **Use when** | [When to use this skill] |
| **Do not use when** | [When NOT to use this skill] |
| **Primary agent** | [Which agent uses this] |
| **Runtime risk** | [Low/Medium/High] |
| **Outputs** | [What this skill produces] |
| **Validation** | [How to validate output] |
| **Deep mode trigger** | [How to trigger full mode] |



# Diagnose Skill

## Phase 1 — Build a Feedback Loop

Everything else is mechanical. A fast, deterministic, agent-runnable pass/fail signal is required. **MUST NOT proceed without a loop.**

10 strategies (pick best available):
1. Failing test at the seam
2. curl/HTTP script
3. CLI with fixture diff
4. Headless browser (Playwright/Puppeteer)
5. Replay captured trace
6. Throwaway harness
7. Property/fuzz loop
8. Bisection harness
9. Differential loop
10. HITL bash script

If no loop can be built → stop and ask for environment access, captured artifact, or permission to add instrumentation.

## Phase 2 — Reproduce

Run the loop. Watch bug appear. Confirm loop produces the failure mode the **user described** (not a nearby failure). Confirm reproducibility across runs. Do not proceed until reproduction.

**Non-deterministic bugs:** Goal is higher reproduction rate, not clean repro. Loop 100×, parallelize, stress, narrow timing windows until rate >50% is debuggable. Not "non-reproducible" until 100+ iterations attempted.

**Override:** If user explicitly overrides and says proceed without a loop → best-effort mode: document degraded confidence, proceed to Phase 2 with best available signal, note caveat in Phase 6.

## Phase 3 — Hypothesise

Generate 3–5 ranked hypotheses **before** testing any. Format: "If X is the cause, then Y will happen." Each must be falsifiable.

Show ranked list to user (they often have domain knowledge that re-ranks). Proceed with ranking if user is AFK.

## Phase 4 — Instrument

Each probe maps to ONE specific prediction from Phase 3. Change one variable at a time. Preference: (1) debugger/REPL, (2) targeted logs with `[DEBUG-xxxx]` tags, (3) never "log everything". Tag every debug log. Cleanup is a single `grep`.

## Phase 5 — Fix + Regression Test

Write regression test **before** fix, at the **correct seam** — where the bug manifests in real usage, not a shallow single-caller mock-amenable seam. If no correct seam exists → document finding explicitly as architectural problem; do NOT produce false-confidence regression test.

## Phase 6 — Cleanup + Post-Mortem

Original repro no longer reproduces. All `[DEBUG-...]` instrumentation removed. Throwaway prototypes deleted or moved to debug location.

Ask: "What would have prevented this bug?" If answer involves architectural change → offer to hand off to `improve-codebase-architecture`.

Commit/PR message should state the correct hypothesis.

## Memory Payload

Output at end of Phase 6:
```
[MEMORY-PAYLOAD]
Type: lesson
Tags: [bug-pattern, diagnostic-method]
Context: "[Root cause category]: [1-sentence summary]. [Correct seam]: [seam name or 'architectural — no seam available']. [Handoff]: [architectural recommendation or 'none']."
Importance: high (if architectural handoff) | normal
[/MEMORY-PAYLOAD]
```

The orchestrator chain step wrapping this skill will call `memory-write` with this payload.
