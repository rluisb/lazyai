# Issue 245 LazyAI / vibe-lab Parity Audit Plan

Date: 2026-06-15
Issue: https://github.com/rluisb/lazyai/issues/245
Research: `specs/issues/245-parity-audit/research.md`
Status: Draft — pending RPI Plan human approval

## Purpose

Produce a checked-in parity report for issue #245. The report will classify the current LazyAI asset set against the local vibe-lab baseline with exact source paths, exact LazyAI paths, adapter emission status, rationale, and follow-up recommendations.

This plan does not approve asset ports, manifest edits, adapter-output changes, or follow-up issue creation. Those actions are gated by the parity report findings and explicit human approval.

## Summary

Implementation target: `specs/issues/245-parity-audit/parity-report.md`.

The implementation phase should be documentation-only: refresh the inventory evidence, write one parity matrix per issue category, classify every observed vibe-lab baseline item, classify every LazyAI-only active/default item, and document stale provenance-path mismatches. No canonical assets, adapter mappings, manifests, CLI commands, MCP catalog entries, or migrations should change in this issue unless the human explicitly expands scope after reviewing the plan.

## Technical context

| Aspect                 | Decision                                                                                                                                                                                                                                                           | Rationale                                                                                                                   |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| Primary deliverable    | Markdown report under `specs/issues/245-parity-audit/`                                                                                                                                                                                                             | Satisfies issue acceptance without changing runtime behavior.                                                               |
| LazyAI source of truth | Current repo files under `packages/cli/library/`, `packages/cli/internal/adapter/output_mapping.go`, `docs/concepts/library-manifests.md`, `docs/concepts/product-boundaries.md`, `packages/cli/library/manifests/*.yaml`, `packages/cli/library/mcp/catalog.json` | These are the issue-named evidence sources.                                                                                 |
| vibe-lab baseline      | Local checkout at `/Users/ricardo/code/vibe-lab`                                                                                                                                                                                                                   | Issue #245 names this checkout as the baseline.                                                                             |
| Runtime behavior       | Unchanged                                                                                                                                                                                                                                                          | The audit should not port assets or reintroduce retired defaults.                                                           |
| Adapter output         | Unchanged                                                                                                                                                                                                                                                          | Report only; active output must remain neutral LazyAI.                                                                      |
| Manifest edits         | None planned                                                                                                                                                                                                                                                       | `provenance.yaml` and `curation.yaml` are updated only if approved asset changes happen; no asset changes are planned here. |
| Verification           | Markdown self-check plus conditional commands                                                                                                                                                                                                                      | Go tests/token-rent are only required if implementation expands to code, manifest, or canonical-library changes.            |

## Hard constraints

- LazyAI is the runtime/product; vibe-lab is a baseline/source of principles and adapter expectations only.
- Do not introduce a runtime dependency on `/Users/ricardo/code/vibe-lab`.
- Do not port, delete, or rename assets during this issue's planned implementation.
- Do not edit `packages/cli/library/manifests/provenance.yaml` or `packages/cli/library/manifests/curation.yaml` unless the human explicitly approves resulting asset changes.
- Do not change `packages/cli/internal/adapter/output_mapping.go`.
- Do not reintroduce Fortnite, orchestrator, eval, task, workflow, or orchestration defaults.
- Do not touch `packages/cli/internal/db/migrations.go`.
- Treat issue #245's initial snapshot as evidence, not authority, where the repo has moved on; document current state and stale snapshot differences.

## Constitution check

| Article                         | Verdict | Justification                                                                                                |
| ------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------ |
| I — Library-First               | PASS    | Reuses existing library/manifests/docs as evidence; no new dependency or runtime path.                       |
| II — Test-First                 | PASS    | Planned implementation is docs-only; if scope expands to code/assets, tests are mandatory before completion. |
| III — Docs as Source of Truth   | PASS    | Produces the requested checked-in report before any implementation-gap work.                                 |
| IV — Anti-Speculation           | PASS    | Classifies and recommends; does not port unapproved assets or create speculative defaults.                   |
| V — Simplicity Over Abstraction | PASS    | One report file plus existing research/plan artifacts.                                                       |
| VI — Anti-Overengineering       | PASS    | No schema, adapter, manifest, generator, or workflow machinery changes.                                      |

Verdict: plan is compliant for a documentation-only parity audit, pending human approval.

## Planned files

```text
specs/issues/245-parity-audit/
├── research.md        # existing evidence-only research artifact
├── plan.md            # this RPI Plan artifact
└── parity-report.md   # planned implementation deliverable
```

No `packages/cli/` file changes are planned.

## Classification taxonomy

### vibe-lab baseline item classifications

