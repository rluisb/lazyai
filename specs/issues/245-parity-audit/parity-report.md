# Issue 245 LazyAI / vibe-lab Parity Report

Date: 2026-06-15
Issue: https://github.com/rluisb/lazyai/issues/245
Research: `specs/issues/245-parity-audit/research.md`
Plan: `specs/issues/245-parity-audit/plan.md`
Baseline checkout: `/Users/ricardo/code/vibe-lab`

## Scope and evidence lock

This report audits current LazyAI asset parity against the local vibe-lab checkout named in issue #245. vibe-lab paths are evidence paths only; LazyAI must not depend on `/Users/ricardo/code/vibe-lab` at runtime.

Initial implementation for this issue was documentation-only. The subsequent Plan C follow-up changed setup-library assets, hook creation support, MCP catalog entries/tests, and documentation; it did not change adapter output mappings or migrations.

Evidence used:

- `issue://245`
- `docs/concepts/product-boundaries.md`
- `docs/concepts/library-manifests.md`
- `packages/cli/internal/adapter/output_mapping.go`
- `packages/cli/library/mcp/catalog.json`
- `packages/cli/library/manifests/provenance.yaml`
- `packages/cli/library/manifests/curation.yaml`
- `packages/cli/library/canonical/`
- `packages/cli/library/claudecode/commands/`
- `packages/cli/library/opencode/commands/`
- `packages/cli/library/templates/`
- `packages/cli/library/skills/`
- `packages/cli/library/hooks/`
- `packages/cli/library/rules/`
- `packages/cli/library/standards/starter/`
- `packages/cli/library/specs-agents/`
- `/Users/ricardo/code/vibe-lab/.agents/agents/`
- `/Users/ricardo/code/vibe-lab/.agents/skills/`
- `/Users/ricardo/code/vibe-lab/.agents/hooks/`
- `/Users/ricardo/code/vibe-lab/.agents/workflows/`
- `/Users/ricardo/code/vibe-lab/canonical/`
- `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/`
- `/Users/ricardo/code/vibe-lab/specs/_templates/`
- `/Users/ricardo/code/vibe-lab/bin/`

## Executive summary

- Active LazyAI adapter output remains neutral LazyAI. `output_mapping.go` emits canonical agents, canonical skills, templates, and tool-specific command assets; it does not emit Fortnite, orchestrator, eval, task, workflow, or orchestration defaults.
- The highest-confidence active default matches are `planner`, `reviewer`, `codebase-exploration`, `diagnose`, `test-first-change`, and the Speckit command/template set.
- The main agent gaps are `evidence-verifier` as a possible follow-up candidate and deliberate exclusions for deploy/SRE-specific agents (`deployer`, `responder`) from active setup defaults.
- The Plan C MCP follow-up adds disabled opt-in Context7 and GitHub remote MCP entries. Figma and Slack are closed as documented exclusions until exact supported server shapes are verified and separately approved.
- The main hook-policy gap is `block-destructive-shell`; LazyAI has security/tool-use rules and repository hooks, but no equivalent baseline hook policy asset.
- Several provenance paths point at historical pre-provenance paths that do not exist in the current vibe-lab checkout. This report documents them; it does not rewrite manifest history.

Baseline matrix counts:

| Category | Rows | Classification breakdown |
|---|---:|---|
| MCP examples | 16 | 2 implemented active default; 3 implemented non-default/setup-library; 11 intentionally excluded; 0 missing and should be ported/curated |
| CLI tools/scripts/commands | 5 | 1 renamed/equivalent; 1 implemented non-default/setup-library; 1 intentionally excluded; 2 obsolete because LazyAI runtime replaced it |
| Adapter command assets | 8 | 8 implemented active default |
| Agents | 7 | 2 implemented active default; 2 renamed/equivalent; 2 intentionally excluded; 1 missing and should be ported/curated |
| Skills | 22 | 3 implemented active default; 1 implemented non-default/setup-library; 10 renamed/equivalent; 2 intentionally excluded; 3 missing and should be ported/curated; 3 obsolete because LazyAI runtime replaced it |
| Hooks/policies | 4 | 2 implemented non-default/setup-library; 1 renamed/equivalent; 1 missing and should be ported/curated |
| Rules/standards/protocols | 12 | 3 implemented non-default/setup-library; 7 renamed/equivalent; 2 intentionally excluded |
| Workflows | 11 | 7 implemented non-default/setup-library; 3 renamed/equivalent; 1 missing and should be ported/curated |
| Templates/presets | 15 | 5 implemented active default; 1 implemented non-default/setup-library; 6 renamed/equivalent; 1 intentionally excluded; 2 missing and should be ported/curated |

## Classification legend

### vibe-lab baseline item classifications

| Classification | Meaning |
|---|---|
| `implemented active default` | LazyAI actively emits or ships the equivalent item by default. |
| `implemented non-default/setup-library` | LazyAI contains the item or equivalent as setup-library/docs/manual material, but not active default adapter output. |
| `renamed/equivalent` | LazyAI item fulfills the same role under a different name or different asset kind. |
| `intentionally excluded` | Item is outside LazyAI product scope or conflicts with product boundaries. |
| `missing and should be ported/curated` | No LazyAI equivalent was observed and the item appears valuable enough for a follow-up decision. |
| `obsolete because LazyAI runtime replaced it` | LazyAI Go runtime/CLI supersedes the vibe-lab script/workflow/pattern. |

### LazyAI-only classifications

| Classification | Meaning |
|---|---|
| `LazyAI runtime-specific` | Exists because LazyAI is a runtime/product, not only a setup baseline. |
| `setup-core extension` | Extends setup creation, validation, compilation, adapter output, or update behavior. |
| `compatibility/historical material` | Retained for compatibility or archived history, not active/default behavior. |
| `candidate for removal/archive` | Appears unnecessary, stale, or outside current boundaries. |

Adapter emission statuses used below:

- `canonical default`: emitted from canonical/default library paths by supported adapters.
- `adapter-specific command`: emitted from `claudecode/commands` or `opencode/commands`.
- `setup-library only`: present as setup guidance/library material but not emitted as a default adapter asset.
- `CLI runtime`: provided by the shipped `lazyai-cli` binary.
- `repo harness`: repository maintenance script, not shipped CLI surface.
- `not emitted`: present or absent but not emitted by current adapter mapping.
- `N/A`: classification does not map to adapter output.

## MCP parity matrix

