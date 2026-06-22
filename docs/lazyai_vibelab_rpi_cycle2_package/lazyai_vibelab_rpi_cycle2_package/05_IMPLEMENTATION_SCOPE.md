# RPI Cycle 2 — Implementation Scope

## Priority A — Skill semantic validation

Implement conservative semantic validation for skills.

A high-quality skill should clearly state:

```text
- when to use it
- when not to use it
- required evidence
- expected output or done criteria
- required tools or dependencies, if any
- human gate requirements, if any
- examples or anti-examples when useful
- progressive-disclosure behavior where relevant
```

### Error-level skill checks

Use errors only for objective issues:

```text
- malformed frontmatter if frontmatter exists
- invalid schema if schema exists
- broken referenced skill/fragment/tool path
- unsupported adapter target
- empty skill body
- missing title/name if current format requires it
```

### Warning-level skill checks

Use warnings for quality issues:

```text
- missing “when to use” / trigger guidance
- missing “when not to use” / misuse guidance
- missing evidence requirement
- missing expected output / done criteria
- missing human gate guidance for risky skills
- missing examples/anti-examples for broad skills
- too-broad name or description
- duplicated always-on rule content
- excessive always-loaded content / token-rent concern
- no progressive-disclosure cue where the skill is large or procedural
```

### Acceptable heading variants

```text
when to use:
- When to use
- Trigger
- Triggers
- Use when
- Invocation
- Activation

when not to use:
- When not to use
- Do not use
- Non-goals
- Misuse
- Avoid

evidence:
- Evidence
- Required evidence
- Verification
- Validation
- Proof
- Done criteria

output:
- Output
- Expected output
- Deliverable
- Result
- Response format
- Done
```

### Potential files

```text
packages/cli/internal/compiler/skill_validate.go
packages/cli/internal/compiler/skill_validate_test.go
packages/cli/internal/validate/
packages/cli/internal/schema/
packages/cli/library/templates/skill-quality.md
docs/concepts/skill-quality.md
```

### Skill validation acceptance criteria

```text
- Validation detects clearly bad skills.
- Existing shipped skills either pass or produce intentional, actionable warnings.
- Tests cover good skill, missing trigger, missing non-trigger, missing evidence, empty body, and broken reference if references exist.
- No unnecessary mass rewrite of all skills.
```

---

## Priority B — Agent role contract validation

A high-quality agent should define:

```text
- role / purpose
- when to use
- when not to use
- expected workflow
- referenced skills/tools/fragments
- evidence requirements
- human gates
- output format
- handoff behavior
- safety boundaries
```

Default agents to validate:

```text
guide
implementer
researcher
planner
reviewer
deployer
responder
evidence-verifier
```

### Error-level agent checks

Use errors for objective issues:

```text
- missing/empty agent file
- malformed metadata/frontmatter if required
- invalid adapter target if declared
- broken referenced skill/fragment/tool
- missing required structural fields if the existing schema already requires them
```

### Warning-level agent checks

Use warnings for semantic quality:

```text
- missing role/purpose
- missing when-to-use guidance
- missing when-not-to-use guidance
- missing workflow
- missing evidence requirements
- missing human gate guidance
- missing output format
- missing handoff behavior
- ambiguous overlap with another agent
```

### Potential files

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/compiler/agent_validate_test.go
packages/cli/library/templates/agent-contract.md
docs/concepts/agent-contracts.md
```

### Agent validation acceptance criteria

```text
- Default agents have explicit contracts or actionable warnings.
- Tests cover valid agent, missing role, missing evidence, missing handoff, and broken reference if references exist.
- No runtime behavior added.
```

---

## Priority C — Validation result quality

Each semantic validation message should include:

```text
- asset type: skill or agent
- asset path
- severity: error/warning/info
- rule id
- short message
- fix suggestion
```

Example:

```text
warning skill.missing_non_trigger packages/cli/library/skills/review.md
Skill does not explain when not to use it.
Add a “When not to use” or “Misuse” section to prevent accidental invocation.
```

Acceptance criteria:

```text
- Validation output is actionable.
- Tests assert rule IDs.
- Existing CLI output style is preserved where possible.
```

---

## Priority D — Docs/templates for future authors

Suggested files:

```text
docs/concepts/skill-quality.md
docs/concepts/agent-contracts.md
packages/cli/library/templates/skill-quality.md
packages/cli/library/templates/agent-contract.md
```

Content should explain:

```text
- skill vs agent distinction
- progressive disclosure
- trigger and non-trigger guidance
- evidence requirements
- human gates
- output contracts
- examples and anti-examples
- how validation works
```

Acceptance criteria:

```text
- Docs are linked from harness principles if appropriate.
- Templates are included in curation/manifest if required by repo conventions.
- No duplicate or conflicting docs are created if equivalent docs already exist.
```
