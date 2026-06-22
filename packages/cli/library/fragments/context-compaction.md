<context-compaction>

# Context Compaction

Compaction is the practice of reducing session context to its essential signal
before a handoff, a new phase, or a new session. It is not summarization for
its own sake — it is a deliberate filter that preserves decision-critical
information and discards noise.

## When to compact (compact-after)

Compact after any of these events:

| Trigger | Rationale |
|---|---|
| A phase boundary (research → plan → implement → verify → cleanup) | Each phase has different context needs; stale investigation details from research are noise during implementation. |
| Before a handoff to another agent or session | The receiving agent needs a focused summary, not a full transcript. |
| After 15+ tool calls without a decision | The session is drifting; compact to re-anchor on the goal. |
| When context exceeds 70% of the available window | Performance degrades; compact to free working memory. |
| After a failed attempt and before retrying | Clear the failure trace from working memory; keep only the root-cause insight. |
| Before a human gate review | The human reviewer needs a concise status, not the full execution trace. |

## What to never compact away (never-compact-away)

These items MUST survive every compaction:

| Item | Why |
|---|---|
| The current task description and acceptance criteria | Without them, the agent loses its goal. |
| Active constraints and non-goals | Forgetting a constraint causes rework or out-of-scope work. |
| Decisions made with rationale (including rejected alternatives) | Re-deriving a decision wastes time; re-picking a rejected option wastes more. |
| Open questions and unresolved unknowns | These are the next investigation targets. |
| Evidence of completed verification (test results, command output, lint passes) | Without evidence, the next agent cannot trust the claimed state. |
| File paths of modified or created files | The next agent needs to know where to continue. |
| The current phase and next immediate action | Without this, the agent must re-derive its position. |
| Handoff or recovery context if one is pending | A pending handoff is a commitment; compacting it away breaks the chain. |

## How to compact

1. **Identify the trigger** — which compact-after event fired.
2. **Preserve the never-compact-away items** — copy them verbatim or reference
   their source (file path, artifact id).
3. **Drop execution trace** — tool call sequences, intermediate outputs,
   exploration dead ends, retry attempts (keep only the final successful
   result).
4. **Drop verbose tool output** — full file reads, long search results,
   command output dumps. Keep only the summary or the relevant excerpt.
5. **Drop resolved questions** — once a question is answered and the answer
   is recorded, the investigation trail is noise.
6. **Format the result** — use the [session-compaction template] or
   [context-handoff template] as appropriate.

## Anti-patterns

- ❌ Compacting away the task description — the agent loses its goal.
- ❌ Compacting away a constraint because it was already satisfied — the
  constraint may apply to the next phase too.
- ❌ Keeping full tool output "just in case" — that is the opposite of
  compaction.
- ❌ Compacting before a decision is made — compaction is for after a
  decision boundary, not before one.
- ❌ Replacing evidence with a claim — "tests passed" without the test
  output or command exit code is not evidence.

## Relationship to other assets

| Asset | Relationship |
|---|---|
| [context-discipline fragment](context-discipline.md) | Reading discipline (what to read) is the intake side; compaction (what to keep) is the retention side. |
| [context-handoff template](../templates/context-handoff.md) | Use when the compacted output is for another agent or session. |
| [session-compaction template](../templates/session-compaction.md) | Use when the compacted output is for the same agent mid-session. |
| [recovery-handoff template](../templates/recovery-handoff.md) | Use when compaction follows a failure and the next step is recovery. |

</context-compaction>