Source for vibe-lab examples: `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md`.
LazyAI source: `packages/cli/library/mcp/catalog.json`.

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| MCP | Context7 | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `implemented non-default/setup-library` | `packages/cli/library/mcp/catalog.json` server `context7`, enabled false | setup-library only | Context7 is cataloged as a disabled remote MCP entry using the verified remote endpoint and API-key header shape. | keep opt-in |
| MCP | memory | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `implemented active default` | `packages/cli/library/mcp/catalog.json` server `memory`, enabled true | canonical default | LazyAI enabled memory MCP matches vibe-lab Context category. | keep |
| MCP | filesystem | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `implemented active default` | `packages/cli/library/mcp/catalog.json` server `filesystem`, enabled true | canonical default | LazyAI enabled filesystem MCP matches vibe-lab Context category. | keep |
| MCP | linter | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | LazyAI uses project tests/lint commands as verification, not a cataloged MCP server. | exclude |
| MCP | test runner | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | Test execution is repo command behavior, not a default MCP dependency. | exclude |
| MCP | coverage | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | Coverage is project-specific verification, not current LazyAI setup-core MCP surface. | exclude |
| MCP | Docker | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | DevOps MCPs are not active LazyAI setup defaults. | exclude |
| MCP | Kubernetes | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | DevOps MCPs are not active LazyAI setup defaults. | exclude |
| MCP | Terraform | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | DevOps MCPs are not active LazyAI setup defaults. | exclude |
| MCP | PostgreSQL | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | Data MCPs are project-specific and not current LazyAI defaults. | exclude |
| MCP | SQLite | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | LazyAI internal SQLite packages, not MCP catalog | CLI runtime | LazyAI owns local SQLite state internally; it does not expose SQLite as default MCP. | exclude |
| MCP | Redis | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | Data MCPs are project-specific and not current LazyAI defaults. | exclude |
| MCP | GitHub | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `implemented non-default/setup-library` | `packages/cli/library/mcp/catalog.json` server `github`, enabled false; CLI tool `gh`, enabled true | setup-library only | The opt-in GitHub MCP entry supplements existing `gh` CLI support; it does not replace CLI-first workflows. | keep opt-in plus `gh` |
| MCP | Figma | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | Gate A closed Figma as a documented exclusion because no exact supported server shape was verified. | exclude until exact shape is approved |
| MCP | Playwright | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `implemented non-default/setup-library` | `packages/cli/library/mcp/catalog.json` server `playwright`, enabled false; CLI tool `playwright`, enabled false | setup-library only | LazyAI has optional disabled Playwright support rather than a default server. | document |
| MCP | Slack | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `intentionally excluded` | `None observed` | not emitted | Gate A closed Slack as a documented exclusion because no exact supported server shape was verified. | exclude until exact shape is approved |

LazyAI-only enabled MCP/catalog entries: `ripgrep`, `memoria`, `codegraph`, `qmd`, `graphify`, and `obsidian` are covered in the LazyAI-only section.

## CLI tools, scripts, and commands matrix

Source for shipped LazyAI command boundaries: `docs/concepts/product-boundaries.md`.
Source for vibe-lab scripts: `/Users/ricardo/code/vibe-lab/bin/`.

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| CLI/script/command | `bootstrap-project` | `/Users/ricardo/code/vibe-lab/bin/bootstrap-project` | `obsolete because LazyAI runtime replaced it` | `packages/cli/cmd/create.go`; `packages/cli/cmd/init.go`; `packages/cli/cmd/setup.go` | CLI runtime | LazyAI setup/bootstrap behavior lives in shipped Go commands instead of a repo script. | keep LazyAI CLI path |
| CLI/script/command | `doctor` | `/Users/ricardo/code/vibe-lab/bin/doctor` | `renamed/equivalent` | `packages/cli/cmd/doctor.go`; repo harness `bin/doctor` | CLI runtime plus repo harness | LazyAI has a shipped `doctor` command; root `bin/doctor` is dev-harness per product boundaries. | keep |
| CLI/script/command | `inject` | `/Users/ricardo/code/vibe-lab/bin/inject` | `obsolete because LazyAI runtime replaced it` | `packages/cli/cmd/compile.go`; `packages/cli/internal/adapter/output_mapping.go`; repo harness `bin/inject` | CLI runtime plus repo harness | LazyAI compiles adapter output through Go CLI/compiler; repo script remains maintainer harness. | keep LazyAI CLI path |
| CLI/script/command | `inject.original` | `/Users/ricardo/code/vibe-lab/bin/inject.original` | `intentionally excluded` | `None observed` | not emitted | Historical source script is not a shipped LazyAI command surface. | exclude |
| CLI/script/command | `startup-self-heal` | `/Users/ricardo/code/vibe-lab/bin/startup-self-heal` | `implemented non-default/setup-library` | repo harness `bin/startup-self-heal`; `packages/cli/library/canonical/hooks/session-start.md` | repo harness | LazyAI has repository self-heal script and session-start guidance, but not a shipped product command. | document |

## Adapter command assets matrix

Source for LazyAI adapter destinations: `packages/cli/internal/adapter/output_mapping.go`.
Claude Code command source: `packages/cli/library/claudecode/commands/` → `.claude/commands/`.
OpenCode command source: `packages/cli/library/opencode/commands/` → `.opencode/commands/`.
Copilot has no slash command surface.

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| Adapter command | `speckit.analyze.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.analyze.md` | `implemented active default` | `packages/cli/library/claudecode/commands/speckit.analyze.md`; `packages/cli/library/opencode/commands/speckit.analyze.md` | adapter-specific command | Same command name emitted for Claude Code and OpenCode. | keep |
| Adapter command | `speckit.checklist.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.checklist.md` | `implemented active default` | `packages/cli/library/claudecode/commands/speckit.checklist.md`; `packages/cli/library/opencode/commands/speckit.checklist.md` | adapter-specific command | Same command name emitted for Claude Code and OpenCode. | keep |
| Adapter command | `speckit.clarify.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.clarify.md` | `implemented active default` | `packages/cli/library/claudecode/commands/speckit.clarify.md`; `packages/cli/library/opencode/commands/speckit.clarify.md` | adapter-specific command | Same command name emitted for Claude Code and OpenCode. | keep |
| Adapter command | `speckit.constitution.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.constitution.md` | `implemented active default` | `packages/cli/library/claudecode/commands/speckit.constitution.md`; `packages/cli/library/opencode/commands/speckit.constitution.md` | adapter-specific command | Same command name emitted for Claude Code and OpenCode. | keep |
| Adapter command | `speckit.implement.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.implement.md` | `implemented active default` | `packages/cli/library/claudecode/commands/speckit.implement.md`; `packages/cli/library/opencode/commands/speckit.implement.md` | adapter-specific command | Same command name emitted for Claude Code and OpenCode. | keep |
| Adapter command | `speckit.plan.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.plan.md` | `implemented active default` | `packages/cli/library/claudecode/commands/speckit.plan.md`; `packages/cli/library/opencode/commands/speckit.plan.md` | adapter-specific command | Same command name emitted for Claude Code and OpenCode. | keep |
| Adapter command | `speckit.specify.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.specify.md` | `implemented active default` | `packages/cli/library/claudecode/commands/speckit.specify.md`; `packages/cli/library/opencode/commands/speckit.specify.md` | adapter-specific command | Same command name emitted for Claude Code and OpenCode. | keep |
| Adapter command | `speckit.tasks.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/commands/speckit.tasks.md` | `implemented active default` | `packages/cli/library/claudecode/commands/speckit.tasks.md`; `packages/cli/library/opencode/commands/speckit.tasks.md` | adapter-specific command | Same command name emitted for Claude Code and OpenCode. | keep |

