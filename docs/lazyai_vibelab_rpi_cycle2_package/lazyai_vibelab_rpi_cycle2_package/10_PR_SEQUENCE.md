# Suggested PR Sequence — RPI Cycle 2

Use this sequence if Cycle 2 is too large for one patch.

## PR 1 — Skill semantic validation foundation

Title:

```text
Add semantic validation warnings for skills
```

Scope:

```text
packages/cli/internal/compiler/skill_validate.go
packages/cli/internal/compiler/skill_validate_test.go
```

Checks:

```text
skill.empty_body
skill.missing_trigger
skill.missing_non_trigger
skill.missing_evidence
skill.missing_output
skill.missing_progressive_disclosure
```

Acceptance:

```text
- Warnings are actionable.
- Tests cover valid and invalid skills.
- Existing skills are not broken by subjective checks.
```

---

## PR 2 — Agent role contract validation

Title:

```text
Add semantic validation warnings for canonical agent contracts
```

Scope:

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/compiler/agent_validate_test.go
```

Checks:

```text
agent.missing_role
agent.missing_trigger
agent.missing_non_trigger
agent.missing_workflow
agent.missing_evidence
agent.missing_handoff
agent.missing_output
```

Acceptance:

```text
- Default agents have explicit contracts or actionable warnings.
- Tests cover valid and invalid agents.
```

---

## PR 3 — Validation message quality

Title:

```text
Add rule IDs and fix suggestions to skill and agent validation
```

Scope:

```text
validation result formatting
compiler validation tests
cmd validate output tests if available
```

Acceptance:

```text
- Messages include severity, rule ID, asset path, message, suggestion.
- Tests assert rule IDs.
```

---

## PR 4 — Skill and agent author docs/templates

Title:

```text
Document skill quality and agent contract standards
```

Scope:

```text
docs/concepts/skill-quality.md
docs/concepts/agent-contracts.md
packages/cli/library/templates/skill-quality.md
packages/cli/library/templates/agent-contract.md
```

Acceptance:

```text
- Docs explain trigger/non-trigger guidance, evidence, output contracts, human gates, handoff, and progressive disclosure.
- Docs preserve LazyAI’s compiler boundary.
```

---

## PR 5 — Minimal asset updates

Title:

```text
Align shipped skills and agents with semantic validation guidance
```

Scope:

```text
Only update representative shipped assets if required by validation.
Avoid broad rewrites.
```

Acceptance:

```text
- Shipped assets produce no unexpected errors.
- Any warnings are intentional and documented.
```

---

## Recommended order

```text
1. PR 1
2. PR 2
3. PR 3
4. PR 4
5. PR 5 only if necessary
```

## Guardrail

Do not mix runtime/orchestration changes into any Cycle 2 PR.
