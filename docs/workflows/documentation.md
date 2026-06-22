---
name: documentation
description: Use when the deliverable is docs or reference material and every claim must stay tied to observed behavior or cited source evidence.
status: draft
---

# Documentation Workflow

## Trigger

Start when the requested output is a guide, setup document, reference, decision note, or other docs-first artifact.

## Inputs

- `goal` — what the document should help the reader do or understand
- `audience` — who the doc is for
- `sources` — code/docs/commands that back the claims
- `doc_type` — `reference | how-to | decision | setup | roadmap`
- `validate` — doc build, runnable commands, or manual proof path

## Purpose

Create or update documentation that is accurate, scoped, and grounded in observed behavior.

## Four Points

- WHAT: Deliver the requested documentation outcome.
- HOW: Gather sources, write for the stated audience, verify command claims, keep docs aligned to real behavior.
- DON'T WANT: Marketing language, speculative support claims, or commands that were not verified.
- VALIDATE: Docs build and every operational claim is backed by source or observed command output.

## Behavior Scenario

n/a — documentation workflow unless the document describes a new behavior that belongs in a separate implementation workflow.

## TDD Mode

n/a — no behavior change unless the docs generation path itself is executable behavior.

## Steps

1. Confirm audience, doc type, and validation path.
2. Read the minimal set of source material that grounds the doc.
3. Write or update the doc with explicit boundaries and cited facts.
4. Verify any commands or examples that the doc claims should work.
5. Run the docs build or the smallest applicable documentation gate.

## Adapters

- Claude Code: markdown workflow only; docs remain repo artifacts, not runtime behavior.
- OpenCode: markdown workflow only.
- Pi: markdown-only guidance.

## Exit Gate

Stop only when the doc matches observed behavior, unsupported claims are removed or marked, and the docs validation path passes.

## Failure

Stop and report when the needed evidence cannot be gathered, the requested claim cannot be verified, or the change actually requires implementation first.
