<reasoning-protocol>

# Reasoning Protocol

Pick the smallest reasoning technique that fits the task. Show the reasoning before acting on non-trivial work; skip on trivial edits.

---

## Default — Chain-of-Thought (CoT)

Use for any task that modifies logic, architecture, or > 20 lines.

<cot>
1. **Affected** — files, functions, tests, callers.
2. **Plan** — minimum change, concrete steps.
3. **Risks** — what could break, edge cases.
4. **Verdict** — proceed / clarify / blocked.
</cot>

Then implement.

---

## Technique catalog — when to use which

| Technique | When to use | What it produces |
|---|---|---|
| **Chain-of-Thought (CoT)** | Default for non-trivial tasks. Linear, single-author reasoning. | A single ordered trace of thought before acting. |
| **ReAct (Reason + Act)** | When the task requires interleaved reasoning, tool calls, and observations (e.g., debugging, exploration, research). | Alternating `<thought>` / `<action>` / `<observation>` loops until the goal is reached. |
| **Tree-of-Thoughts (ToT)** | Architectural decisions, multi-path design problems, ADR work. Used by the Decision Protocol. | At least two viable alternatives, each evaluated against complexity, reversibility, performance, team familiarity; one selected with rationale; rejected risks recorded. |
| **Self-Consistency** | High-stakes analysis where one chain might be wrong. Cross-artifact consistency checks (spec vs plan vs tasks). | Multiple independent reasoning passes, then majority/consensus + dissent notes. |
| **Reflexion** | After a failed attempt, before retrying. Self-improvement loops at session end. | A written critique of the previous attempt naming the specific flaw, and a corrected approach for the next attempt. |
| **Few-Shot** | Pattern-driven tasks where in-context examples accelerate convergence (e.g., bug-fix patterns, refactor patterns). | A worked example or two pulled from prior sessions or `specs/standards/`. |
| **LLM-as-Judge** | Verification, code review, multi-lens synthesis. Used by the `review` skill. | A judge prompt that evaluates one or more candidate artifacts against a rubric, citing the rubric clauses. |
| **Generated Knowledge** | Pre-research before a spike or hard plan. Pre-generate domain context so later steps reason over a known surface. | A short, sourced "background" document used as input to subsequent steps. |
| **Prompt Chaining** | Workflow execution (specify → clarify → plan → tasks → analyze → implement). Each step's output is the next step's input. | A pipeline where each prompt has a declared `consumes` and `produces_for` contract. |

---

## Selection rules

1. Default to CoT.
2. Add ReAct when tools are required (file reads, command runs, MCP queries).
3. Use ToT when the task could go more than one way and the wrong way is expensive.
4. Use Self-Consistency only when stakes justify the extra cost (consistency analysis, security review).
5. Apply Reflexion at session end as part of the `self-improve` skill.
6. LLM-as-Judge is invoked by the verifier role, never by the generator.

---

## Skip rules

Skip explicit reasoning for: renames, formatting, typo fixes, single-line changes, comment edits, dependency-version bumps with no API change.

</reasoning-protocol>