LazyAI also has `packages/cli/library/claudecode/commands/{init,review,test,commit}.md` and `packages/cli/library/opencode/commands/{init,review,test,commit}.md`; these are LazyAI setup-core adapter extensions, not vibe-lab Speckit baseline gaps.

`packages/cli/library/canonical/commands/{graphify,handoff}.md` are provenance-covered canonical command assets, but `output_mapping.go` currently maps command emission from tool-specific command directories, not `canonical/commands`. That is a documented boundary, not changed here.

## Agents matrix

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| Agent | `planner` | `/Users/ricardo/code/vibe-lab/.agents/agents/planner.md` | `implemented active default` | `packages/cli/library/canonical/agents/planner.md` | canonical default | Same role survives as compressed neutral planning agent emitted to supported adapters. | keep |
| Agent | `reviewer` | `/Users/ricardo/code/vibe-lab/.agents/agents/reviewer.md` | `implemented active default` | `packages/cli/library/canonical/agents/reviewer.md` | canonical default | Same role survives as compressed neutral review agent emitted to supported adapters. | keep |
| Agent | `implementer` | `/Users/ricardo/code/vibe-lab/.agents/agents/implementer.md` | `renamed/equivalent` | `packages/cli/library/canonical/agents/builder.md`; archived historical `archive/issue-244-historical-library/packages/cli/library/agents/implementor.md` | canonical default | vibe-lab implementer builds from specs and verifies; LazyAI builder implements approved changes and verifies behavior. | keep renamed LazyAI name |
| Agent | `researcher` | `/Users/ricardo/code/vibe-lab/.agents/agents/researcher.md` | `renamed/equivalent` | `packages/cli/library/canonical/agents/scout.md` | canonical default | vibe-lab researcher gathers evidence before code; LazyAI scout gathers grounded code/doc facts before planning or implementation. | keep renamed LazyAI name |
| Agent | `deployer` | `/Users/ricardo/code/vibe-lab/.agents/agents/deployer.md` | `intentionally excluded` | `None observed` | not emitted | Deployment and CI/CD operations are not active LazyAI setup-core adapter defaults. | exclude from defaults |
| Agent | `evidence-verifier` | `/Users/ricardo/code/vibe-lab/.agents/agents/evidence-verifier.md` | `missing and should be ported/curated` | `None observed` | not emitted | LazyAI reviewer/test-first assets verify changes, but no exact evidence-claim verifier agent exists. | follow-up candidate |
| Agent | `responder` | `/Users/ricardo/code/vibe-lab/.agents/agents/responder.md` | `intentionally excluded` | `None observed` | not emitted | SRE incident response is outside LazyAI active setup defaults. | exclude from defaults |

