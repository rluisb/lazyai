# Issue 245 LazyAI / vibe-lab Parity Research

Date: 2026-06-15
Issue: https://github.com/rluisb/lazyai/issues/245
Phase: RPI Research only

## Scope

Research a 1:1 parity audit between LazyAI and the local vibe-lab baseline. This phase records evidence and open classification questions only. It does not port assets, edit manifests, change adapter output, or create follow-up implementation issues.

## Hard constraints

- LazyAI is the runtime/product. vibe-lab supplies principles, setup assets, and adapter expectations only.
- LazyAI must not depend on `/Users/ricardo/code/vibe-lab` at runtime.
- Active LazyAI defaults must not reintroduce retired Fortnite/orchestrator/eval/task/workflow surfaces.
- `packages/cli/internal/db/migrations.go` remains out of scope unless a human explicitly approves migration edits.
- RPI gates apply: research must be approved before planning.

## Evidence sources

### LazyAI

- `issue://245`
- `docs/concepts/product-boundaries.md`
- `docs/concepts/library-manifests.md`
- `packages/cli/library/manifests/provenance.yaml`
- `packages/cli/library/manifests/curation.yaml`
- `packages/cli/internal/adapter/output_mapping.go`
- `packages/cli/library/mcp/catalog.json`
- `packages/cli/library/canonical/`
- `packages/cli/library/skills/`
- `packages/cli/library/hooks/`
- `packages/cli/library/rules/`
- `packages/cli/library/standards/starter/`
- `packages/cli/library/specs-agents/`
- `packages/cli/library/templates/`
- `packages/cli/library/claudecode/commands/`
- `packages/cli/library/opencode/commands/`
- `packages/cli/library/prompts/`
- `packages/cli/library/chatmodes/`
- `archive/issue-244-historical-library/`

### vibe-lab baseline checkout

