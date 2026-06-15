# Caveman + ai-memory Balance

Caveman reduces working-context verbosity. ai-memory preserves reusable shared memory. They are not substitutes.

## Use Caveman For

- Single-session planning, comparison, or handoff compression.
- Reducing verbose specs into Goal / Must / Must Not / Can / Cannot.
- Token-saving summaries that link back to the full source.

## Use ai-memory For

- Decisions, traps, conventions, and root causes likely to recur.
- Knowledge that must survive sessions, agents, or people.
- Facts with enough context to be reused without re-reading the whole thread.

## Sweet Spot

- Compress with caveman while thinking.
- Promote only the stable reusable insight to memory.
- Memory entries need context: source, evidence, decision, scope, and expiry/removal condition when relevant.
- Never store a bare caveman bullet as durable memory.

## Hook / Plugin Opportunity

A `SessionEnd` or `PreCompact` hook/plugin may scan for caveman summaries and ask whether reusable facts should enter `memory-promotion`.

It must not write memory automatically. It should emit a proposal using `canonical/learning-template.md`.