## Skills matrix

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| Skill | `adhd-engineer` | `/Users/ricardo/code/vibe-lab/.agents/skills/adhd-engineer/SKILL.md` | `intentionally excluded` | `None observed` | not emitted | Personal workflow optimization is not part of current LazyAI setup-core defaults. | exclude |
| Skill | `architecture-review` | `/Users/ricardo/code/vibe-lab/.agents/skills/architecture-review/SKILL.md` | `renamed/equivalent` | `packages/cli/library/skills/improve-codebase-architecture.md`; `packages/cli/library/templates/adr.md` | setup-library only | LazyAI keeps architecture improvement/ADR guidance as setup-library material. | document |
| Skill | `caveman` | `/Users/ricardo/code/vibe-lab/.agents/skills/caveman/SKILL.md` | `implemented non-default/setup-library` | `packages/cli/library/prompts/compact.md`; `packages/cli/library/skills/update-memory.md` | setup-library only | LazyAI has compacting/memory-adjacent prompts but no default caveman skill. | document |
| Skill | `codebase-exploration` | `/Users/ricardo/code/vibe-lab/.agents/skills/codebase-exploration/SKILL.md` | `implemented active default` | `packages/cli/library/canonical/skills/codebase-exploration.md` | canonical default | Same skill is compressed into canonical adapter output. | keep |
| Skill | `create-agent` | `/Users/ricardo/code/vibe-lab/.agents/skills/create-agent/SKILL.md` | `obsolete because LazyAI runtime replaced it` | `packages/cli/cmd/create.go`; `packages/cli/internal/generator/agent.go` | CLI runtime | LazyAI creates agents through the shipped `create` command and generator. | keep CLI path |
| Skill | `create-hook` | `/Users/ricardo/code/vibe-lab/.agents/skills/create-hook/SKILL.md` | `missing and should be ported/curated` | `None observed` as a create generator type | not emitted | LazyAI `create` supports agent, skill, prompt, command, and template; no hook generator was observed. | follow-up candidate |
| Skill | `create-skill` | `/Users/ricardo/code/vibe-lab/.agents/skills/create-skill/SKILL.md` | `obsolete because LazyAI runtime replaced it` | `packages/cli/cmd/create.go`; `packages/cli/internal/generator/skill.go` | CLI runtime | LazyAI creates skills through the shipped `create` command and generator. | keep CLI path |
| Skill | `create-workflow` | `/Users/ricardo/code/vibe-lab/.agents/skills/create-workflow/SKILL.md` | `intentionally excluded` | `None observed` as a create generator type | not emitted | Workflow command surfaces are retired from active LazyAI defaults. | exclude |
| Skill | `diagnose` | `/Users/ricardo/code/vibe-lab/.agents/skills/diagnose/SKILL.md` | `implemented active default` | `packages/cli/library/canonical/skills/diagnose.md` | canonical default | Same debugging skill is compressed into canonical adapter output. | keep |
| Skill | `doc-backed-clarify` | `/Users/ricardo/code/vibe-lab/.agents/skills/doc-backed-clarify/SKILL.md` | `renamed/equivalent` | `packages/cli/library/skills/speckit-clarify.md`; `packages/cli/library/claudecode/commands/speckit.clarify.md`; `packages/cli/library/opencode/commands/speckit.clarify.md` | setup-library only plus adapter-specific command | LazyAI keeps clarification as Speckit skill/command flow. | document |
| Skill | `fast-feedback` | `/Users/ricardo/code/vibe-lab/.agents/skills/fast-feedback/SKILL.md` | `renamed/equivalent` | `packages/cli/library/skills/chain-verify.md`; `packages/cli/library/skills/iterate.md` | setup-library only | LazyAI keeps verification loop guidance under chain/iteration naming. | document |
| Skill | `four-point-vibe-coding` | `/Users/ricardo/code/vibe-lab/.agents/skills/four-point-vibe-coding/SKILL.md` | `renamed/equivalent` | `packages/cli/library/canonical/agents/primary-agent.md`; `packages/cli/library/root/AGENTS.template.md`; `packages/cli/library/root/copilot-instructions.template.md` | canonical default | LazyAI root/primary-agent guidance keeps four-point intake without vibe-lab branding. | keep |
| Skill | `handoff` | `/Users/ricardo/code/vibe-lab/.agents/skills/handoff/SKILL.md` | `renamed/equivalent` | `packages/cli/library/canonical/commands/handoff.md`; `packages/cli/internal/handoff` | CLI runtime plus canonical command asset | LazyAI has runtime handoff support and a canonical handoff command asset, not a default skill. | document |
| Skill | `issue-triage` | `/Users/ricardo/code/vibe-lab/.agents/skills/issue-triage/SKILL.md` | `missing and should be ported/curated` | `None observed` | not emitted | No issue-triage skill or generator output was observed in LazyAI library. | follow-up candidate |
| Skill | `memory-promotion` | `/Users/ricardo/code/vibe-lab/.agents/skills/memory-promotion/SKILL.md` | `renamed/equivalent` | `packages/cli/library/skills/memory-write.md`; `packages/cli/library/skills/update-memory.md`; `packages/cli/library/skills/impact-check.md` | setup-library only | LazyAI keeps memory writing/update guidance as setup-library skills. | document |
| Skill | `no-workarounds` | `/Users/ricardo/code/vibe-lab/.agents/skills/no-workarounds/SKILL.md` | `renamed/equivalent` | `packages/cli/library/skills/anti-speculation.md`; `packages/cli/library/rules/code-style.md`; `packages/cli/library/rules/testing.md` | setup-library only | LazyAI expresses the same discipline through anti-speculation and testing/code rules. | document |
| Skill | `project-guardrails-init` | `/Users/ricardo/code/vibe-lab/.agents/skills/project-guardrails-init/SKILL.md` | `renamed/equivalent` | `packages/cli/cmd/init.go`; `packages/cli/library/root/AGENTS.template.md`; `packages/cli/library/rules/` | CLI runtime plus setup-library | LazyAI initializes project guardrails through `init` and embedded rules/templates. | keep |
| Skill | `skill-authoring` | `/Users/ricardo/code/vibe-lab/.agents/skills/skill-authoring/SKILL.md` | `obsolete because LazyAI runtime replaced it` | `packages/cli/cmd/create.go`; `packages/cli/internal/generator/skill.go`; `packages/cli/cmd/validate.go` | CLI runtime | Skill authoring is generator/validation behavior in LazyAI. | keep CLI path |
| Skill | `task-to-issues` | `/Users/ricardo/code/vibe-lab/.agents/skills/task-to-issues/SKILL.md` | `missing and should be ported/curated` | `None observed` | not emitted | No task-to-issue workflow asset was observed in LazyAI library. | follow-up candidate |
| Skill | `tdd-planning` | `/Users/ricardo/code/vibe-lab/.agents/skills/tdd-planning/SKILL.md` | `renamed/equivalent` | `packages/cli/library/canonical/skills/test-first-change.md`; `packages/cli/library/skills/tdd-loop.md` | canonical default plus setup-library | LazyAI splits TDD into canonical test-first skill and deeper TDD loop guidance. | keep |
| Skill | `test-first-change` | `/Users/ricardo/code/vibe-lab/.agents/skills/test-first-change/SKILL.md` | `implemented active default` | `packages/cli/library/canonical/skills/test-first-change.md` | canonical default | Same skill is compressed into canonical adapter output. | keep |
| Skill | `zoom-out` | `/Users/ricardo/code/vibe-lab/.agents/skills/zoom-out/SKILL.md` | `renamed/equivalent` | `packages/cli/library/skills/improve-codebase-architecture.md`; `packages/cli/library/skills/spike.md`; `packages/cli/library/skills/plan.md` | setup-library only | LazyAI covers broader design re-evaluation through planning, spike, and architecture-improvement skills. | document |

## Hooks and policies matrix

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| Hook/policy | `block-destructive-shell` | `/Users/ricardo/code/vibe-lab/.agents/hooks/block-destructive-shell/POLICY.md` | `missing and should be ported/curated` | `packages/cli/library/rules/tool-use.md`; `packages/cli/library/rules/security.md` are partial only | not emitted | LazyAI has safety rules but no same-purpose hook policy asset. | follow-up candidate |
| Hook/policy | `caveman-memory-promotion` | `/Users/ricardo/code/vibe-lab/.agents/hooks/caveman-memory-promotion/POLICY.md` | `implemented non-default/setup-library` | `packages/cli/library/skills/memory-write.md`; `packages/cli/library/skills/update-memory.md`; `packages/cli/library/prompts/compact.md` | setup-library only | LazyAI has memory/compact guidance but no hook automation. | document |
| Hook/policy | `objective-workflow-gate` | `/Users/ricardo/code/vibe-lab/.agents/hooks/objective-workflow-gate/POLICY.md` | `renamed/equivalent` | `packages/cli/library/hooks/rpi-gate-check.yml`; `packages/cli/library/skills/rpi.md` | setup-library only | LazyAI expresses workflow gating through RPI gate policy and skill guidance. | keep |
| Hook/policy | `startup-self-heal` | `/Users/ricardo/code/vibe-lab/.agents/hooks/startup-self-heal/POLICY.md` | `implemented non-default/setup-library` | `bin/startup-self-heal`; `packages/cli/library/canonical/hooks/session-start.md` | repo harness plus canonical hook asset | LazyAI keeps startup/session guardrails, but not as a default adapter-emitted hook policy. | document |