- `/Users/ricardo/code/vibe-lab/.agents/agents/`
- `/Users/ricardo/code/vibe-lab/.claude/agents/`
- `/Users/ricardo/code/vibe-lab/.opencode/agents/`
- `/Users/ricardo/code/vibe-lab/.github/agents/`
- `/Users/ricardo/code/vibe-lab/.agents/skills/`
- `/Users/ricardo/code/vibe-lab/docs/skills/`
- `/Users/ricardo/code/vibe-lab/.agents/hooks/`
- `/Users/ricardo/code/vibe-lab/.claude/hooks/`
- `/Users/ricardo/code/vibe-lab/.gemini/hooks/vibe-lab/`
- `/Users/ricardo/code/vibe-lab/.github/hooks/`
- `/Users/ricardo/code/vibe-lab/canonical/`
- `/Users/ricardo/code/vibe-lab/.agents/workflows/`
- `/Users/ricardo/code/vibe-lab/docs/workflows/`
- `/Users/ricardo/code/vibe-lab/specs/_templates/`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/`
- `/Users/ricardo/code/vibe-lab/bin/`

## LazyAI observed inventory

### Product boundary

`docs/concepts/product-boundaries.md` defines these active categories:

- `setup-core`: setup creation, validation, compile, update, manifests, adapters, generated tool-native files, workspace/scope resolution.
- `ops-runtime-extra`: local runtime-adjacent state such as sessions, ledger, memory, messages, metrics, costs, secrets, notifications, backups, auth probes.
- `dev-harness`: repository maintenance scripts and generated catalog maintenance.
- `retired/archived`: migration, rollback, historical, or deprecated compatibility paths.

The same document explicitly says repository scripts under `bin/` are not shipped product commands, and removed command surfaces such as `task`, `workflow`, `orchestration`, `mcp-setup`, and obsolete `eval` are not active top-level CLI commands.

### Canonical adapter-emitted agents

| Agent | Path | Current manifest state |
|---|---|---|
| `builder` | `packages/cli/library/canonical/agents/builder.md` | `adapter-support`, compressed |
| `planner` | `packages/cli/library/canonical/agents/planner.md` | `adapter-support`, compressed |
| `primary-agent` | `packages/cli/library/canonical/agents/primary-agent.md` | `setup-core`, LazyAI-authored |
| `reviewer` | `packages/cli/library/canonical/agents/reviewer.md` | `adapter-support`, compressed |
| `scout` | `packages/cli/library/canonical/agents/scout.md` | `adapter-support`, compressed |

Archived historical/non-default agents after issue #244:

- `archive/issue-244-historical-library/packages/cli/library/agents/builder.md`
- `archive/issue-244-historical-library/packages/cli/library/agents/documenter.md`
- `archive/issue-244-historical-library/packages/cli/library/agents/implementor.md`
- `archive/issue-244-historical-library/packages/cli/library/agents/orchestrator.md`
- `archive/issue-244-historical-library/packages/cli/library/agents/planner.md`
- `archive/issue-244-historical-library/packages/cli/library/agents/red-team.md`
- `archive/issue-244-historical-library/packages/cli/library/agents/reviewer.md`
- `archive/issue-244-historical-library/packages/cli/library/agents/scout.md`

### Canonical adapter-emitted skills

| Skill | Path | Current manifest state |
|---|---|---|
| `codebase-exploration` | `packages/cli/library/canonical/skills/codebase-exploration.md` | `adapter-support`, compressed |
| `diagnose` | `packages/cli/library/canonical/skills/diagnose.md` | `adapter-support`, compressed |
| `pr-review` | `packages/cli/library/canonical/skills/pr-review.md` | `adapter-support`, compressed |
| `test-first-change` | `packages/cli/library/canonical/skills/test-first-change.md` | `adapter-support`, compressed |

### Setup-library skills, not canonical adapter defaults

`packages/cli/library/skills/` contains:

- `anti-speculation.md`
- `bugfix.md`
- `chain-verify.md`
- `diagnose.md`
- `extract-standards.md`
- `housekeeping.md`
- `impact-check.md`
- `implement.md`
- `improve-codebase-architecture.md`
- `iterate.md`
- `memory-write.md`
- `parallel-execution.md`
- `plan.md`
- `populate/SKILL.md`
- `process-audit.md`
- `proof-of-concept.md`
- `red-team-plan.md`
- `research.md`
- `review.md`
- `rpi.md`
- `self-improve.md`
- `speckit-analyze.md`
- `speckit-checklist.md`
- `speckit-clarify.md`
- `speckit-constitution.md`
- `speckit-implement.md`
- `speckit-plan.md`
- `speckit-specify.md`
- `speckit-tasks.md`
- `spike.md`
- `tdd-loop.md`
- `update-memory.md`

Issue #244 archived `packages/cli/library/skills/orchestrate.md` to `archive/issue-244-historical-library/packages/cli/library/skills/orchestrate.md`.

### Hooks and policies

Canonical hooks:

- `packages/cli/library/canonical/hooks/pre-commit.md`
- `packages/cli/library/canonical/hooks/session-start.md`

Setup-library hooks:

- `packages/cli/library/hooks/pre-commit`
- `packages/cli/library/hooks/rpi-gate-check.yml`

### Rules, standards, protocols

Setup rules under `packages/cli/library/rules/`:

- `access.md`
- `agent-security.md`
- `agent-state.md`
- `auto-recovery.md`
- `code-style.md`
- `cost.md`
- `review.md`
- `security.md`
- `self-consistency.md`
- `structured-feedback.md`
- `testing.md`
- `tool-use.md`
- `typescript.md`
- `workflow.md`

Starter standards under `packages/cli/library/standards/starter/`:

- `agent-security.md`
- `context-loading.md`
- `error-handling.md`
- `test-patterns.md`

Issue #244 archived `packages/cli/library/standards/starter/orchestration-patterns.md` to `archive/issue-244-historical-library/packages/cli/library/standards/starter/orchestration-patterns.md`.

### Workflow/spec agent docs

`packages/cli/library/specs-agents/` contains:

- `adrs.md`
- `bugfixes.md`
- `features.md`
- `memory.md`
- `prompts.md`
- `refactors.md`
- `rules.md`
- `standards.md`
- `tech-debt.md`
- `templates.md`
- `workflows.md`

These are setup documentation assets, not retired runtime workflow command surfaces.

### Templates

`packages/cli/library/templates/` contains:

- `adr.md`
- `audit-template.md`
- `bugfix-rca-template.md`
- `checklist-template.md`
- `code-review-template.md`
- `housekeeping-template.md`
- `ledger-template.md`
- `plan-template.md`
- `poc-template.md`
- `postmortem-template.md`
- `spec-template.md`
- `spike-template.md`
- `standard.md`
- `task-harness-template.md`
- `tasks-template.md`
- `tech-debt-template.md`

### MCP catalog

`packages/cli/library/mcp/catalog.json` contains enabled servers:

- `memory`
- `filesystem`
- `ripgrep`
- `memoria`
- `codegraph`
- `qmd`
- `graphify`
- `obsidian`

Optional or disabled servers:

- `playwright`
- `atlassian`
- `fetch`

CLI tools listed in the catalog:

- `acli`
- `gh`
- `rtk`
- `codegraph`
- `qmd`
- `graphify`
- `ob`
- `memoria`
- `playwright`
- `fetch`

### Adapter output mapping

`packages/cli/internal/adapter/output_mapping.go` maps active adapter output:

| Tool | Agents source → destination | Skills source → destination | Commands source → destination | Other output |
|---|---|---|---|---|
| Claude Code | `canonical/agents` → `.claude/agents` | `canonical/skills` → `.claude/skills/<name>/SKILL.md` | `claudecode/commands` → `.claude/commands` | `templates` → `.claude/templates`, `claudecode/output-styles` → `.claude/output-styles` |
| OpenCode | `canonical/agents` → `.opencode/agents` | `canonical/skills` → `.opencode/skills/<name>/SKILL.md` | `opencode/commands` → `.opencode/commands` | `templates` → `.opencode/templates`, `opencode/modes` → `.opencode/modes` |
| Copilot | `canonical/agents` → `.github/agents/*.agent.yaml` | `canonical/skills` → `.github/agents/*.agent.yaml` | no slash command surface | `templates` → `.github/instructions`, `chatmodes` → `.github/chatmodes`, `prompts` → `.github/prompts/*.prompt.md` |

Adapter command assets:

- Canonical: `packages/cli/library/canonical/commands/graphify.md`, `packages/cli/library/canonical/commands/handoff.md`
- Claude Code: `packages/cli/library/claudecode/commands/init.md`, `review.md`, `test.md`, `commit.md`, and `speckit.*.md`
- OpenCode: `packages/cli/library/opencode/commands/init.md`, `review.md`, `test.md`, `commit.md`, and `speckit.*.md`
- Copilot: no slash commands; uses generated agents, instructions, chatmodes, and prompts.

### CLI command surface

`docs/concepts/product-boundaries.md` records active top-level `lazyai-cli` commands and categories. Active setup-core examples include `init`, `compile`, `doctor`, `update`, `status`, `server`, `workspace`, `build-plugin`, and `validate`. Runtime-adjacent commands include `session`, `message`, `ledger`, `memory`, `metrics`, `cost`, `secret`, `notify`, `backup`, `restore-runtime-db`, and `auth`. `completions` is hidden/deprecated in `packages/cli/cmd/completions.go` and replaced by `completion`.

## vibe-lab observed inventory

### Agents

Baseline agents under `/Users/ricardo/code/vibe-lab/.agents/agents/`:

- `deployer.md`
- `evidence-verifier.md`
- `implementer.md`
- `planner.md`
- `researcher.md`
- `responder.md`
- `reviewer.md`

Duplicate tool-specific copies exist under:

- `/Users/ricardo/code/vibe-lab/.claude/agents/`
- `/Users/ricardo/code/vibe-lab/.opencode/agents/`
- `/Users/ricardo/code/vibe-lab/.github/agents/`

### Skills

Baseline Agent Skills under `/Users/ricardo/code/vibe-lab/.agents/skills/`:

- `adhd-engineer/SKILL.md`
- `architecture-review/SKILL.md`
- `caveman/SKILL.md`
- `codebase-exploration/SKILL.md`
- `create-agent/SKILL.md`
- `create-hook/SKILL.md`
- `create-skill/SKILL.md`
- `create-workflow/SKILL.md`
- `diagnose/SKILL.md`
- `doc-backed-clarify/SKILL.md`
- `fast-feedback/SKILL.md`
- `four-point-vibe-coding/SKILL.md`
- `handoff/SKILL.md`
- `issue-triage/SKILL.md`
- `memory-promotion/SKILL.md`
- `no-workarounds/SKILL.md`
- `project-guardrails-init/SKILL.md`
- `skill-authoring/SKILL.md`
- `task-to-issues/SKILL.md`
- `tdd-planning/SKILL.md`
- `test-first-change/SKILL.md`
- `zoom-out/SKILL.md`

Rendered docs also exist under `/Users/ricardo/code/vibe-lab/docs/skills/`.

### Hooks / policies

Canonical hook policies under `/Users/ricardo/code/vibe-lab/.agents/hooks/`:

- `block-destructive-shell/POLICY.md`
- `caveman-memory-promotion/POLICY.md`
- `objective-workflow-gate/POLICY.md`
- `startup-self-heal/POLICY.md`

Tool-specific hook scripts/configs exist under:

- `/Users/ricardo/code/vibe-lab/.claude/hooks/`
- `/Users/ricardo/code/vibe-lab/.gemini/hooks/vibe-lab/`
- `/Users/ricardo/code/vibe-lab/.github/hooks/`

### Canonical rules, protocols, and templates

`/Users/ricardo/code/vibe-lab/canonical/` contains:

- `agent-template.md`
- `artifact-rules.md`
- `caveman-ai-memory.md`
- `clarification-levels.md`
- `clean-code.md`
- `constitution.md`
- `cookbook-recipe.md`
- `engineering-principles.md`
- `hook-template.md`
- `learning-template.md`
- `llm-engineering-relationships.md`
- `mcp-setup.md`
- `policy-template.md`
- `prd-plan-todo.md`
- `self-assessment.md`
- `skill-template.md`
- `tdd-planning.md`
- `workflow-template.md`

### Workflows

Workflow definitions exist under both `/Users/ricardo/code/vibe-lab/.agents/workflows/` and `/Users/ricardo/code/vibe-lab/docs/workflows/`:

- `adversarial-review.md`
- `bugfix.md`
- `code-review.md`
- `documentation.md`
- `feature.md`
- `refactor.md`
- `spike.md`
- `verified-research.md`

Verified-research templates:

- `/Users/ricardo/code/vibe-lab/.agents/workflows/verified-research/templates/artifact-set.md`
- `/Users/ricardo/code/vibe-lab/.agents/workflows/verified-research/templates/checklist.md`
- `/Users/ricardo/code/vibe-lab/.agents/workflows/verified-research/templates/prompt-template.md`

### Templates and presets

Spec templates:

- `/Users/ricardo/code/vibe-lab/specs/_templates/prd.md`
- `/Users/ricardo/code/vibe-lab/specs/_templates/tasks.md`
- `/Users/ricardo/code/vibe-lab/specs/_templates/techspec.md`

Speckit preset:

- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/preset.yml`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/checklist-template.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/constitution-template.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/plan-template.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/spec-template.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/tasks-template.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.analyze.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.checklist.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.clarify.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.constitution.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.implement.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.plan.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.specify.md`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.tasks.md`

### Scripts

`/Users/ricardo/code/vibe-lab/bin/` contains:

- `bootstrap-project`
- `doctor`
- `inject`
- `inject.original`
- `startup-self-heal`

### MCP guidance

`/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` supports:

- Claude Code project `.mcp.json`
- OpenCode `mcp` block in existing `opencode.json` or `opencode.jsonc`
- OMP/Pi marked unsupported for project-local MCP config in vibe-lab

MCP categories in the baseline doc:

- Context: Context7, memory, filesystem
- Code Quality: linter, test runner, coverage
- DevOps: Docker, Kubernetes, Terraform
- Data: PostgreSQL, SQLite, Redis
- External: GitHub, Figma, Playwright, Slack

## Preliminary deltas to classify in the Plan phase

### Agents

| vibe-lab item | vibe-lab source | LazyAI current state | Research note |
|---|---|---|---|
| `planner` | `/Users/ricardo/code/vibe-lab/.agents/agents/planner.md` | `packages/cli/library/canonical/agents/planner.md` | likely implemented active default/compressed |
| `reviewer` | `/Users/ricardo/code/vibe-lab/.agents/agents/reviewer.md` | `packages/cli/library/canonical/agents/reviewer.md` | likely implemented active default/compressed |
| `implementer` | `/Users/ricardo/code/vibe-lab/.agents/agents/implementer.md` | `packages/cli/library/canonical/agents/builder.md`; archived `archive/issue-244-historical-library/packages/cli/library/agents/implementor.md` | likely renamed/equivalent, but name differs and archived `implementor` is historical |
| `researcher` | `/Users/ricardo/code/vibe-lab/.agents/agents/researcher.md` | `packages/cli/library/canonical/agents/scout.md` | likely renamed/equivalent |
| `deployer` | `/Users/ricardo/code/vibe-lab/.agents/agents/deployer.md` | no observed active LazyAI canonical agent | classify as missing, intentionally excluded, or not product scope |
| `evidence-verifier` | `/Users/ricardo/code/vibe-lab/.agents/agents/evidence-verifier.md` | no observed active LazyAI canonical agent | classify as missing or covered by reviewer/test-first contracts |
| `responder` | `/Users/ricardo/code/vibe-lab/.agents/agents/responder.md` | no observed active LazyAI canonical agent | classify as missing or intentionally excluded |

LazyAI-only active agents requiring rationale:

- `packages/cli/library/canonical/agents/primary-agent.md`: LazyAI runtime entry point.
- `packages/cli/library/canonical/agents/builder.md`: likely renamed/equivalent for vibe-lab `implementer`.
- `packages/cli/library/canonical/agents/scout.md`: likely renamed/equivalent for vibe-lab `researcher`.

### Skills

Likely active defaults already represented:

- `codebase-exploration`: vibe-lab `.agents/skills/codebase-exploration/SKILL.md`; LazyAI `packages/cli/library/canonical/skills/codebase-exploration.md`
- `diagnose`: vibe-lab `.agents/skills/diagnose/SKILL.md`; LazyAI `packages/cli/library/canonical/skills/diagnose.md`
- `test-first-change`: vibe-lab `.agents/skills/test-first-change/SKILL.md`; LazyAI `packages/cli/library/canonical/skills/test-first-change.md`

vibe-lab skills with no observed LazyAI canonical or setup-library path by name:

- `adhd-engineer`
- `architecture-review`
- `caveman`
- `create-agent`
- `create-hook`
- `create-skill`
- `create-workflow`
- `doc-backed-clarify`
- `fast-feedback`
- `four-point-vibe-coding`
- `handoff`
- `issue-triage`
- `memory-promotion`
- `no-workarounds`
- `project-guardrails-init`
- `skill-authoring`
- `task-to-issues`
- `tdd-planning`
- `zoom-out`

LazyAI-only setup-library skills requiring rationale include `rpi`, `plan`, `research`, `implement`, `review`, `bugfix`, `spike`, `parallel-execution`, `tdd-loop`, `speckit-*`, `populate`, and memory/housekeeping/process-audit variants.

### Hooks

vibe-lab hooks with no LazyAI same-name asset:

- `block-destructive-shell`
- `caveman-memory-promotion`
- `objective-workflow-gate`
- `startup-self-heal`

LazyAI has `packages/cli/library/canonical/hooks/pre-commit.md`, `packages/cli/library/canonical/hooks/session-start.md`, `packages/cli/library/hooks/pre-commit`, and `packages/cli/library/hooks/rpi-gate-check.yml`. Plan must decide whether these are equivalents, partial equivalents, or separate LazyAI-only surfaces.

### Rules / standards / protocols

vibe-lab canonical docs are not 1:1 represented by LazyAI rules. Some concepts may be folded into LazyAI root templates, setup rules, standards, or skills. Plan must define a mapping table rather than using name-only matching.

High-value mapping candidates:

- `canonical/clean-code.md` → LazyAI `packages/cli/library/rules/code-style.md`, `testing.md`, `tool-use.md`, and agent prompts.
- `canonical/tdd-planning.md` → LazyAI `packages/cli/library/skills/tdd-loop.md`, `canonical/skills/test-first-change.md`, and setup rules.
- `canonical/mcp-setup.md` → LazyAI `packages/cli/library/mcp/catalog.json` plus MCP compiler behavior.
- `canonical/agent-template.md`, `skill-template.md`, `hook-template.md`, `policy-template.md`, `workflow-template.md` → LazyAI `packages/cli/library/templates/`, `tool-templates/`, `tool-agents/`, or artifact creation commands.

### Workflows

vibe-lab workflow names are not present under `packages/cli/library/specs-agents/` by name except broad categories such as `bugfixes`, `features`, and `refactors`. Plan must classify whether LazyAI `specs-agents` files are equivalent setup guides, whether vibe-lab workflows should become setup-library workflow docs, or whether they are intentionally excluded.

### Templates

Likely matched or near-matched:

- vibe-lab `spec-template.md` → LazyAI `packages/cli/library/templates/spec-template.md`
- vibe-lab `plan-template.md` → LazyAI `packages/cli/library/templates/plan-template.md`
- vibe-lab `tasks-template.md` → LazyAI `packages/cli/library/templates/tasks-template.md`
- vibe-lab `checklist-template.md` → LazyAI `packages/cli/library/templates/checklist-template.md`

Potential missing by name:

- vibe-lab `canonical/speckit-vibe-lab-preset/templates/constitution-template.md`
- vibe-lab `specs/_templates/prd.md`
- vibe-lab `specs/_templates/techspec.md`

LazyAI-only templates requiring rationale include `adr.md`, `audit-template.md`, `bugfix-rca-template.md`, `code-review-template.md`, `housekeeping-template.md`, `ledger-template.md`, `poc-template.md`, `postmortem-template.md`, `spike-template.md`, `standard.md`, `task-harness-template.md`, and `tech-debt-template.md`.

### Commands / scripts

vibe-lab scripts under `/Users/ricardo/code/vibe-lab/bin/` should not be assumed product commands. LazyAI has Go CLI equivalents or repository harness boundaries:

- `doctor`: LazyAI product command `packages/cli/cmd/doctor.go`; vibe-lab script remains baseline script evidence.
- `inject`: LazyAI adapter compiler/output mapping likely replaces script behavior.
- `bootstrap-project`: likely init/setup-adjacent; no direct LazyAI script observed.
- `startup-self-heal`: vibe-lab hook/script; LazyAI has repository `bin/startup-self-heal` and canonical/session-start hooks, but product classification is open.
- `inject.original`: likely historical baseline script, not candidate runtime behavior without evidence.

### MCP

LazyAI has a concrete catalog, while vibe-lab has category guidance. The parity report must compare categories, not just names:

| vibe-lab MCP category | Examples in vibe-lab | LazyAI observed state |
|---|---|---|
| Context | Context7, memory, filesystem | memory/filesystem present; Context7 not observed in LazyAI catalog |
| Code Quality | linter, test runner, coverage | no explicit catalog server observed; may be CLI/test workflow rather than MCP |
| DevOps | Docker, Kubernetes, Terraform | no explicit catalog server observed |
| Data | PostgreSQL, SQLite, Redis | no explicit catalog server observed |
| External | GitHub, Figma, Playwright, Slack | Playwright optional; GitHub appears as CLI tool `gh`; Figma/Slack not observed |

## Stale or conflicting evidence to resolve

`packages/cli/library/manifests/provenance.yaml` points at several vibe-lab source paths that are not present in the current `/Users/ricardo/code/vibe-lab` checkout:

| LazyAI path | Provenance source path | Current checkout result |
|---|---|---|
| `packages/cli/library/canonical/agents/builder.md` | `.agents/agents/builder.md` | missing; current analogous path appears to be `.agents/agents/implementer.md` |
| `packages/cli/library/canonical/agents/scout.md` | `.agents/agents/scout.md` | missing; current analogous path appears to be `.agents/agents/researcher.md` |
| `packages/cli/library/canonical/commands/graphify.md` | `commands/graphify.md` | missing in observed baseline search |
| `packages/cli/library/canonical/commands/handoff.md` | `commands/handoff.md` | missing; current handoff exists as `.agents/skills/handoff/SKILL.md` and docs variants |
| `packages/cli/library/canonical/hooks/session-start.md` | `hooks/session-start.md` | missing in observed baseline search |
| `packages/cli/library/canonical/skills/codebase-exploration.md` | `skills/codebase-exploration/SKILL.md` | missing; current path is `.agents/skills/codebase-exploration/SKILL.md` |
| `packages/cli/library/canonical/skills/diagnose.md` | `skills/diagnose/SKILL.md` | missing; current path is `.agents/skills/diagnose/SKILL.md` |
| `packages/cli/library/canonical/skills/pr-review.md` | `skills/pr-review/SKILL.md` | missing in current vibe-lab baseline |
| `packages/cli/library/canonical/skills/test-first-change.md` | `skills/test-first-change/SKILL.md` | missing; current path is `.agents/skills/test-first-change/SKILL.md` |

This does not prove the LazyAI assets are wrong; the provenance file says several imports are pre-provenance with exact upstream commit unknown. It does mean the parity report must distinguish historical source notes from current baseline source paths.

## Open classification questions for Planning

1. Should the parity report classify `builder` and `scout` as renamed/equivalent to vibe-lab `implementer` and `researcher`, or as LazyAI-only agents with historical inspiration?
2. Should `deployer`, `evidence-verifier`, and `responder` be candidates for canonical adapter output, setup-library-only assets, or intentional exclusions?
3. Should vibe-lab authoring skills (`create-agent`, `create-hook`, `create-skill`, `create-workflow`, `skill-authoring`) become LazyAI setup-library skills, or stay out because LazyAI already has `lazyai-cli create`?
4. Should hook policies be represented as LazyAI hook assets, security rules, pre-commit checks, session-start checks, or excluded as vibe-lab harness behavior?
5. Should MCP parity be category-level only, or should the report recommend concrete catalog additions such as Context7, GitHub, Slack, Figma, Docker, Kubernetes, Terraform, PostgreSQL, SQLite, and Redis?
6. Should workflow docs from vibe-lab become LazyAI setup-library workflow docs, or should LazyAI keep only specs-agent guides and Speckit command assets?
7. Should `constitution-template.md`, `prd.md`, and `techspec.md` be added to LazyAI templates, mapped to existing equivalents, or intentionally excluded?
8. Should `provenance.yaml` source paths be corrected during the implementation phase, or should the parity report only document current-path mismatches and defer manifest changes?

## Research conclusion

Proceed to Planning only after human approval. The next phase should design a checked-in parity report, not asset ports. The report should use exact path-to-path rows and these classification values from issue #245:

- implemented active default
- implemented non-default/setup-library
- renamed/equivalent
- intentionally excluded
- missing and should be ported/curated
- obsolete because LazyAI runtime replaced it
- LazyAI runtime-specific
- setup-core extension
- compatibility/historical material
- candidate for removal/archive

Recommended implementation target for the report, if approved in Plan: `specs/issues/245-parity-audit/parity-report.md`.

⛔ Human gate: approve this research before planning.
