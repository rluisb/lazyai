---
name: research-rpi
description: Bounded research for RPI feature workflow. Produces structured findings that feed into the plan phase.
argument-hint: "[topic-or-question]"
trigger: /research
phase: research
techniques: [chain-of-thought, trace-protocol]
output: specs/features/NNN-name/research.md
output_schema:
  sections:
    - Research Question (what we need to understand)
    - Scope (what is and isn't in scope)
    - Findings (structured answers to research question)
    - Patterns Identified (how codebase handles similar things)
    - Impact Assessment (what could be affected by changes)
    - Trace Log (Thought/Action/Observation/Decision per step)
    - Recommendations (what the plan should incorporate)
consumes:
  - specs/features/NNN-name/spec.md (feed-forward from specify phase)
  - existing standards and ADRs
  - KNOWLEDGE_MAP.md
produces_for:
  - plan skill (direct feed-forward)
  - memory (if new patterns discovered)
mcp_tools: [filesystem, ripgrep, qmd]
harness:
  feed_forward: [spec.md from specify phase]
  contract: [findings-structured, trace-logged]
  sensors: [gate-2]
  memory: [ledger.md if new patterns found]
  anti_slope: [no-silent-research, scope-locked]
workspace:
  scope: [project]
  reads: [codebase, specs, standards, ADRs, KNOWLEDGE_MAP.md]
  writes: [specs/features/NNN-name/research.md]
  cross_repo: false
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



# Research (RPI-Bounded) Skill

## When to Use
- RPI workflow: after specify phase, before plan phase
- Explicit research question from spec.md
- Need to understand existing patterns before planning implementation

## Workflow
1. **Clarify scope** — read spec.md, identify what exactly needs research
2. **Set boundaries** — what is in scope, what is out of scope (prevents open-ended wandering)
3. **Trace protocol** — for each research step:
   - Thought: what am I looking for next?
   - Action: file read / search / grep
   - Observation: what I found
   - Decision: continue searching / enough context
4. **Identify patterns** — how does the codebase handle similar things?
5. **Assess impact** — what could be affected by changes in this area?
6. **Produce findings** — structured output at `specs/features/NNN-name/research.md`

## Trace Protocol (required for bounded research)
```
Thought: What pattern is used for error handling?
Action: ripgrep "func.*error" internal/
Observation: All functions return custom Error type from internal/errors
Decision: Enough context. Document the pattern.
```

## Integration
- Agent: Scout
- Feeds: `plan` skill (direct output)
- Triggered by: spec.md existing with Clarifications section complete
- Output: `specs/features/NNN-name/research.md`