## Rules, standards, and protocols matrix

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| Rule/protocol | `artifact-rules.md` | `/Users/ricardo/code/vibe-lab/canonical/artifact-rules.md` | `implemented non-default/setup-library` | `packages/cli/library/rules/tool-use.md`; `packages/cli/library/rules/workflow.md`; `packages/cli/library/templates/` | setup-library only | LazyAI keeps artifact/tool/workflow constraints as project rules and templates. | document |
| Rule/protocol | `caveman-ai-memory.md` | `/Users/ricardo/code/vibe-lab/canonical/caveman-ai-memory.md` | `renamed/equivalent` | `packages/cli/library/prompts/compact.md`; `packages/cli/library/skills/memory-write.md`; `packages/cli/library/skills/update-memory.md` | setup-library only | LazyAI splits memory compaction and memory updates across prompt/skill assets. | document |
| Rule/protocol | `clarification-levels.md` | `/Users/ricardo/code/vibe-lab/canonical/clarification-levels.md` | `renamed/equivalent` | `packages/cli/library/skills/speckit-clarify.md`; `packages/cli/library/claudecode/commands/speckit.clarify.md`; `packages/cli/library/opencode/commands/speckit.clarify.md` | setup-library only plus adapter-specific command | LazyAI maps clarification to Speckit clarify assets. | keep |
| Rule/protocol | `clean-code.md` | `/Users/ricardo/code/vibe-lab/canonical/clean-code.md` | `implemented non-default/setup-library` | `packages/cli/library/rules/code-style.md`; `packages/cli/library/rules/testing.md`; `packages/cli/library/rules/tool-use.md` | setup-library only | LazyAI has starter rules for code style, tests, and tool use. | keep |
| Rule/protocol | `constitution.md` | `/Users/ricardo/code/vibe-lab/canonical/constitution.md` | `renamed/equivalent` | `packages/cli/library/skills/speckit-constitution.md`; `packages/cli/library/claudecode/commands/speckit.constitution.md`; `packages/cli/library/opencode/commands/speckit.constitution.md` | setup-library only plus adapter-specific command | LazyAI supports constitution workflows through Speckit assets, not the exact static canonical doc. | document |
| Rule/protocol | `cookbook-recipe.md` | `/Users/ricardo/code/vibe-lab/canonical/cookbook-recipe.md` | `intentionally excluded` | `None observed` | not emitted | Cookbook authoring is not current LazyAI setup-core output. | exclude |
| Rule/protocol | `engineering-principles.md` | `/Users/ricardo/code/vibe-lab/canonical/engineering-principles.md` | `implemented non-default/setup-library` | `packages/cli/library/rules/code-style.md`; `packages/cli/library/rules/security.md`; `packages/cli/library/rules/testing.md`; `packages/cli/library/standards/starter/` | setup-library only | LazyAI distributes engineering principles as rules and starter standards. | keep |
| Rule/protocol | `llm-engineering-relationships.md` | `/Users/ricardo/code/vibe-lab/canonical/llm-engineering-relationships.md` | `intentionally excluded` | `None observed` | not emitted | Relationship/knowledge-base guidance is not current setup-core output. | exclude |
| Rule/protocol | `mcp-setup.md` | `/Users/ricardo/code/vibe-lab/canonical/mcp-setup.md` | `renamed/equivalent` | `packages/cli/library/mcp/catalog.json`; `packages/cli/cmd/server.go`; `packages/cli/cmd/compile.go` | CLI runtime plus setup-library | LazyAI uses a structured catalog and CLI compilation instead of a standalone guide only. | keep |
| Rule/protocol | `prd-plan-todo.md` | `/Users/ricardo/code/vibe-lab/canonical/prd-plan-todo.md` | `renamed/equivalent` | `packages/cli/library/templates/spec-template.md`; `packages/cli/library/templates/plan-template.md`; `packages/cli/library/templates/tasks-template.md`; `packages/cli/library/specs-agents/` | canonical default plus setup-library | LazyAI maps product planning into spec/plan/tasks templates and specs-agent guides. | keep |
| Rule/protocol | `self-assessment.md` | `/Users/ricardo/code/vibe-lab/canonical/self-assessment.md` | `renamed/equivalent` | `packages/cli/library/skills/review.md`; `packages/cli/library/skills/process-audit.md`; `packages/cli/library/templates/audit-template.md` | setup-library only | LazyAI keeps assessment as review/process audit guidance. | document |
| Rule/protocol | `tdd-planning.md` | `/Users/ricardo/code/vibe-lab/canonical/tdd-planning.md` | `renamed/equivalent` | `packages/cli/library/canonical/skills/test-first-change.md`; `packages/cli/library/skills/tdd-loop.md`; `packages/cli/library/rules/testing.md` | canonical default plus setup-library | LazyAI splits TDD planning across canonical skill, detailed skill, and testing rule. | keep |

