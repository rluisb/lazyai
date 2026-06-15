---
name: adversarial-review
description: Use when a claim, plan, spec, doc, or design must be attacked against source evidence before implementation or approval.
status: draft
---

# Adversarial Review Workflow

## Trigger

Start when a proposal, claim, workflow, spec, architecture note, or support statement needs skeptical review before implementation or adoption.

## Inputs

- `artifact_type` — plan, spec, doc, code diff, design, workflow, or claim set
- `claim_or_goal` — what is being tested
- `sources` — evidence allowed for review
- `strictness` — `normal | adversarial`
- `validate` — required review output shape

## Purpose

Stress the artifact against repo principles, source evidence, and explicit constraints before it hardens into code or policy.

## Four Points

- WHAT: Produce a grounded review verdict.
- HOW: Steelman the proposal, attack it from principles/evidence, isolate actionable disagreements.
- DON'T WANT: Vibes-only approval, fake certainty, or hidden assumptions.
- VALIDATE: Findings are classified as supported, contradicted, or inconclusive with exact evidence.

## Behavior Scenario

n/a — read-only review workflow.

## TDD Mode

n/a — no behavior change is implemented in this workflow.

## Steps

1. State the artifact under review and the baseline principles.
2. Read only the sources needed to test the claims.
3. Record advocate and skeptic arguments.
4. Classify each important claim as supported, contradicted, or inconclusive.
5. Return a verdict, key risks, and the smallest safe recommendation.

## Adapters

- Claude Code: markdown workflow only; reviewer and evidence-verifier style agents/skills perform the work.
- OpenCode: markdown workflow only; no workflow runtime claims.
- OMP/Pi: markdown-only guidance.

## Exit Gate

Stop only when the artifact has a concrete verdict, evidence-backed findings, and unresolved disagreements are local and explicit.

## Failure

Stop and report when the review lacks enough source material, the baseline principles are unclear, or the requested verdict would require speculation.
