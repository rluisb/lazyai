<rule>
  <scope>auto</scope>
  <globs>skills/**,commands/**,prompts/**</globs>
  <description>Skill workflow guidance for reusable execution chains across skills, commands, and prompts directories</description>
</rule>

# Skills and Commands Directory

This directory contains workflow skills that guide execution quality. Skills are not isolated commands; they are staged methods used during a task lifecycle.

## Skills Inventory

| Skill | Purpose | Triggers | Chain Position |
|------|---------|----------|----------------|
| research | Build factual context and identify constraints | Task starts, requirements unclear, dependencies unknown | 1 |
| plan | Produce an ordered, testable execution approach | Research completed, implementation requested | 2 |
| implement | Apply the approved change set with minimal scope | Plan approved, execution authorized | 3 |
| iterate | Refine based on review feedback and verification results | Failing checks, review comments, partial acceptance | 4 |
| anti-speculation | Prevent assumptions and require evidence for claims | Ambiguity appears, undocumented behavior inferred | Cross-cutting guardrail |
| tdd-loop | Drive changes through fail → pass → refactor cycles | Behavioral changes, bug fixes, regression prevention | During implement + iterate |
| memory-write | Capture decisions, rationale, and reusable context | Implementation accepted, notable tradeoff made | 5 |
| lessons-learned | Distill improvements for future execution | Task wrap-up, recurring friction detected | 6 |
| parallel-execution | Coordinate independent tracks safely and efficiently | Multiple non-blocking tasks identified | Optional accelerator |

## Mandatory Skill Chain

`research → plan → implement → iterate → memory-write → lessons-learned → task closed`

Use this chain as the default operating sequence. Deviations are allowed only when a step is clearly not applicable and the reason is explicit.

## Invocation Rules

- Skills are invoked as workflows inside task execution, not as standalone goals.
- Enter each step with clear inputs and leave with concrete outputs.
- Do not start implement before plan quality and acceptance criteria are explicit.
- Use anti-speculation whenever confidence is based on inference rather than evidence.
- Use tdd-loop whenever behavior can be validated via repeatable tests.

## Usage Guidelines

1. Keep one primary skill active at a time.
2. Complete or checkpoint the current skill before switching.
3. Respect chain order unless a justified exception is documented.
4. Prefer parallel-execution only for independent tasks with isolated risk.
5. Feed iterate outputs back into verification until acceptance criteria are met.

## Coordination Expectations

- Every transition should include what was learned, what remains, and what is next.
- If blocked, return to the previous chain step rather than guessing forward.
- Close tasks only after memory-write and lessons-learned are complete.

## Self-Improvement

When this directory changes:
- Update the skills inventory table for every new or retired skill.
- Keep chain position metadata accurate and consistent with workflow reality.
- Refresh the mandatory chain whenever lifecycle stages evolve.
- Expand invocation examples when teams adopt new execution patterns.
