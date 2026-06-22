# Context Discipline

Context discipline is the practice of managing what an AI agent reads, retains,
and passes forward. It is the intake, retention, and handoff hygiene that
prevents context bloat, reduces token costs, and keeps agents focused on their
assigned task.

## The Three Practices

Context discipline has three complementary practices:

| Practice | Question | Asset |
|---|---|---|
| **Reading discipline** | What should I read? | [context-discipline fragment](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/fragments/context-discipline.md) |
| **Compaction** | What should I keep? | [context-compaction fragment](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/fragments/context-compaction.md) |
| **Handoff** | What should I pass forward? | [context-handoff template](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/templates/context-handoff.md) |

### 1. Reading discipline

Reading discipline governs what an agent reads before acting. It is the intake
side of context management: read the minimum to answer the question, stop when
you have enough, and never read "just in case."

The [context-discipline fragment](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/fragments/context-discipline.md)
defines the rules: a context budget (max 10 files before a change), a priority
order (task first, then files being modified, then types, then standards), and
anti-patterns (no speculative reads, no re-reading unchanged files).

### 2. Compaction

Compaction governs what an agent retains after a phase boundary, a handoff, or
a context-pressure event. It is the retention side of context management:
preserve decisions, constraints, and evidence; drop execution trace, verbose
output, and resolved questions.

The [context-compaction fragment](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/fragments/context-compaction.md)
defines the compact-after triggers (phase boundary, handoff, drift, window
full, retry, human gate) and the never-compact-away items (task, constraints,
decisions, open questions, evidence, file paths, next action, pending
handoffs).

The [session-compaction template](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/templates/session-compaction.md)
provides a structured format for mid-session compaction. The
[recovery-handoff template](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/templates/recovery-handoff.md)
provides a structured format for compaction after a failure.

### 3. Handoff

Handoff governs what an agent passes to another agent or a future session. It
is the transfer side of context management: package the essential state,
decisions, and evidence into a structured document that the next agent can
consume without re-deriving context.

The [context-handoff template](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/templates/context-handoff.md)
provides a structured format with evidence requirements, trace evidence
entries, and a required-evidence checklist that the receiving agent must
verify before proceeding.

## Why context discipline matters

Without context discipline, agents exhibit predictable failure modes:

- **Context bloat** — the agent reads more than it needs, filling the context
  window with irrelevant information and degrading performance.
- **Decision loss** — the agent makes a decision, then forgets it, and
  re-derives it (or worse, makes a contradictory decision).
- **Constraint drift** — the agent satisfies a constraint early, compacts it
  away, and violates it in a later phase.
- **Evidence erosion** — the agent claims "tests pass" without preserving the
  test output, leaving the next agent unable to trust the claim.
- **Handoff gaps** — the agent ends a session without recording what was done,
  what was decided, or what comes next.

Context discipline prevents all five failure modes by making the intake,
retention, and transfer of context explicit and structured.

## Relationship to other concepts

| Concept | Relationship |
|---|---|
| [Token rent](token-rent.md) | Token rent is the hard byte budget for the canonical library; context discipline is the behavioral practice that keeps agent sessions within their effective context window. |
| [Minimality](minimality.md) | Minimality is the principle that every byte in the canonical library earns its place; context discipline is the same principle applied to session context. |
| [Agent contracts](agent-contracts.md) | Agent contracts define what an agent does; context discipline defines how an agent manages its working memory while doing it. |
| [Harness principles](harness-principles.md) | The "Handoff and memory hygiene" principle is the harness-level expression of context discipline. |
| [Trace/eval improvement loop](trace-eval-improvement-loop.md) | Trace evidence feeds the improvement loop; context discipline ensures trace evidence survives compaction and handoff. |

## See also

- [context-discipline fragment](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/fragments/context-discipline.md) — reading discipline rules
- [context-compaction fragment](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/fragments/context-compaction.md) — compaction rules and triggers
- [context-handoff template](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/templates/context-handoff.md) — structured handoff format
- [session-compaction template](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/templates/session-compaction.md) — mid-session compaction format
- [recovery-handoff template](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/templates/recovery-handoff.md) — recovery handoff format