| Classification                                | Meaning                                                                                                              | Required evidence in report                                                                    |
| --------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `implemented active default`                  | LazyAI actively emits or ships the equivalent item by default.                                                       | Exact LazyAI path, adapter/CLI emission status, and equivalence rationale.                     |
| `implemented non-default/setup-library`       | LazyAI contains the item or equivalent as setup-library/docs/manual material, but not active default adapter output. | Exact LazyAI path and why non-default is correct.                                              |
| `renamed/equivalent`                          | LazyAI item fulfills the same role under a different name.                                                           | Exact vibe-lab path, exact LazyAI path, role-level equivalence rationale, and any naming risk. |
| `intentionally excluded`                      | Item is outside LazyAI product scope or conflicts with product boundaries.                                           | Boundary rationale and source evidence, usually `docs/concepts/product-boundaries.md`.         |
| `missing and should be ported/curated`        | No LazyAI equivalent was observed and the item appears valuable for LazyAI.                                          | Proposed LazyAI target area and follow-up issue candidate.                                     |
| `obsolete because LazyAI runtime replaced it` | LazyAI Go runtime/CLI supersedes the vibe-lab script/workflow/pattern.                                               | Exact replacement command/path or runtime component.                                           |

### LazyAI-only classifications

| Classification                      | Meaning                                                                      | Required evidence in report                           |
| ----------------------------------- | ---------------------------------------------------------------------------- | ----------------------------------------------------- |
| `LazyAI runtime-specific`           | Exists because LazyAI is a runtime/product, not just a setup baseline.       | Exact path and runtime/product role.                  |
| `setup-core extension`              | Extends setup creation/validation/compilation/update behavior.               | Exact path and setup-core role.                       |
| `compatibility/historical material` | Retained for compatibility or archived history, not active/default behavior. | Exact path, archive/curation state, and removal risk. |
| `candidate for removal/archive`     | Appears unnecessary, stale, or outside current boundaries.                   | Exact path, reason, and follow-up candidate.          |

## Report structure

`parity-report.md` must contain these sections:

1. **Scope and evidence lock**
   - Date, issue link, repo branch if observed during implementation, and local vibe-lab baseline path.
   - Statement that vibe-lab paths are local evidence, not runtime dependencies.
   - Statement that no assets/manifests/adapters were changed unless explicitly listed.

2. **Executive summary**
   - Count rows by category and classification.
   - List high-priority gaps and deliberate exclusions.
   - State whether active adapter output remains neutral LazyAI.

3. **Classification legend**
   - Include the taxonomy above so report rows are self-contained.

4. **MCP parity matrix**
   - Compare vibe-lab MCP categories/examples from `canonical/mcp-setup.md` to LazyAI `packages/cli/library/mcp/catalog.json`.
   - Include enabled/optional/disabled LazyAI catalog status and recommendation.

5. **CLI tools, scripts, and commands matrix**
   - Compare vibe-lab `bin/*` scripts to LazyAI Go CLI commands, repo `bin/*` harness scripts, and adapter command assets.
   - Keep repo scripts distinct from shipped product commands.

6. **Adapter command assets matrix**
   - Compare LazyAI canonical commands and Claude/OpenCode command assets against vibe-lab preset commands.
   - Include adapter target destinations from `output_mapping.go`.

7. **Agents matrix**
   - One row per vibe-lab `.agents/agents/*.md` item.
   - Include LazyAI canonical agent equivalents and active adapter status.
   - Include LazyAI-only active/default agents with rationale, especially `primary-agent`, `builder`, and `scout`.

8. **Skills matrix**
   - One row per vibe-lab `.agents/skills/*/SKILL.md` item.
   - Distinguish LazyAI canonical adapter skills from setup-library skills.
   - Include LazyAI-only active/default skills, especially `pr-review`.

9. **Hooks and policies matrix**
   - Compare vibe-lab `.agents/hooks/*/POLICY.md` items to LazyAI canonical hooks, setup-library hooks, and `rpi-gate-check.yml`.
   - Classify partial replacements explicitly; do not imply full equivalence without path evidence.

10. **Rules, standards, and protocols matrix**
    - Map vibe-lab `canonical/*.md` principles/templates/protocol docs to LazyAI rules, standards, templates, skills, or intentional exclusions.
    - Use concept-level mapping, not name-only matching.

11. **Workflows matrix**
    - Compare vibe-lab `.agents/workflows/*.md` and verified-research templates to LazyAI `packages/cli/library/specs-agents/*` and setup skills.
    - Preserve the boundary between setup documentation and retired runtime workflow command surfaces.

12. **Templates and presets matrix**
    - Compare vibe-lab speckit preset templates and `specs/_templates/*` to LazyAI `packages/cli/library/templates/*` and command assets.
    - Call out `constitution-template.md`, `prd.md`, and `techspec.md` by name.

13. **LazyAI-only active/default assets**
    - Classify every active/default LazyAI item not present in vibe-lab baseline.
    - Required minimum: canonical agents, canonical skills, canonical hooks, canonical commands, enabled MCP catalog entries, and active adapter output families.

14. **Stale provenance and current-baseline path mismatches**
    - Table for every observed `provenance.yaml` path that does not exist in the current vibe-lab checkout.
    - Distinguish historical source notes from current baseline paths.
    - Recommendation should be `document only` unless a later approved asset/manifest change requires a manifest update.

