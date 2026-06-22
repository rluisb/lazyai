# RPI Cycle 2 Plan

## 1. Current state

**Validation Paths:**
- Agents are validated by two engines: `validate.All()` structurally audits `.ai/agents/` (frontmatter, fields). Compile-time `ValidateAgentResolutions` checks model resolution for tools.
- Skills are structurally audited by `validate.All()` (generic Markdown/YAML parsing) and compile-time contract validation checks skill chain completeness (`ValidateChain`).
- Both errors and warnings exist in the `Report` struct.

**Skill Format:**
- 50+ skills using YAML frontmatter + Markdown body.
- No formal JSON schema for skills.
- Currently lacks structured validation for subjective quality (e.g., triggers, non-trigger guidance, outputs, handoff behavior).

**Agent Format:**
- 8 canonical agents form a specialist architecture routed by `guide`.
- Each has YAML frontmatter (name, role, steps, skills, etc.) and a system prompt.
- Missing explicit schema definitions for subjective components like handoff protocols, gate markers, and output formats.

**Test Style:**
- Imperative `t.Run()` table-driven tests.
- Inline fixtures using string literals, `fstest.MapFS`, and `t.TempDir()`.
- No golden files.
- Assertions use `t.Errorf`/`t.Fatalf` on issue presence (e.g., `hasError(report, "skill")`).

## 2. Implementation scope

- Add **Semantic validation** for skills and agent contracts in the `validate` or `compiler` packages.
- Rules will focus on ensuring the body of the markdown contains critical quality markers: triggers, evidence requirements, and human gate guidance.
- We will rely on warning-level severity for subjective checks, while maintaining hard errors for structural faults.

## 3. Warning rules

Skill rules:
- `skill.missing_trigger`: missing “when to use” / trigger guidance
- `skill.missing_non_trigger`: missing “when not to use” / misuse guidance
- `skill.missing_evidence`: missing evidence requirement
- `skill.missing_output`: missing expected output / done criteria

Agent rules:
- `agent.missing_role`: missing role/purpose
- `agent.missing_trigger`: missing when-to-use guidance
- `agent.missing_non_trigger`: missing when-not-to-use guidance
- `agent.missing_workflow`: missing workflow
- `agent.missing_evidence`: missing evidence requirements
- `agent.missing_human_gate`: missing human gate guidance
- `agent.missing_handoff`: missing handoff behavior
- `agent.missing_output`: missing output format

## 4. Error rules

We will stick to the existing objective errors (invalid schema, malformed files, unsupported targets) and only add warnings for subjective qualities. `skill.empty_body` and `agent.empty_body` will be checked if not already present.

## 5. Files to change

- `packages/cli/internal/validate/validate.go`: Add new semantic validation sweeps for agents and skills.
- `packages/cli/internal/validate/validate_test.go`: Add semantic validation test cases.

## 6. Tests to add/update

- Test cases in `validate_test.go` using inline string fixtures representing valid/invalid semantic skill/agent bodies.
- Assert `hasError(report, "skill-missing-trigger")` evaluates correctly.

## 7. Docs/templates

- `docs/lazyai_vibelab_rpi_cycle2_package/lazyai_vibelab_rpi_cycle2_package/07_DOCS_AND_TEMPLATES.md` - use this to update templates in the library so new authors follow the patterns.
- Ensure the existing `canonical/agents` or `library/skills/` templates include standard headers for "When to use", "Evidence required", "Output".

## 8. Risks

- Adding semantic text-matching rules can introduce false positives if users use non-standard phrasing.
- To mitigate, the rules will be warnings (not errors) and look for broad keywords/headers.

## 9. Validation commands

- `go test ./packages/cli/internal/validate/...`
- `go test ./packages/cli/internal/compiler/...`
- `go test ./packages/cli/cmd/...`