## Workflows matrix

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| Workflow | `adversarial-review.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/adversarial-review.md` | `implemented non-default/setup-library` | `packages/cli/library/skills/red-team-plan.md`; `packages/cli/library/templates/audit-template.md` | setup-library only | LazyAI has adversarial planning/audit guidance, not a runtime workflow command. | document |
| Workflow | `bugfix.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/bugfix.md` | `implemented non-default/setup-library` | `packages/cli/library/specs-agents/bugfixes.md`; `packages/cli/library/skills/bugfix.md`; `packages/cli/library/templates/bugfix-rca-template.md` | setup-library only | LazyAI keeps bugfix workflow as specs-agent guide, skill, and template. | keep |
| Workflow | `code-review.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/code-review.md` | `implemented non-default/setup-library` | `packages/cli/library/canonical/agents/reviewer.md`; `packages/cli/library/skills/review.md`; `packages/cli/library/templates/code-review-template.md` | setup-library only plus canonical agent | LazyAI covers review through canonical reviewer and setup-library review assets. | keep |
| Workflow | `documentation.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/documentation.md` | `implemented non-default/setup-library` | `packages/cli/library/specs-agents/templates.md`; `packages/cli/library/specs-agents/standards.md`; `packages/cli/library/templates/standard.md` | setup-library only | LazyAI documentation workflow exists as specs-agent/template guidance. | keep |
| Workflow | `feature.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/feature.md` | `implemented non-default/setup-library` | `packages/cli/library/specs-agents/features.md`; `packages/cli/library/templates/spec-template.md`; `packages/cli/library/templates/plan-template.md`; `packages/cli/library/templates/tasks-template.md` | setup-library only plus canonical templates | LazyAI keeps feature workflow as specs-agent guides/templates. | keep |
| Workflow | `refactor.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/refactor.md` | `implemented non-default/setup-library` | `packages/cli/library/specs-agents/refactors.md`; `packages/cli/library/templates/adr.md`; `packages/cli/library/skills/improve-codebase-architecture.md` | setup-library only | LazyAI keeps refactor workflow as specs-agent guide, ADR template, and architecture skill. | keep |
| Workflow | `spike.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/spike.md` | `implemented non-default/setup-library` | `packages/cli/library/skills/spike.md`; `packages/cli/library/templates/spike-template.md`; `packages/cli/library/templates/poc-template.md` | setup-library only | LazyAI keeps spike/POC material without runtime workflow command surfaces. | keep |
| Workflow | `verified-research.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/verified-research.md` | `renamed/equivalent` | `packages/cli/library/skills/research.md`; `packages/cli/library/skills/chain-verify.md`; `packages/cli/library/canonical/agents/scout.md` | setup-library only plus canonical agent | LazyAI separates research and verification into scout/research/chain assets. | document |
| Workflow | `verified-research/templates/artifact-set.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/verified-research/templates/artifact-set.md` | `missing and should be ported/curated` | `None observed` | not emitted | No artifact-set template equivalent was observed. | follow-up candidate |
| Workflow | `verified-research/templates/checklist.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/verified-research/templates/checklist.md` | `renamed/equivalent` | `packages/cli/library/templates/checklist-template.md`; `packages/cli/library/skills/chain-verify.md` | canonical default plus setup-library | LazyAI has generic checklist and verification assets. | document |
| Workflow | `verified-research/templates/prompt-template.md` | `/Users/ricardo/code/vibe-lab/.agents/workflows/verified-research/templates/prompt-template.md` | `renamed/equivalent` | `packages/cli/library/prompts/local-example.md`; `packages/cli/library/prompts/local-examples/preflight-task-framing.md`; `packages/cli/library/prompts/local-examples/react-trace-and-handoff.md` | setup-library only | LazyAI has local prompt examples instead of this exact verified-research prompt template. | document |

## Templates and presets matrix