15. **Recommendations and follow-up candidates**
    - Group recommendations by priority: no-action, document-only, follow-up issue candidate, and blocked pending human decision.
    - Do not create GitHub issues automatically. List issue titles/bodies as candidates only unless the human explicitly approves creation.

16. **Acceptance checklist**
    - Map every issue #245 acceptance criterion to `Done`, `N/A`, or `Pending human approval` with evidence.

## Matrix row schema

Every vibe-lab baseline row must use this schema:

| Field                      | Required content                                                                                                               |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Category                   | MCP, CLI/script/command, adapter command, agent, skill, hook/policy, rule/standard/protocol, workflow, template/preset.        |
| vibe-lab item              | Stable item name.                                                                                                              |
| vibe-lab source path       | Exact local path or category source.                                                                                           |
| LazyAI classification      | One taxonomy value.                                                                                                            |
| LazyAI current/target path | Exact path(s), or `None observed`.                                                                                             |
| Adapter emission status    | `canonical default`, `adapter-specific command`, `setup-library only`, `CLI runtime`, `repo harness`, `not emitted`, or `N/A`. |
| Rationale                  | One concise, evidence-backed reason.                                                                                           |
| Recommendation             | `keep`, `document`, `follow-up candidate`, `exclude`, `archive candidate`, or `human decision needed`.                         |

Every LazyAI-only active/default row must use this schema:

| Field                   | Required content                          |
| ----------------------- | ----------------------------------------- |
| Category                | Asset category.                           |
| LazyAI item             | Stable item name.                         |
| LazyAI path             | Exact path.                               |
| Adapter emission status | Same vocabulary as above.                 |
| Classification          | One LazyAI-only taxonomy value.           |
| Rationale               | Why LazyAI should keep or revisit it.     |
| Recommendation          | Retain, document, or follow-up candidate. |

## Implementation steps after plan approval

1. **Refresh targeted inventories**
   - Re-check the LazyAI and vibe-lab path sets needed for the report.
   - Use the existing research artifact as the starting point, but update rows if the filesystem has changed.
   - Do not inspect unrelated files.

2. **Write `parity-report.md`**
   - Add the required report sections and table schemas.
   - Fill every category from issue #245.
   - Ensure every vibe-lab baseline item from research has one classification row.
   - Ensure every LazyAI-only active/default item has one rationale row.

3. **Document provenance mismatches**
   - Include the known mismatches from `provenance.yaml`.
   - Do not correct manifest paths in this issue unless the human explicitly expands scope.

4. **Document recommendations, not ports**
   - Mark candidate ports/curations as follow-up candidates.
   - Do not add assets, remove assets, or change adapter defaults.

5. **Verify the report against issue #245 acceptance**
   - Re-read the report sections that contain the acceptance checklist and matrix headers.
   - Confirm all issue categories are represented.
   - Confirm no planned-out-of-scope files were changed.

6. **Stop at Feedback gate**
   - Ask the human to approve classifications and any follow-up issue creation.
   - Create GitHub issues only for explicitly approved gaps.

## Conditional verification commands

For the planned docs-only implementation:

- No Go test is required because runtime code, assets, manifests, and adapter output do not change.
- Verify by re-reading `parity-report.md` and checking it against issue #245 acceptance.

If the human expands scope to include implementation changes:

- Run `go test ./packages/cli/...` after code, manifest, library, or adapter changes.
- Run `go run ./packages/cli/internal/tokenrent/cmd/token-rent-check` after canonical library changes.
- Run `go build ./packages/cli/...` if code changes affect CLI build behavior.
- Do not use root `go build ./...`; it is intentionally invalid for this repo.

## Risks and mitigations

| Risk                                                            | Mitigation                                                                                                     |
| --------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| Local vibe-lab checkout changes during audit                    | Record date and exact paths; refresh targeted inventory during implementation.                                 |
| Name-only matching hides semantic gaps                          | Require role-level rationale for `renamed/equivalent` rows.                                                    |
| Issue #245 snapshot includes stale LazyAI retained agents       | Report current state and note issue-snapshot drift.                                                            |
| Provenance paths do not match current vibe-lab checkout         | Document as provenance mismatch; do not silently rewrite history.                                              |
| Report recommendations are mistaken for approved implementation | Keep recommendations separate from asset changes and gate follow-up issues.                                    |
| Retired runtime defaults are accidentally revived               | No adapter/asset changes in this issue; report must explicitly state neutral adapter output remains unchanged. |

## Out of scope

- Porting vibe-lab agents, skills, hooks, workflows, templates, scripts, or MCP entries.
- Changing active adapter output or command destinations.
- Editing `packages/cli/library/manifests/provenance.yaml` or `curation.yaml`.
- Editing `packages/cli/library/mcp/catalog.json`.
- Editing Go CLI/runtime code.
- Creating GitHub follow-up issues without explicit human approval of specific gaps.
- Modifying archived historical assets.
- Modifying `.worktrees/feature-skills-consolidation-local`.

## Plan gate

Human Gate: APPROVED by rluisb at 2026-06-15 15:34:44 GMT-3
