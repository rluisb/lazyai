# Issue 245 Plan C — Follow-up Implementation Plan

Date: 2026-06-15
Issue: https://github.com/rluisb/lazyai/issues/245
Basis: `specs/issues/245-parity-audit/parity-report.md`
Status: Implemented follow-up wave; gate attestations are external to this artifact

## Purpose

Turn the approved Plan C scope into one phased implementation plan.

This plan covers the report findings that should become actual repository work, while preserving the current LazyAI boundary:

- no Fortnite/orchestrator/eval/task/workflow default comeback
- no `deployer` or `responder` active default agent addition
- no generic DevOps/Data MCP defaults
- no runtime dependency on `/Users/ricardo/code/vibe-lab`

## Summary

Plan C is a five-phase follow-up wave:

1. Safe content parity
2. Hook/policy authoring parity
3. Issue-workflow parity
4. MCP parity decisions plus bounded implementation
5. Command/catalog alignment cleanup

Phases 1–3 are straightforward embedded-library and scaffolding work.
Phase 4 is the only product-policy-heavy phase and requires a second human decision gate before edits land.
Phase 5 closes the remaining command/catalog ambiguity without widening adapter defaults.

## Technical context

| Aspect | Decision | Rationale |
|---|---|---|
| Delivery shape | Checked-in plan plus phased repo changes | The report is complete; next work is implementation, not more auditing. |
| Evidence source | `parity-report.md` is the authoritative gap list for this wave | Prevents scope drift away from the completed audit. |
| Default adapter surface | Keep unchanged unless a phase explicitly approves an expansion | The report confirmed active adapter output is already neutral LazyAI. |
| New assets | Prefer setup-library assets over canonical-default additions unless parity requires otherwise | Lowest blast radius; avoids accidental default-surface growth. |
| MCP policy | Opt-in only for any newly added MCP servers | Preserves current setup-core defaults and secret-handling expectations. |
| Provenance semantics | Correct current-path mismatches only where human-approved and historically honest | Avoids rewriting import history silently. |
| Verification | Focused package tests per phase; `go test ./packages/cli/...` at the end if any Go code changes land | Matches repo guidance: targeted checks first, broader package verification after seam stability. |

## Design decisions embedded in this plan

These decisions remove ambiguity from the report and keep the implementation boring.

1. `evidence-verifier` lands as a setup-library skill, not a new canonical default agent.
2. `block-destructive-shell` lands as a setup-library hook/policy document, not an active adapter-emitted hook.
3. `constitution-template.md` lands in `packages/cli/library/templates/`.
4. verified-research `artifact-set.md` lands in `packages/cli/library/templates/` as a reusable template asset.
5. hook authoring parity adds `hook` as a supported `lazyai-cli create` artifact type; it does not bring back `workflow`, `domain`, or `mode` artifact creation.
6. `issue-triage` and `task-to-issues` land as setup-library skills, not canonical defaults.
7. MCP additions remain opt-in. Concrete additions require exact supported config shape; no placeholder or guessed server entries.
8. `graphify` and `handoff` canonical command assets stay non-default unless a later explicit product decision changes command emission.

## Planned files and expected surfaces

```text
specs/issues/245-parity-audit/
├── parity-report.md                 ← existing audit artifact
└── follow-up-plan-c.md              ← this plan

packages/cli/library/
├── skills/
│   ├── evidence-verifier.md         ← new, Phase 1
│   ├── issue-triage.md              ← new, Phase 3
│   └── task-to-issues.md            ← new, Phase 3
├── hooks/
│   └── block-destructive-shell.md   ← new, Phase 1
├── templates/
│   ├── constitution-template.md                 ← new, Phase 1
│   ├── verified-research-artifact-set-template.md ← new, Phase 1
│   ├── hook-template.md                           ← new, Phase 2
│   └── policy-template.md                         ← new, Phase 2
└── manifests/
    ├── curation.yaml                ← update in Phases 1–3, maybe 5
    └── provenance.yaml              ← update in Phase 1 and maybe 5

packages/cli/internal/
├── generator/
│   ├── hook.go                      ← new, Phase 2
│   ├── hook_test.go                 ← new, Phase 2
│   └── registry.go                  ← update in Phase 2
├── types/
│   └── types.go                     ← update in Phase 2
├── validation/
│   ├── validation.go                ← update in Phase 2
│   └── validation_test.go           ← update in Phase 2
├── library/
│   └── asset_manifest_test.go       ← update if manifest coverage rules need extension
├── adapter/
│   ├── asset_manifest_test.go       ← update as curation coverage changes
│   └── mcp_compiler_test.go         ← update in Phase 4
└── scaffold/
    └── mcp_test.go                  ← update in Phase 4

packages/cli/cmd/
├── create.go                        ← update in Phase 2
└── command_excision_test.go         ← update in Phase 2

packages/cli/library/mcp/
└── catalog.json                     ← update in Phase 4

docs/concepts/
└── library-manifests.md             ← update in Phase 5 if command/catalog stance is clarified in docs
```

Generated user-project output introduced by Phase 2:

```text
library/hooks/<name>.md
```

