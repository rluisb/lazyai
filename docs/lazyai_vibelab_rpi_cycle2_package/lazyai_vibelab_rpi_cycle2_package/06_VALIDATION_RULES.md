# RPI Cycle 2 — Validation Rules

## Severity model

Use errors only for objective breakage.

Use warnings for subjective or semantic quality problems.

Use info for suggestions that should not affect success/failure.

## Error-level rules

```text
skill.empty_body
skill.invalid_frontmatter
skill.invalid_schema
skill.broken_reference
skill.unsupported_adapter
skill.missing_required_name

agent.empty_body
agent.invalid_frontmatter
agent.invalid_schema
agent.broken_reference
agent.unsupported_adapter
agent.missing_required_name
```

## Warning-level rules

```text
skill.missing_trigger
skill.missing_non_trigger
skill.missing_evidence
skill.missing_output
skill.missing_progressive_disclosure
skill.missing_human_gate
skill.missing_examples
skill.too_broad
skill.token_rent
skill.duplicates_always_on_rule

agent.missing_role
agent.missing_trigger
agent.missing_non_trigger
agent.missing_workflow
agent.missing_evidence
agent.missing_human_gate
agent.missing_handoff
agent.missing_output
agent.ambiguous_overlap
```

## Recommended validation message shape

Each validation message should include:

```text
- severity
- rule id
- asset type
- asset path
- message
- suggestion
```

Example:

```text
warning skill.missing_non_trigger packages/cli/library/skills/review.md
Skill does not explain when not to use it.
Add a “When not to use” or “Misuse” section to prevent accidental invocation.
```

## Skill quality checks

A skill should have:

```text
- clear trigger guidance
- clear non-trigger or misuse guidance
- evidence requirements
- expected output or done criteria
- progressive-disclosure behavior where relevant
- human gate guidance for risky actions
- examples or anti-examples when broad or ambiguous
```

## Agent quality checks

An agent should have:

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

## Heading variants to recognize

Do not require exact headings. Recognize reasonable variants.

### Trigger / when-to-use

```text
When to use
Trigger
Triggers
Use when
Invocation
Activation
```

### Non-trigger / misuse

```text
When not to use
Do not use
Non-goals
Misuse
Avoid
```

### Evidence

```text
Evidence
Required evidence
Verification
Validation
Proof
Done criteria
```

### Output

```text
Output
Expected output
Deliverable
Result
Response format
Done
```

## Compatibility requirement

Do not force a breaking schema migration.

If current assets are Markdown-only, implement Markdown/heading linting.

If current assets have frontmatter, validate frontmatter conservatively.

If current assets have schema support, extend it backward-compatibly when possible.
