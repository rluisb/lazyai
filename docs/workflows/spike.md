---
name: spike
description: Use when the goal is to investigate, map constraints, and return evidence without committing to implementation.
status: draft
---

# Spike Workflow

## Trigger

Start when requirements, constraints, or codebase shape are unclear enough that implementation would be guesswork.

## Inputs

- `topic` — question or domain to investigate
- `timebox` — explicit exploration limit
- `sources` — code/docs/systems allowed for review
- `validate` — expected research deliverable

## Purpose

Reduce uncertainty and return an evidence-backed recommendation or boundary without quietly slipping into implementation.

## Four Points

- WHAT: Produce grounded findings about the topic.
- HOW: Read targeted sources, map constraints, compare options, summarize evidence.
- DON'T WANT: Unapproved implementation, speculative architecture, or source-free opinions.
- VALIDATE: Findings cite the reviewed sources and end with a concrete recommendation or explicit open questions.

## Behavior Scenario

n/a — research-only workflow unless the spike is explicitly paired with a behavior scenario in a separate implementation workflow.

## TDD Mode

n/a — no behavior change is implemented in this workflow.

## Steps

1. Confirm the research question, timebox, and allowed sources.
2. Gather only the evidence required to answer the question.
3. Compare viable options or boundaries.
4. Return a concise recommendation, explicit non-recommendation, and cited evidence.
5. If implementation becomes necessary, stop and hand off to a feature/bugfix/refactor workflow.

## Adapters

- Claude Code: markdown workflow only; researcher/planner style work stays read-only.
- OpenCode: markdown workflow only; no orchestration claims.
- Pi: markdown-only guidance.

## Exit Gate

Stop only when the research question is answered with cited evidence or the remaining unknowns are stated explicitly.

## Failure

Stop and report when the allowed sources cannot answer the question, the timebox is exhausted without evidence, or the work would require implementation to continue.