That output is for `lazyai-cli create hook <name>` and is separate from the embedded repo library under `packages/cli/library/`.

## Phase 1 — Safe content parity

**Goal:** land the low-risk report gaps that are pure content or manifest work.

### Scope

- Add `evidence-verifier` as a setup-library skill.
- Add `block-destructive-shell` as a setup-library hook/policy document.
- Add `constitution-template.md`.
- Add verified-research artifact-set template.
- Correct approved stale provenance source paths.
- Update curation/provenance manifests accordingly.

### Files

**New**
- `packages/cli/library/skills/evidence-verifier.md`
- `packages/cli/library/hooks/block-destructive-shell.md`
- `packages/cli/library/templates/constitution-template.md`
- `packages/cli/library/templates/verified-research-artifact-set-template.md`

**Changed**
- `packages/cli/library/manifests/curation.yaml`
- `packages/cli/library/manifests/provenance.yaml`

### Change details

1. Compress `/Users/ricardo/code/vibe-lab/.agents/agents/evidence-verifier.md` into a LazyAI-neutral skill focused on claim classification against cited source evidence.
2. Port `/Users/ricardo/code/vibe-lab/.agents/hooks/block-destructive-shell/POLICY.md` into a LazyAI setup-library policy doc without claiming active adapter enforcement.
3. Port `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/constitution-template.md` into the embedded template set.
4. Port `/Users/ricardo/code/vibe-lab/.agents/workflows/verified-research/templates/artifact-set.md` into a reusable template asset with LazyAI naming and path examples.
5. Update `curation.yaml` coverage for the new skill/hook/template files.
6. Update `provenance.yaml` only for the approved stale current-path mismatches already called out in the report. Keep any historical notes that must remain historical.

### Verification

Run after Phase 1 edits:

- `go test ./packages/cli/internal/library ./packages/cli/internal/adapter`

Conditional:
- Run `go run ./packages/cli/internal/tokenrent/cmd/token-rent-check` only if Phase 1 ends up touching canonical library assets instead of the planned setup-library/template paths.

### Exit

- New files exist at the planned paths.
- Both manifests validate.
- No adapter output mapping changed.

## Phase 2 — Hook/policy authoring parity

**Goal:** make hook creation a first-class `lazyai-cli create` artifact without reviving retired artifact types.

### Scope

- Add `hook` artifact support to `lazyai-cli create`.
- Add embedded hook/policy templates that support that flow.
- Keep workflow/domain/mode creation rejected.

### Files

**New**
- `packages/cli/internal/generator/hook.go`
- `packages/cli/internal/generator/hook_test.go`
- `packages/cli/library/templates/hook-template.md`
- `packages/cli/library/templates/policy-template.md`

**Changed**
- `packages/cli/internal/generator/registry.go`
- `packages/cli/internal/types/types.go`
- `packages/cli/internal/validation/validation.go`
- `packages/cli/internal/validation/validation_test.go`
- `packages/cli/cmd/create.go`
- `packages/cli/cmd/command_excision_test.go`
- `packages/cli/library/manifests/curation.yaml`
- `packages/cli/internal/adapter/asset_manifest_test.go` if template adapter targets need expectations updated

### Change details

1. Add `ArtifactTypeHook` to `packages/cli/internal/types/types.go`.
2. Register the new artifact type in `validation.go` and tests.
3. Extend `create.go` interactive options and valid type checking to include `hook` only.
4. Keep the existing test that rejects `workflow`, `domain`, and `mode`; update error text expectations to include `hook` in the allowed set.
5. Add a `HookGenerator` that writes `library/hooks/<name>.md` in generated projects.
6. Add embedded `hook-template.md` and `policy-template.md` so the generated hook shape follows one canonical pattern.
7. Update curation coverage for the new embedded templates.

### Verification

Run after Phase 2 edits:

- `go test ./packages/cli/internal/generator ./packages/cli/internal/validation ./packages/cli/cmd ./packages/cli/internal/library ./packages/cli/internal/adapter`

### Exit

- `lazyai-cli create hook <name>` is a supported artifact type.
- `workflow`, `domain`, and `mode` remain rejected.
- Embedded templates are covered by curation.

## Phase 3 — Issue-workflow parity

**Goal:** add the missing issue-oriented setup-library skills without expanding canonical defaults.

### Scope

- Add `issue-triage`.
- Add `task-to-issues`.
- Keep both as setup-library skills only.

### Files

**New**
- `packages/cli/library/skills/issue-triage.md`
- `packages/cli/library/skills/task-to-issues.md`

**Changed**
- `packages/cli/library/manifests/curation.yaml`

### Change details

1. Compress `/Users/ricardo/code/vibe-lab/.agents/skills/issue-triage/SKILL.md` into a LazyAI setup-library skill.
2. Compress `/Users/ricardo/code/vibe-lab/.agents/skills/task-to-issues/SKILL.md` into a LazyAI setup-library skill.
3. Record both files in `curation.yaml` as setup-core support, adapter target `none`.

### Verification

Run after Phase 3 edits:

- `go test ./packages/cli/internal/library ./packages/cli/internal/adapter`

### Exit