| Category | vibe-lab item | vibe-lab source path | LazyAI classification | LazyAI current/target path | Adapter emission status | Rationale | Recommendation |
|---|---|---|---|---|---|---|---|
| Template/preset | `agent-template.md` | `/Users/ricardo/code/vibe-lab/canonical/agent-template.md` | `renamed/equivalent` | `packages/cli/internal/generator/agent.go`; `packages/cli/cmd/create.go` | CLI runtime | LazyAI generates agent artifacts through code, not a static baseline template. | keep generator |
| Template/preset | `hook-template.md` | `/Users/ricardo/code/vibe-lab/canonical/hook-template.md` | `missing and should be ported/curated` | `None observed` as a generator type or library template | not emitted | LazyAI has hook assets but no hook authoring template/generator. | follow-up candidate |
| Template/preset | `learning-template.md` | `/Users/ricardo/code/vibe-lab/canonical/learning-template.md` | `intentionally excluded` | `None observed` | not emitted | Learning-note templates are not current LazyAI setup-core output. | exclude |
| Template/preset | `policy-template.md` | `/Users/ricardo/code/vibe-lab/canonical/policy-template.md` | `missing and should be ported/curated` | `None observed` as a generator type or library template | not emitted | LazyAI has RPI hook policy YAML but no generic policy template. | follow-up candidate |
| Template/preset | `skill-template.md` | `/Users/ricardo/code/vibe-lab/canonical/skill-template.md` | `renamed/equivalent` | `packages/cli/internal/generator/skill.go`; `packages/cli/cmd/create.go`; `packages/cli/library/skills/populate/SKILL.md` | CLI runtime plus setup-library | LazyAI generates skills through code and keeps a minimal Agent Skills-compatible seed. | keep generator |
| Template/preset | `workflow-template.md` | `/Users/ricardo/code/vibe-lab/canonical/workflow-template.md` | `renamed/equivalent` | `packages/cli/library/specs-agents/workflows.md`; `packages/cli/library/templates/task-harness-template.md` | setup-library only | LazyAI preserves workflow documentation/templates but not retired runtime workflow commands. | document |
| Template/preset | `preset.yml` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/preset.yml` | `implemented non-default/setup-library` | `packages/cli/library/claudecode/commands/speckit.*.md`; `packages/cli/library/opencode/commands/speckit.*.md`; `packages/cli/library/templates/*-template.md` | setup-library only plus adapter-specific command | LazyAI carries the preset contents as commands/templates, not the exact preset manifest file. | document |
| Template/preset | `checklist-template.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/checklist-template.md` | `implemented active default` | `packages/cli/library/templates/checklist-template.md` | canonical default | Same template name is emitted to adapter template surfaces. | keep |
| Template/preset | `constitution-template.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/constitution-template.md` | `missing and should be ported/curated` | `None observed` | not emitted | LazyAI has Speckit constitution command/skill but no constitution template file. | follow-up candidate |
| Template/preset | `plan-template.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/plan-template.md` | `implemented active default` | `packages/cli/library/templates/plan-template.md` | canonical default | Same template name is emitted to adapter template surfaces. | keep |
| Template/preset | `spec-template.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/spec-template.md` | `implemented active default` | `packages/cli/library/templates/spec-template.md` | canonical default | Same template name is emitted to adapter template surfaces. | keep |
| Template/preset | `tasks-template.md` | `/Users/ricardo/code/vibe-lab/canonical/speckit-vibe-lab-preset/templates/tasks-template.md` | `implemented active default` | `packages/cli/library/templates/tasks-template.md` | canonical default | Same template name is emitted to adapter template surfaces. | keep |
| Template/preset | `prd.md` | `/Users/ricardo/code/vibe-lab/specs/_templates/prd.md` | `renamed/equivalent` | `packages/cli/library/templates/spec-template.md`; `packages/cli/library/specs-agents/features.md` | canonical default plus setup-library | LazyAI RPI/spec flow replaces standalone PRD in many setup guides. | document |
| Template/preset | `tasks.md` | `/Users/ricardo/code/vibe-lab/specs/_templates/tasks.md` | `implemented active default` | `packages/cli/library/templates/tasks-template.md` | canonical default | LazyAI has active tasks template output. | keep |
| Template/preset | `techspec.md` | `/Users/ricardo/code/vibe-lab/specs/_templates/techspec.md` | `renamed/equivalent` | `packages/cli/library/templates/plan-template.md`; `packages/cli/library/templates/bugfix-rca-template.md`; `packages/cli/library/templates/tech-debt-template.md` | canonical default | LazyAI uses plan/RCA/tech-debt templates instead of one generic techspec file. | document |

## LazyAI-only active/default assets

This section covers active/default LazyAI assets that are not direct same-name vibe-lab baseline items. Setup-library-only inventories are intentionally not expanded into one row per file unless they affect active/default parity.

| Category | LazyAI item | LazyAI path | Adapter emission status | Classification | Rationale | Recommendation |
|---|---|---|---|---|---|---|
| Agent | `primary-agent` | `packages/cli/library/canonical/agents/primary-agent.md` | canonical default | `LazyAI runtime-specific` | Default LazyAI runtime entry point; provenance marks it LazyAI-authored and independent of retired runtime concepts. | Retain |
| Agent | `builder` | `packages/cli/library/canonical/agents/builder.md` | canonical default | `setup-core extension` | LazyAI-neutral name for implementation role replacing vibe-lab `implementer`. | Retain |
| Agent | `scout` | `packages/cli/library/canonical/agents/scout.md` | canonical default | `setup-core extension` | LazyAI-neutral name for research role replacing vibe-lab `researcher`. | Retain |
| Skill | `pr-review` | `packages/cli/library/canonical/skills/pr-review.md` | canonical default | `setup-core extension` | LazyAI emits a compact PR review skill; no same-name vibe-lab baseline skill was observed. | Retain |
| Hook | `pre-commit` | `packages/cli/library/canonical/hooks/pre-commit.md` | not emitted | `setup-core extension` | Documents LazyAI token-rent pre-commit responsibility; curation says adapter targets are none. | Document |
| Hook | `session-start` | `packages/cli/library/canonical/hooks/session-start.md` | not emitted | `setup-core extension` | Curated LazyAI session-start guardrail without vibe-lab runtime dependency. | Document |
| Command asset | `graphify` | `packages/cli/library/canonical/commands/graphify.md` | not emitted by `output_mapping.go` command mapping | `setup-core extension` | Provenance-covered command asset; command curation is excluded from current curation wave. | Human decision needed on emission/curation |
| Command asset | `handoff` | `packages/cli/library/canonical/commands/handoff.md` | not emitted by `output_mapping.go` command mapping | `setup-core extension` | Provenance-covered command asset with related runtime handoff package. | Human decision needed on emission/curation |
| Adapter command | `init` | `packages/cli/library/claudecode/commands/init.md`; `packages/cli/library/opencode/commands/init.md` | adapter-specific command | `setup-core extension` | LazyAI-specific setup command asset emitted to supported command surfaces. | Retain |
| Adapter command | `review` | `packages/cli/library/claudecode/commands/review.md`; `packages/cli/library/opencode/commands/review.md` | adapter-specific command | `setup-core extension` | LazyAI-specific review command asset emitted to supported command surfaces. | Retain |
| Adapter command | `test` | `packages/cli/library/claudecode/commands/test.md`; `packages/cli/library/opencode/commands/test.md` | adapter-specific command | `setup-core extension` | LazyAI-specific test command asset emitted to supported command surfaces. | Retain |
| Adapter command | `commit` | `packages/cli/library/claudecode/commands/commit.md`; `packages/cli/library/opencode/commands/commit.md` | adapter-specific command | `setup-core extension` | LazyAI-specific commit command asset emitted to supported command surfaces. | Retain |
| MCP | `ripgrep` | `packages/cli/library/mcp/catalog.json` server `ripgrep` | canonical default | `setup-core extension` | LazyAI includes fast code search in the catalog. | Retain |
| MCP | `memoria` | `packages/cli/library/mcp/catalog.json` server `memoria` | canonical default | `setup-core extension` | LazyAI includes git/history memory support. | Retain |
| MCP | `codegraph` | `packages/cli/library/mcp/catalog.json` server `codegraph` | canonical default | `setup-core extension` | LazyAI includes code knowledge graph support. | Retain |
| MCP | `qmd` | `packages/cli/library/mcp/catalog.json` server `qmd` | canonical default | `setup-core extension` | LazyAI includes markdown knowledge-base search support. | Retain |
| MCP | `graphify` | `packages/cli/library/mcp/catalog.json` server `graphify` | canonical default | `setup-core extension` | LazyAI includes graphify knowledge graph support. | Retain |
| MCP | `obsidian` | `packages/cli/library/mcp/catalog.json` server `obsidian` | canonical default | `setup-core extension` | LazyAI includes Obsidian vault support. | Retain |
| Adapter output family | Claude output styles | `packages/cli/library/claudecode/output-styles/explanatory.md`; `packages/cli/library/claudecode/output-styles/terse.md` | canonical default for Claude Code output-styles | `setup-core extension` | Output styles are LazyAI adapter support; not a vibe-lab baseline category in issue #245. | Document |
| Adapter output family | OpenCode modes | `packages/cli/library/opencode/modes/audit.md`; `packages/cli/library/opencode/modes/plan.md` | canonical default for OpenCode modes | `setup-core extension` | OpenCode chat modes are adapter support and not a vibe-lab baseline item. | Document |
| Adapter output family | Copilot chatmodes | `packages/cli/library/chatmodes/architect.chatmode.md`; `packages/cli/library/chatmodes/reviewer.chatmode.md` | canonical default for Copilot chatmodes | `setup-core extension` | Copilot chatmodes are adapter support and not a vibe-lab baseline item. | Document |
| Adapter output family | Copilot prompts | `packages/cli/library/prompts/init.md`; `packages/cli/library/prompts/plan.md`; `packages/cli/library/prompts/research.md`; `packages/cli/library/prompts/implement.md`; `packages/cli/library/prompts/compact.md`; `packages/cli/library/prompts/local-example.md` | canonical default for Copilot prompts | `setup-core extension` | Copilot prompt output is adapter support; Claude/OpenCode prompts ship as commands per output mapping notes. | Document |

LazyAI-only non-default setup-library skills, rules, standards, templates, and specs-agent guides are curated in `packages/cli/library/manifests/curation.yaml`. They are setup-core support unless listed as archived exclusions in that manifest. This report does not recommend removing them.

## Stale provenance and current-baseline path mismatches

Source: `packages/cli/library/manifests/provenance.yaml` compared to current `/Users/ricardo/code/vibe-lab` checkout at audit time. These were documentation findings in the original audit and were later addressed in issue `#251` by either correcting current baseline paths or reclassifying entries as `LazyAI-authored` where no current vibe-lab source exists.

| LazyAI path | Audit-time provenance source path | Audit-time baseline result | Resolution in `#251` |
|---|---|---|---|
| `packages/cli/library/canonical/agents/builder.md` | `.agents/agents/builder.md` | Missing; analogous current vibe-lab path is `/Users/ricardo/code/vibe-lab/.agents/agents/implementer.md` | Not applicable to the current manifest; canonical exact-baseline agent set uses `implementer.md` directly. |
| `packages/cli/library/canonical/agents/scout.md` | `.agents/agents/scout.md` | Missing; analogous current vibe-lab path is `/Users/ricardo/code/vibe-lab/.agents/agents/researcher.md` | Not applicable to the current manifest; canonical exact-baseline agent set uses `researcher.md` directly. |
| `packages/cli/library/canonical/commands/graphify.md` | `commands/graphify.md` | Missing in observed current baseline | Reclassified as `LazyAI-authored` command inventory asset. |
| `packages/cli/library/canonical/commands/handoff.md` | `commands/handoff.md` | Missing; current handoff exists at `/Users/ricardo/code/vibe-lab/.agents/skills/handoff/SKILL.md` | Repointed to `.agents/skills/handoff/SKILL.md`. |
| `packages/cli/library/canonical/hooks/session-start.md` | `hooks/session-start.md` | Missing in observed current baseline | Reclassified as `LazyAI-authored` setup-core extension. |
| `packages/cli/library/canonical/skills/codebase-exploration.md` | `skills/codebase-exploration/SKILL.md` | Missing; current path is `/Users/ricardo/code/vibe-lab/.agents/skills/codebase-exploration/SKILL.md` | Corrected to `.agents/skills/codebase-exploration/SKILL.md`. |
| `packages/cli/library/canonical/skills/diagnose.md` | `skills/diagnose/SKILL.md` | Missing; current path is `/Users/ricardo/code/vibe-lab/.agents/skills/diagnose/SKILL.md` | Corrected to `.agents/skills/diagnose/SKILL.md`. |
| `packages/cli/library/canonical/skills/pr-review.md` | `skills/pr-review/SKILL.md` | Missing in observed current baseline | Reclassified as `LazyAI-authored` setup-core extension. |
| `packages/cli/library/canonical/skills/test-first-change.md` | `skills/test-first-change/SKILL.md` | Missing; current path is `/Users/ricardo/code/vibe-lab/.agents/skills/test-first-change/SKILL.md` | Corrected to `.agents/skills/test-first-change/SKILL.md`. |

## Recommendations and follow-up candidates

### No action / retain

- Keep canonical `planner`, `reviewer`, `builder`, `scout`, and `primary-agent` as the active neutral agent set.
- Keep canonical skills `codebase-exploration`, `diagnose`, `test-first-change`, and LazyAI-only `pr-review`.
- Keep Speckit command assets in `packages/cli/library/claudecode/commands/` and `packages/cli/library/opencode/commands/`.
- Keep active templates `checklist-template.md`, `plan-template.md`, `spec-template.md`, and `tasks-template.md`.
- Keep repo scripts under `bin/` classified as dev-harness, not shipped product commands.

### Plan C follow-up status

Completed in `specs/issues/245-parity-audit/follow-up-plan-c.md`:

1. Curated `evidence-verifier` as a non-default setup-library skill.
2. Curated `block-destructive-shell` as a non-default hook/policy asset.
3. Added `constitution-template.md` and `verified-research-artifact-set-template.md`.
4. Added `hook` as a supported `lazyai-cli create` artifact type while keeping `workflow`, `domain`, and `mode` rejected.
5. Curated `issue-triage` and `task-to-issues` as non-default setup-library skills.
6. Added disabled opt-in remote MCP entries for Context7 and GitHub MCP.
7. Closed Figma and Slack as documented exclusions until exact supported server shapes are verified and separately approved.
8. Documented `graphify` and `handoff` as provenance-covered command assets that are not active command emission sources.

Remaining candidate:

- None for the audit-time provenance mismatch list; issue `#251` resolves the documented source-path and ownership corrections without changing runtime behavior.

### Explicit exclusions from active defaults

- Do not add `deployer` or `responder` to active canonical adapter defaults without a separate product decision.
- Do not reintroduce retired LazyAI `task`, `workflow`, `orchestration`, `mcp-setup`, obsolete `eval`, Fortnite, or orchestrator defaults.
- Do not treat vibe-lab `inject.original` as a product command.
- Do not add generic DevOps/Data MCP defaults without a project-specific opt-in model.

## Acceptance checklist

| Issue #245 criterion | Status | Evidence |
|---|---|---|
| A checked-in parity report exists under `specs/` or `docs/reports/` with a table for every category. | Done | This file: `specs/issues/245-parity-audit/parity-report.md`; sections cover MCP, CLI/scripts/commands, adapter command assets, agents, skills, hooks, rules/protocols, workflows, templates/presets, and LazyAI-only active/default assets. |
| The report names exact vibe-lab source paths and exact LazyAI target/current paths. | Done | Every baseline row includes `vibe-lab source path` and `LazyAI current/target path`. |
| Every vibe-lab baseline asset has a LazyAI classification and recommendation. | Done | Baseline matrices cover 100 rows across the issue categories, each with classification and recommendation. |
| Every LazyAI-only active/default asset has a rationale. | Done | See `LazyAI-only active/default assets`. |
| Active adapter output remains neutral LazyAI and does not reintroduce retired Fortnite/orchestrator runtime defaults. | Done | `packages/cli/internal/adapter/output_mapping.go` was not changed; Plan C tests lock disabled remote MCP opt-in behavior. |
| `packages/cli/library/manifests/provenance.yaml` and `curation.yaml` are updated for any resulting asset changes. | Done | `curation.yaml` covers the new non-default setup-library assets; `provenance.yaml` remains unchanged because no canonical-library history correction was approved. |
| `go test ./packages/cli/...` passes after any implementation changes. | Done | Passed after Plan C package changes. |
| `go run ./packages/cli/internal/tokenrent/cmd/token-rent-check` passes after any canonical library changes. | N/A | No canonical library file changed. |

## Feedback gate

⛔ Human gate: future follow-up issue creation or asset/manifest implementation outside the completed Plan C scope still requires explicit review and approval.
