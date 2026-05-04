---
name: plan
description: Plan implementation approach before writing code.
argument-hint: "[feature-or-task]"
trigger: /plan
phase: plan
---

# Plan Skill

## Workflow
1. Read research — confirm understanding of scope and findings
2. Capture the Knowledge Surface — facts, constraints, assumptions, unknowns, and evidence sources before commitments
3. Capture the Environment Snapshot — repo/tooling signals needed for realistic execution planning
4. Define acceptance criteria — what does "done" look like?
5. Break into phases — group related changes, order by dependency
6. Define tasks — one task per file, each with "Done When" criteria
7. Identify risks — what could go wrong, how to mitigate
8. Produce plan — write plan.md and task files

## Environment Snapshot

Include this section for non-trivial plans when execution depends on repo, runtime, tooling, or delivery constraints. Keep entries short and mark assumptions explicitly as verified or unverified.

```markdown
## Environment Snapshot
- **Toolchain:** language/runtime versions, framework versions, and required CLIs.
- **Package manager:** detected manager, lockfile/install state, and dependency workflow.
- **Available tools:** test runners, linters, build commands, local services, and MCP/tools available to the agent.
- **CI/check latency:** expected runtime for targeted checks and full gates; note slow or flaky checks.
- **Platform constraints:** OS, shell, filesystem, architecture, browser/device, or container constraints.
- **Budget/token constraints:** session budget, context limits, cost ceilings, and when to compact or split work.
- **Network/secrets constraints:** offline mode, credentials availability, secret handling, and external access limits.
- **Verified assumptions:** `[verified]` assumptions with evidence source paths, commands, or human decisions.
- **Unverified assumptions:** `[unverified]` assumptions plus the validation needed before relying on them.
```

If an environment fact is missing and could change the plan, record it as unknown instead of guessing.

## Plan Quality Check

Run this inline check after a feature `plan.md` is produced and before the human approval gate. When invoked as the pre-task plan-quality step, read `spec.md`, `plan.md`, and optional `research.md`; do not require or inspect generated task files, and set `checkedAgainst.tasks` to `null`.

### Rules

- **R1 — Spec coverage in plan:** Check that every Wave 1 `spec.md` AC/FR has at least one matching `plan.md` phase, item-detail, or downstream-contract reference. R1 does not inspect tasks.md because tasks do not exist during pre-task validation.
- **R2 — Phase exit criteria:** Check that every planned phase has observable phase exit criteria.
- **R3 — Risk mitigation completeness:** Check that every risk row has both mitigation detail and an explicit owner.
- **R4 — Wave 1 rollback coverage:** Check that every Wave 1 item describes rollback coverage.

### Verdict Semantics

- Structural uncertainty, malformed Markdown, ambiguous structural parsing, or parser uncertainty must produce `warn`, not fail.
- Confident content omissions may produce `fail`; all verdicts (`pass`, `warn`, and `fail`) proceed to the human approval gate.
- A `fail` verdict is a gate recommendation for the human reviewer; there is no automatic plan loop.

### PlanQualityReport Output

Emit exactly one `PlanQualityReport` JSON object with this contract:

```json
{
  "schemaVersion": "plan-quality-report/v1",
  "verdict": "pass|warn|fail",
  "findings": [
    {
      "rule": "R1|R2|R3|R4",
      "severity": "info|warn|fail",
      "message": "string",
      "location": {
        "file": "repo-relative/path.md",
        "section": "section heading or null",
        "lineStart": 1,
        "lineEnd": 1
      }
    }
  ],
  "checkedAgainst": {
    "spec": "specs/features/NNN-name/spec.md",
    "plan": "specs/features/NNN-name/plan.md",
    "research": "specs/features/NNN-name/research.md or null",
    "tasks": null
  }
}
```

Location files must be repo-relative. Use 1-based `lineStart`/`lineEnd` when available; use `null` for section or line numbers only when the location cannot be determined confidently.

## Integration
- Agent: Planner
- Requires: research output (research.md)
- Output feeds into: `implement` skill
- Output location: specs/features/NNN-name/plan.md + tasks/