- Both skills exist and are covered by curation.
- Canonical/default skill set remains unchanged.

## Gate A — Human decision before Phase 4

⛔ Required before MCP implementation.

Questions that must be answered explicitly:

1. Add Context7 as opt-in remote MCP entry: yes or no?
2. Add GitHub MCP as opt-in MCP entry in addition to existing `gh` CLI: yes or no?
3. For Figma and Slack, do we require exact verified server shape before adding anything, or do we close them as documented exclusions for now?
4. Should new MCP entries default to `enabled: false`? This plan assumes yes.

No Phase 4 implementation starts until these answers are approved.

## Phase 4 — MCP parity decisions plus bounded implementation

**Goal:** implement the MCP portion of Plan C without speculative servers.

### Scope

- Add only MCP entries with exact supported server shape.
- Keep all new entries opt-in by default.
- Do not invent placeholder Figma/Slack entries.

### Files

**Changed**
- `packages/cli/library/mcp/catalog.json`
- `packages/cli/internal/adapter/mcp_compiler_test.go`
- `packages/cli/internal/scaffold/mcp_test.go`
- `docs/concepts/library-manifests.md` if the opt-in MCP policy needs explicit documentation

**Conditional**
- `packages/cli/cmd/server.go` only if new catalog metadata needs list/report handling changes
- `packages/cli/cmd/server_*test.go` if command behavior changes and tests exist or are added

### Change details

1. Add `context7` only if Gate A approves it. Use the exact remote MCP pattern already evidenced in vibe-lab.
2. Add `github` MCP only if Gate A approves it. Keep `gh` CLI support; do not remove it.
3. Figma and Slack require exact verified server shape before catalog additions. If Gate A says “no fresh evidence, no addition,” record them as explicit exclusions in docs instead of adding guessed entries.
4. Ensure any new MCP entry remains opt-in (`enabled: false`) unless a separate product decision says otherwise.
5. Update compiler/scaffold tests to assert emitted shape for any newly added server entries.

### Verification

Run after Phase 4 edits:

- `go test ./packages/cli/internal/adapter ./packages/cli/internal/scaffold ./packages/cli/cmd`

### Exit

- Any added MCP entry has an exact supported shape.
- No speculative Figma/Slack placeholders landed.
- Existing enabled/default MCP behavior remains unchanged unless explicitly approved.

## Phase 5 — Command/catalog alignment cleanup

**Goal:** close the remaining ambiguity around canonical commands and document the final post-Plan-C stance.

### Scope

- Resolve the `graphify` / `handoff` command-asset ambiguity found in the report.
- Align manifest/docs language with actual adapter behavior.
- Avoid widening command emission by default unless explicitly approved.

### Files

**Preferred changed set**
- `docs/concepts/library-manifests.md`
- `packages/cli/library/manifests/curation.yaml` if command exclusions or notes need clarification

**Conditional, only if a product decision changes command emission**
- `packages/cli/internal/adapter/output_mapping.go`
- `packages/cli/internal/adapter/*command*test*.go` or adjacent adapter tests
- `packages/cli/library/claudecode/commands/`
- `packages/cli/library/opencode/commands/`

### Change details

Preferred safe path:
1. Keep `packages/cli/library/canonical/commands/{graphify,handoff}.md` as provenance-covered command assets that are not active command emission sources.
2. Clarify that actual command emission still comes from tool-specific command directories.
3. Avoid changing `output_mapping.go` unless a separate human-approved product decision explicitly wants canonical command emission.

If a human later approves emitted canonical commands, treat that as a sub-project with its own focused implementation branch and verification.

### Verification

Safe-path verification:
- `go test ./packages/cli/internal/library ./packages/cli/internal/adapter`

If command emission changes:
- `go test ./packages/cli/internal/adapter ./packages/cli/cmd`
- `go build ./packages/cli/...`

### Exit

- The docs and manifests no longer imply that canonical command files are active command emission sources when they are not.
- No accidental command-surface widening occurred.

## Final verification

After all approved phases that touch Go code:

- `go test ./packages/cli/...`

Conditional:
- `go run ./packages/cli/internal/tokenrent/cmd/token-rent-check` only if the approved implementation touches canonical library assets.
- `go build ./packages/cli/...` only if command/adaptor code changes justify a build seam beyond tests.

## Out of scope

- Adding `deployer` or `responder` to active defaults
- Reintroducing `task`, `workflow`, `orchestration`, `mcp-setup`, obsolete `eval`, Fortnite, or retired orchestrator defaults
- Adding generic Docker/Kubernetes/Terraform/PostgreSQL/SQLite/Redis MCP defaults
- Replacing `gh` CLI with GitHub MCP by force
- Inventing Figma or Slack server definitions without exact supported shape
- Touching `packages/cli/internal/db/migrations.go`
- Touching `.worktrees/feature-skills-consolidation-local`

## Human gates

- ⛔ Gate 1: approve this plan before any implementation work
- ⛔ Gate 2: answer the MCP policy questions before Phase 4
- ⛔ Gate 3: if Phase 5 would change command emission semantics, approve that separately before changing `output_mapping.go`
