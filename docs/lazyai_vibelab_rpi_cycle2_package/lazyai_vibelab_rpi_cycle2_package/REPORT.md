# LazyAI / vibe-lab RPI Cycle 2 Report — Semantic Skill and Agent Validation

## 1. Research summary

- **Repo state:** RPI Cycle 1 completed successfully. Codebase resides in `packages/cli` for Go source and `packages/cli/library` for assets.
- **Current validation paths:** The `validate.All()` engine handles structure and frontmatter validation, running 7 distinct validation sweeps over `.ai/`. It populates a `Report` with `SeverityError` and `SeverityWarning`.
- **Current skill format:** YAML frontmatter + Markdown body. No existing formal schema validation for semantic headers.
- **Current agent format:** YAML frontmatter + Markdown body. 8 canonical agents form a specialist architecture. No formal semantic validation for handoff or gating instructions.
- **Existing tests:** Table-driven tests utilizing `t.Errorf`/`t.Fatalf` on issue presence (e.g., `len(issuesFor(report, rule))`). Inline fixtures; no golden tests.
- **Constraints confirmed:** No runtime, orchestration, or scoring agents have been added. Implementation stayed purely within the compiler and structural validation layers.

## 2. Plan executed

| Item | Status | Notes |
|---|---|---|
| Complete Research via Parallel Agents | Done | 4 Subagents thoroughly mapped the system. |
| Produce execution plan | Done | Approved by user before continuing. |
| Implement Semantic Validation for Skills | Done | Warning rules added for missing semantics. |
| Implement Semantic Validation for Agents | Done | Warning rules added for missing agent contracts. |
| Add corresponding Tests | Done | Inline `TestValidateSkillsSemanticChecks` added. |
| Add Docs and Templates | Done | Quality docs and templates generated. |
| Pass all validations and generate report | Done | `go test ./...` passed across all packages. |

## 3. Changes made

| Area | Files changed | Summary |
|---|---|---|
| **Validation Logic** | `packages/cli/internal/validate/validate.go` | Added `semanticHas` helper. Implemented semantic warning/error sweeps during `validateSkills` and `validateAgents`. Fixed missing generic rule IDs to specific ones. |
| **Validation Tests** | `packages/cli/internal/validate/validate_test.go` | Added `TestValidateSkillsSemanticChecks` and `TestValidateAgentsSemanticChecks` alongside utility closure `hasIssueForFile`. |
| **Docs** | `docs/concepts/skill-quality.md`, `docs/concepts/agent-contracts.md` | Defined requirements for high-quality skills and agent contracts for authors. |
| **Templates** | `packages/cli/library/templates/skill-quality.md`, `packages/cli/library/templates/agent-contract.md` | Added scaffold templates emphasizing Required Evidence, Workflows, Output and Handoffs. |

## 4. Skill validation

- **New checks:** Detects presence of trigger guidance, misuse avoidance, evidence requirements, and expected outputs.
- **Warning rules:** `skill.missing_trigger`, `skill.missing_non_trigger`, `skill.missing_evidence`, `skill.missing_output`.
- **Error rules:** `skill.invalid_frontmatter`, `skill.missing_required_name`, `skill.empty_body`.
- **Tests:** Covered completely within `TestValidateSkillsSemanticChecks`.
- **Remaining gaps:** Currently uses naive substring checking which might cause false positives for unconventional terminology (mitigated by wide sets of variant keywords).

## 5. Agent validation

- **New checks:** Detects presence of agent roles, triggers, non-triggers, workflows, human gating, and handoffs.
- **Warning rules:** `agent.missing_role`, `agent.missing_workflow`, `agent.missing_trigger`, `agent.missing_non_trigger`, `agent.missing_human_gate`, `agent.missing_handoff`, `agent.missing_evidence`, `agent.missing_output`.
- **Error rules:** `agent.invalid_frontmatter`, `agent.missing_required_name`, `agent.empty_body`.
- **Tests:** Covered completely within `TestValidateAgentsSemanticChecks`.
- **Remaining gaps:** Does not enforce sequence matching between referenced skills and the text described in workflows.

## 6. Validation output quality

- **Rule IDs added:** Yes, updated the generic `"skill"` and `"agent"` error reports to granular rule IDs matching documentation (e.g. `skill.missing_output`).
- **Example output:** 
  `SeverityWarning: [skill.missing_trigger] missing 'when to use' / trigger guidance`
- **Fix suggestions:** Added descriptions indicating exactly what heading is missing.

## 7. Docs/templates

- **Added/updated:** Added `skill-quality.md` and `agent-contracts.md` in both `docs/concepts/` and `packages/cli/library/templates/`.
- **Linked from:** N/A natively, will be included into the library manifest if requested.
- **Remaining gaps:** Integrating templates into standard `lazyai-cli init` generation.

## 8. Tests and validation

Commands run:

```bash
go test -v ./packages/cli/internal/validate/...
go test ./packages/cli/...
```

Results:

```text
go test: 42 packages ok, 11 no tests
Wall time: 84.48 seconds
```

## 9. Changed files

```text
M packages/cli/internal/validate/validate.go
M packages/cli/internal/validate/validate_test.go
?? docs/concepts/agent-contracts.md
?? docs/concepts/skill-quality.md
?? packages/cli/library/templates/agent-contract.md
?? packages/cli/library/templates/skill-quality.md
```

## 10. Risks

| Risk | Impact | Mitigation |
|---|---|---|
| False positives in semantic string matching | Low | Evaluated at `SeverityWarning` level so CI does not crash. We included many synonyms for heading matching. |
| Existing library assets throwing warnings | Low/Medium | Since it relies on warnings, users are nudged toward correct behaviors without halting workflows. |

## 11. Remaining work

| Item | Why not completed | Suggested next step |
|---|---|---|
| Updating existing standard library skills | Out of Scope | Mass rewrite of the 50+ existing skills was specifically flagged as a non-goal for this cycle. Run validations against them and fix iteratively later. |
| Exposing new templates in init flow | Wait for Review | Need maintainer approval before integrating into `lazyai-cli init`. |

## 12. Product boundary confirmation

Confirm explicitly:

- No runtime/orchestration surface was added.
- No old task/workflow/eval command was reintroduced.
- No mandatory judge/scoring engine was added.
- No mandatory trace daemon was added.
- No mandatory RAG core was added.
- LazyAI remains a harness asset manager/compiler.
- vibe-lab remains the process/quality asset layer.