# Mission: RPI implementation plan for LazyAI / vibe-lab missing improvements

You are running locally inside the LazyAI repository.

Repository context:

- Repo root: `/Users/ricardo/projects/teachable/lazyai`
- Product: `lazyai-cli`, a Go CLI that manages a canonical `.ai/` source tree and compiles it into tool-native AI-harness surfaces.
- LazyAI is **not** an agent runtime.
- vibe-lab is the embedded quality/process/taste layer: RPI workflow, human gates, anti-slop rules, skills, hooks, rubrics, handoff discipline, clean-code-for-agents, and trace/eval thinking.
- Host tools execute the generated assets: OpenCode, Claude Code, GitHub Copilot, Pi, OMP, Gemini/Antigravity, Kiro.
- LazyAI owns canonical source, schemas, manifests, validation, compilation, adapter output, doctor/status/update/eject, and optional runtime-adjacent state.
- LazyAI must not become a workflow/orchestration/runtime engine.

Your job is to use **RPI**:

```text
R = Research the current implementation carefully.
P = Plan the safest implementation sequence.
I = Implement the missing improvements in small, reviewable, test-backed changes.
```

This is a repo-local engineering task. Be evidence-based. Do not guess.

---

## Non-negotiable product boundaries

Do **not** implement or reintroduce:

```text
lazyai run workflow
lazyai autonomous task
lazyai orchestrate
lazyai background subagents
lazyai judge
mandatory eval engine
mandatory trace daemon
mandatory RAG core
LangChain/LangGraph/CrewAI dependency in core
old task/workflow/orchestration/eval command surfaces
Codex adapter
```

Allowed direction:

```text
- canonical .ai source improvements
- validation improvements
- adapter conformance fixtures
- golden tests
- docs/concepts consolidation
- hook lifecycle catalog
- trace/eval schemas and templates
- skill/agent quality validation
- MCP/catalog description improvements
- beta adapter documentation and capability mapping
- optional runtime-adjacent state documentation
```

The goal is to make LazyAI more trustworthy as a cross-tool harness asset compiler, not to turn it into an execution runtime.

---

## Safety rules

Before modifying anything:

1. Inspect the repo.
2. Confirm branch and dirty state.
3. Produce a short research summary.
4. Produce an implementation plan.
5. Only then start editing.

Never expose secret values.

Do not run destructive commands.

Never run:

```bash
rm -rf
git clean
git reset
git push
git commit
npm publish
go install
curl | sh
brew install
docker compose up
terraform apply
kubectl apply
lazyai update
lazyai eject
lazyai compile --write
lazyai init --force
```

Safe read-only commands:

```bash
pwd
git rev-parse --show-toplevel
git status --short
git branch --show-current
git log -1 --oneline
find . -maxdepth 3 -type f | sort
find . -maxdepth 4 -type d | sort
rg "TODO|FIXME|deprecated|retired|archived|not implemented|stub|coming soon" .
rg "task|workflow|orchestration|eval|judge|trace|rubric|skill|agent|hook|middleware|mcp|adapter|fixture|testdata" packages specs docs .
```

Safe tests, if available and local-only:

```bash
go test ./...
go vet ./...
```

If tests are expensive or fail due to unrelated environment issues, report exactly what happened.

---

# R — Research phase

First, inspect and summarize the current repo state.

## Required research areas

### 1. Compile contract

Inspect:

```text
packages/cli/internal/aimanifest/
packages/cli/internal/lockfile/
packages/cli/internal/writer/
packages/cli/internal/compiler/
packages/cli/cmd/compile.go
packages/cli/internal/schema/
```

Answer:

```text
- How compile selects targets
- How lockfile works
- How managed-region drift is detected
- How dry-run and force behave
- Where contract validation lives
- Where agent validation lives
```

### 2. Adapter support

Inspect:

```text
packages/cli/internal/adapter/
```

Confirm all 7 targets:

```text
opencode
claude
copilot
pi
omp
kiro
antigravity
```

For each adapter, identify:

```text
- output paths
- MCP behavior
- hook behavior
- skill behavior
- agent behavior
- tests
- beta/stable status
```

Special attention:

```text
- OMP beta
- Antigravity beta
- Pi MCP no-op
- Codex rejection
```

### 3. Test/conformance state

Inspect:

```text
packages/cli/testdata/
packages/cli/internal/adapter/*test*
packages/cli/internal/compiler/*test*
packages/cli/cmd/*test*
```

Answer:

```text
- Does packages/cli/testdata exist?
- Are adapter fixtures inline or file-based?
- Are golden outputs present?
- Which adapter behaviors are currently regression-tested?
- Which behaviors are missing fixture coverage?
```

### 4. Docs and spec status drift

Inspect:

```text
README.md
packages/cli/KNOWLEDGE_MAP.md
specs/029-lazyai-v2/spec.md
specs/adrs/
docs/concepts/
docs/lazyai-vibelab-product-spec-pack/
```

Answer:

```text
- Which feature/status source is authoritative?
- Is 029 marked Draft or Done?
- Are README, spec, ADRs, and knowledge map consistent?
- Are retired features still presented as active?
```

### 5. Embedded library assets

Inspect:

```text
packages/cli/library/
packages/cli/library/manifests/curation.yaml
packages/cli/library/manifests/provenance.yaml
packages/cli/internal/library/
```

Pay attention to:

```text
skills/
canonical/agents/
rules/
fragments/
prompts/
hooks/
templates/
standards/
mcp/catalog.json
opencode/
claudecode/
copilot/
pi/
omp/
antigravity/
```

Answer:

```text
- Which asset categories exist?
- How many active/excluded assets are declared?
- Which assets represent vibe-lab quality/process?
- Which assets are compiled to which surfaces?
```

### 6. Skills and agent validation

Inspect:

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/validate/
packages/cli/library/skills/
packages/cli/library/canonical/agents/
```

Answer:

```text
- What skill validation exists today?
- What agent validation exists today?
- Are triggers validated?
- Are non-triggers/misuse guidance validated?
- Are evidence requirements validated?
- Is progressive disclosure validated?
- Are missing references detected?
- Are broad/ambiguous skills warned?
```

### 7. Hooks / middleware / policy

Inspect:

```text
packages/cli/library/hooks/
packages/cli/library/claudecode/hooks/
packages/cli/library/opencode/
packages/cli/library/antigravity/
packages/cli/internal/adapter/
packages/cli/library/cupcake/
```

Answer:

```text
- What hook assets exist?
- Are hooks mapped to a neutral lifecycle?
- Which adapters support actual hooks?
- Which adapters are instruction-only?
- Is human approval modeled?
- Is destructive command blocking modeled?
```

### 8. Trace/eval/rubric state

Inspect:

```text
packages/cli/internal/evals/
packages/cli/internal/schema/
packages/cli/library/templates/
packages/cli/library/skills/
packages/cli/library/rules/
packages/cli/internal/runtime/
packages/cli/internal/db/
```

Answer:

```text
- Are eval cases supported?
- Are holdouts supported?
- Are rubrics supported?
- Is there a trace taxonomy?
- Is there a harness-change report template?
- Is there LLM-as-judge behavior?
- Are session/ledger/memory/metrics/cost optional?
```

---

# P — Plan phase

After research, produce a concise plan before implementation.

The plan must include:

```text
1. What you found
2. What is missing
3. What you will implement now
4. What you will intentionally not implement
5. Files likely to change
6. Tests to add/update
7. Risks
8. Validation commands
```

Use this priority order.

---

# Implementation priority order

## Priority 1 — Adapter conformance fixtures and golden tests

Problem:

```text
FR-008 expects conformance fixtures under packages/cli/testdata/, but packages/cli/testdata/ is missing or insufficient.
Some assertions exist inline in tests, but fixture-based coverage is incomplete.
```

Goal:

Create reusable test fixtures and golden output tests that prove the compiler contract across all 7 adapters.

Implement or improve:

```text
packages/cli/testdata/
  projects/
    minimal/
    full-seven-targets/
    opencode-only/
    claude-only/
    copilot-only/
    pi-only/
    omp-only/
    antigravity-only/
    kiro-only/
    drift-conflict/
    legacy-opencode-mcp/
    invalid-codex-target/
    beta-adapters/
  golden/
    opencode/
    claude/
    copilot/
    pi/
    omp/
    antigravity/
    kiro/
```

Required fixture scenarios:

```text
1. Minimal manifest compiles deterministically.
2. Full seven-target manifest emits expected native files.
3. OpenCode emits expected files and migrates legacy MCP shape if applicable.
4. Claude emits expected `.mcp.json` / settings behavior.
5. Copilot emits expected `.github/` and `.vscode/mcp.json` behavior.
6. Pi emits skills/root assets and deliberately does not emit MCP.
7. OMP emits `.omp/mcp.json` and is annotated/treated as beta.
8. Antigravity emits Gemini/Antigravity outputs and is annotated/treated as beta.
9. Kiro emits settings/MCP outputs and omits agents/skills by design if that is current behavior.
10. Codex target is rejected.
11. Drift conflict refuses overwrite unless force is used.
12. Dry-run writes nothing.
13. Managed-region writer preserves user-owned content.
14. Lockfile records source/output hashes.
```

Acceptance criteria:

```text
- `go test ./...` passes.
- Fixture tests are reusable and not only inline string assertions.
- Golden outputs are deterministic.
- Beta behavior is explicitly tested.
- Pi MCP no-op is explicitly tested as intentional.
- Codex rejection is explicitly tested.
```

## Priority 2 — Consolidate harness principles and product boundary docs

Problem:

```text
Harness principles exist across fragments, docs, ADRs, and assets, but may not exist as one canonical doctrine.
```

Goal:

Create a single source of truth for the LazyAI/vibe-lab architecture.

Add or update:

```text
docs/concepts/harness-principles.md
docs/concepts/trace-evidence-over-vibes.md
docs/concepts/no-runtime-orchestration.md
docs/concepts/lazyai-vibelab-boundary.md
```

If equivalent docs already exist, consolidate instead of duplicating.

Required content:

```text
LazyAI principles:
- canonical source first
- tool-native output
- compile before execute
- managed regions and lockfile evidence
- host tools execute
- no framework lock-in
- optional runtime-adjacent state
- no hidden orchestration
- no default RAG core
- human-gated quality
- trace evidence over vibes
- progressive disclosure
- adapter honesty over fake parity

vibe-lab principles:
- RPI workflow
- human gates
- anti-slop rules
- clean-code-for-agents
- skill/hook/rubric discipline
- handoff and memory hygiene
- evidence-verifier behavior
- trace/eval improvement thinking

Boundary:
- LazyAI owns `.ai/`, schemas, manifests, validation, compilation, adapter output.
- vibe-lab owns taste/process/quality assets.
- Host tools execute generated assets.
- Runtime-adjacent features remain optional.
```

Update cross-links from:

```text
README.md
specs/adrs/
specs/029-lazyai-v2/spec.md
packages/cli/KNOWLEDGE_MAP.md
```

Acceptance criteria:

```text
- One canonical doctrine exists.
- README points to it.
- ADRs/specs do not contradict it.
- No doc suggests LazyAI is a general runtime/orchestrator.
```

## Priority 3 — Fix product/spec status drift

Problem:

```text
packages/cli/KNOWLEDGE_MAP.md may say Done while specs/029-lazyai-v2/spec.md says Draft.
```

Goal:

Make status markers consistent.

Implement:

```text
- Identify all status markers for 029 LazyAI V2.
- Decide the correct status from evidence.
- Update inconsistent docs.
- Add a small test or validation check if a status convention exists.
```

Potential files:

```text
packages/cli/KNOWLEDGE_MAP.md
specs/029-lazyai-v2/spec.md
README.md
specs/adrs/
docs/lazyai-vibelab-product-spec-pack/
```

Acceptance criteria:

```text
- No contradictory Draft/Done status for the same feature.
- If the feature is partially complete, say so explicitly.
- Docs distinguish implemented, partial, beta, and planned.
```

## Priority 4 — Semantic skill validation

Problem:

```text
There are many skills, but structural validation is not enough.
Skills must be validated for quality, trigger clarity, misuse prevention, evidence requirements, and progressive disclosure.
```

Goal:

Improve `lazyai validate` / compile-time validation for skill quality.

Research current skill format before changing schemas.

Implement validation checks for:

```text
- missing or weak trigger guidance
- missing non-trigger / misuse guidance
- missing evidence requirements
- too-broad skill names or descriptions
- skill duplicates always-on rules
- skill includes too much always-loaded context
- missing required tool references
- missing fragment references
- unsupported adapter claims
- no progressive-disclosure structure where expected
```

Suggested validation severity:

```text
error:
  - invalid schema
  - broken references
  - unsupported adapter target
  - missing required fields

warning:
  - broad trigger
  - missing non-trigger guidance
  - missing evidence requirements
  - no examples/anti-examples
  - excessive token rent

info:
  - improvement suggestions
```

If the current skill format is Markdown-only, implement conservative linting rather than forcing a breaking schema.

Example quality expectations:

```markdown
# Skill quality contract

A high-quality skill should state:

- when to use it
- when not to use it
- required evidence
- expected output
- required tools, if any
- human gate, if any
- examples or anti-examples when ambiguous
- progressive-disclosure behavior
```

Potential files:

```text
packages/cli/internal/compiler/skill_validate.go
packages/cli/internal/validate/
packages/cli/internal/schema/
packages/cli/library/skills/
packages/cli/library/templates/skill-quality.md
```

Acceptance criteria:

```text
- Existing skills pass or produce actionable warnings.
- New tests cover good and bad skills.
- Validation output tells the user how to fix the skill.
- No mass breaking change unless necessary.
```

## Priority 5 — Agent role contract validation

Problem:

```text
Default agents exist, but agent validation should verify behavioral contracts, not only shape.
```

Goal:

Validate agent quality consistently.

For each agent, validate presence of:

```text
- role / purpose
- when to use
- when not to use
- expected workflow
- referenced skills
- evidence requirements
- human gates
- output format
- allowed/discouraged tools where applicable
- handoff behavior
```

Default agents to cover:

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

Suggested files:

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/schema/
packages/cli/library/canonical/agents/
packages/cli/library/templates/agent-contract.md
```

Acceptance criteria:

```text
- Default agents have explicit contracts.
- Tests cover missing role, missing evidence, missing handoff, bad references.
- Validation warnings are actionable.
```

## Priority 6 — Hook / middleware lifecycle catalog

Problem:

```text
Hook assets exist, but there may not be a neutral LazyAI hook lifecycle abstraction.
```

Goal:

Define a surface-neutral hook lifecycle and adapter capability mapping.

Add or update:

```text
packages/cli/library/hooks/catalog.md
packages/cli/library/hooks/schema.json
packages/cli/internal/hooks/lifecycle.go
packages/cli/internal/hooks/capabilities.go
```

Use this neutral lifecycle:

```text
before_agent
before_model
before_tool
after_tool
after_model
after_agent
on_error
on_compaction
on_handoff
on_human_gate
```

For each hook, record:

```yaml
name:
lifecycle:
purpose:
blocks_actions:
requires_human_approval:
captures_evidence:
surfaces:
  opencode: supported | partial | instruction_only | unsupported
  claude: supported | partial | instruction_only | unsupported
  copilot: supported | partial | instruction_only | unsupported
  pi: supported | partial | instruction_only | unsupported
  omp: supported | partial | instruction_only | unsupported
  antigravity: supported | partial | instruction_only | unsupported
  kiro: supported | partial | instruction_only | unsupported
```

Required hooks to classify:

```text
pre-commit
rpi-gate-check
caveman-memory-promotion
startup-self-heal
block-destructive-shell
objective-workflow-gate
```

Acceptance criteria:

```text
- Hook lifecycle is documented.
- Adapter capability is honest.
- Weak surfaces use instruction-only fallbacks instead of fake hooks.
- Validation catches unsupported hook claims.
- Tests cover at least one supported, partial, instruction-only, and unsupported mapping.
```

## Priority 7 — Trace/eval improvement loop without runtime expansion

Problem:

```text
Eval schema exists, but trace/eval workflow is schema-only and missing trace taxonomy plus harness-change workflow.
```

Goal:

Add file-based, inspectable trace/eval improvement assets without adding a scoring engine or runtime.

Add templates/docs:

```text
packages/cli/library/templates/trace-failure.md
packages/cli/library/templates/harness-change-report.md
packages/cli/library/templates/eval-promotion-checklist.md
packages/cli/library/templates/holdout-review.md
packages/cli/library/templates/evidence-report.md
docs/concepts/trace-eval-improvement-loop.md
```

Add trace taxonomy:

```yaml
context:
  - missing_project_context
  - stale_memory
  - bad_retrieval
  - ignored_user_constraint

tooling:
  - wrong_tool
  - missing_tool
  - unsafe_tool_use
  - unhandled_tool_error

workflow:
  - skipped_research
  - skipped_plan
  - skipped_tests
  - weak_handoff
  - missing_human_gate

quality:
  - hallucinated_api
  - overbroad_change
  - style_mismatch
  - incomplete_fix
  - no_evidence

adapter:
  - generated_output_drift
  - unsupported_surface_feature
  - broken_mcp_mapping
  - hook_not_emitted
```

Required workflow:

```text
trace/failure
→ tagged eval case
→ one targeted harness change
→ holdout check
→ human review
→ promote asset update
```

Acceptance criteria:

```text
- No LLM judge engine added.
- No scoring runtime added.
- No orchestration added.
- Templates are included in library manifest/curation if required.
- Existing eval schemas are reused or extended conservatively.
- Validation checks schema/shape only.
```

## Priority 8 — Rubric assets and validation

Problem:

```text
Rubric schema exists, but agent/skill/reviewer rubrics may not be first-class assets.
```

Goal:

Create reusable LazyAI/vibe-lab rubrics.

Add if missing:

```text
packages/cli/library/rubrics/
  reviewer.rubric.yaml
  evidence-verifier.rubric.yaml
  planner.rubric.yaml
  implementer.rubric.yaml
  skill-quality.rubric.yaml
  hook-quality.rubric.yaml
```

Or, if the project prefers templates instead of a new directory:

```text
packages/cli/library/templates/rubrics/
```

Rubrics should cover:

```text
correctness
evidence
test/lint/build verification
security
minimality
project convention fit
human gate behavior
handoff quality
uncertainty handling
```

Acceptance criteria:

```text
- Rubric assets are discoverable.
- Rubric schema validation passes.
- Reviewer and evidence-verifier agents reference appropriate rubrics if project conventions allow.
- No LLM-as-judge runtime is introduced.
```

## Priority 9 — MCP catalog examples and anti-examples

Problem:

```text
MCP catalog exists, but tool selection can improve through examples and anti-examples.
```

Goal:

Improve MCP/tool descriptions so generated agents select tools better.

Inspect:

```text
packages/cli/library/mcp/catalog.json
packages/cli/internal/adapter/mcp_compiler.go
packages/cli/internal/scaffold/mcp.go
```

For each canonical MCP server, add or validate:

```text
- purpose
- preferred use
- CLI-first vs MCP-first guidance
- examples
- anti-examples
- security notes
- adapter output differences
```

Canonical MCP servers:

```text
ai-memory
filesystem
ripgrep
codegraph
obsidian
```

Example expectation:

```json
{
  "name": "ripgrep",
  "purpose": "Fast deterministic code search",
  "preferred_use": "Use for exact code/text search before semantic exploration",
  "examples": [
    "Find all references to a function before editing it",
    "Search for TODOs in a package"
  ],
  "anti_examples": [
    "Do not use semantic codegraph when exact symbol search is enough",
    "Do not use MCP filesystem for bulk search when rg is available"
  ]
}
```

Acceptance criteria:

```text
- Catalog remains valid.
- Generated MCP output remains compatible.
- Tool descriptions improve without bloating always-loaded instructions.
- Tests updated if schema changes.
```

## Priority 10 — Adapter capability documentation and beta graduation path

Problem:

```text
OMP and Antigravity are beta.
Pi MCP is a deliberate no-op.
Kiro omits agents/skills by design.
Capability differences must be explicit.
```

Goal:

Document and test adapter capability truth.

Add/update:

```text
docs/adapters/capability-matrix.md
docs/adapters/opencode.md
docs/adapters/claude.md
docs/adapters/copilot.md
docs/adapters/pi.md
docs/adapters/omp.md
docs/adapters/antigravity.md
docs/adapters/kiro.md
```

Each adapter doc should include:

```text
- status: stable/beta/partial
- generated files
- supported asset types
- unsupported asset types
- MCP behavior
- hook behavior
- skill behavior
- agent behavior
- known limitations
- official-docs snapshot status, if applicable
- tests/fixtures proving behavior
```

Acceptance criteria:

```text
- Beta status is visible and justified.
- Pi MCP no-op is explicit and intentional.
- Kiro limitations are explicit.
- Capability matrix matches code.
- Tests fail if beta set unexpectedly changes.
```

## Priority 11 — Context compaction and handoff templates

Problem:

```text
Memory/handoff assets exist, but context compaction should be explicitly event-driven.
```

Goal:

Add or consolidate templates for context discipline.

Add/update:

```text
packages/cli/library/fragments/context-compaction.md
packages/cli/library/templates/context-handoff.md
packages/cli/library/templates/session-compaction.md
packages/cli/library/templates/recovery-handoff.md
```

Rules:

```text
Compact after:
- research phase
- planning phase
- implementation phase
- failed verification
- agent/tool handoff
- before context gets too large

Never compact away:
- user constraints
- decisions
- failed attempts
- commands run
- files touched
- unresolved risks
- human approval state
```

Acceptance criteria:

```text
- Handoff templates require evidence.
- Context compaction is framed as guidance/assets, not runtime automation.
- Existing memory/handoff skills reference these templates if appropriate.
```

## Priority 12 — Multi-agent decision guide without orchestration

Problem:

```text
Specialist agents and parallel-execution guidance exist, but the decision boundary should be explicit.
```

Goal:

Document when to use skills vs agents vs host-tool subagents vs parallel research.

Add:

```text
docs/concepts/multi-agent-boundary.md
packages/cli/library/templates/multi-agent-decision.md
packages/cli/library/fragments/specialist-agent-guidance.md
```

Required guidance:

```text
Use one agent + skills when:
- task is local
- context fits
- one person/agent can reason through it
- no isolation needed

Use specialist agents when:
- role separation improves quality
- review should be independent
- domain expertise differs

Use parallel agents when:
- work is read-heavy
- outputs are easy to merge
- boundaries are explicit

Avoid multi-agent when:
- agents need the same full context
- parallel writes would conflict
- merge cost exceeds benefit
- host tool cannot show evidence
```

Acceptance criteria:

```text
- No orchestration command added.
- No runtime scheduler added.
- Guidance is compiled as assets/templates where appropriate.
```

## Priority 13 — Official tool compliance matrix refresh

Problem:

```text
The product targets multiple host tools. Adapter behavior must stay aligned with official docs and current local assumptions.
```

Goal:

Refresh or validate the compliance matrix against current implemented behavior.

Inspect existing:

```text
docs/lazyai-vibelab-product-spec-pack/03_OFFICIAL_TOOL_COMPLIANCE_MATRIX*
docs/adapters/
specs/029-lazyai-v2/
packages/cli/internal/adapter/capabilities.go
packages/cli/internal/adapter/output_mapping.go
```

Update the matrix to include:

```text
OpenCode
Claude Code
GitHub Copilot
Pi
OMP
Gemini/Antigravity
Kiro
```

For each:

```text
- supported outputs
- unsupported outputs
- MCP config path
- hooks support
- commands/prompts/chatmodes support
- skill support
- agent support
- beta/stable status
- local test fixture
- known gaps
```

Do not invent official requirements if the repo does not include them. If web access is unavailable, mark “requires external docs refresh.”

Acceptance criteria:

```text
- Matrix matches code.
- Gaps are explicit.
- Beta adapters have documented graduation criteria.
```

## Priority 14 — Minimality / token-rent integration

Problem:

```text
internal/tokenrent and internal/minimality exist but are lightly documented.
```

Goal:

Make token-rent/minimality visible in validation and docs if already implemented.

Research:

```text
packages/cli/internal/tokenrent/
packages/cli/internal/minimality/
packages/cli/library/manifests/curation.yaml
```

Improve:

```text
- docs explaining token rent
- validation warning for large always-loaded assets
- curation manifest guidance
- minimality report docs
```

Acceptance criteria:

```text
- No breaking changes.
- Users understand why some assets are excluded or progressive-disclosure only.
- Validation output is actionable.
```

## Priority 15 — Init headless populate clarity

Problem:

```text
init headless populate is skipped because host AI tool is needed.
```

Goal:

Make this behavior explicit and safe.

Inspect:

```text
packages/cli/cmd/init.go
```

Improve docs/help text around:

```text
- why AGENTS.md placeholder fill may be skipped
- what host tool should do next
- how to complete setup manually
- how to validate after populate
```

Acceptance criteria:

```text
- No hidden runtime behavior added.
- CLI explains what happened and next steps.
- Tests updated if output/help changes.
```

## Priority 16 — Server L3 handshake clarity

Problem:

```text
server L3 handshake is not implemented in Go because it requires Node.js MCP SDK; only L1 config checks exist.
```

Goal:

Make server capability levels explicit.

Inspect:

```text
packages/cli/cmd/server.go
```

Improve:

```text
- docs
- help output
- doctor/status messaging
- tests if available
```

Acceptance criteria:

```text
- Users can distinguish L1 config check vs L3 handshake.
- No fake support is claimed.
- Missing L3 implementation is documented as intentional/current limitation.
```

---

# Implementation behavior requirements

## Keep changes small and reviewable

Prefer multiple logical commits/patch groups internally, but do not commit unless explicitly instructed.

Suggested patch groups:

```text
1. Conformance fixtures
2. Harness/boundary docs
3. Status drift cleanup
4. Skill/agent validation
5. Hook lifecycle catalog
6. Trace/eval templates
7. Rubrics
8. MCP examples
9. Adapter docs
10. Context/multi-agent docs
```

If the full scope is too large for one pass, implement in the priority order above and leave a clear remaining-work checklist.

## Use existing conventions

Before adding new directories, check existing patterns.

Prefer:

```text
- existing schema style
- existing test style
- existing adapter capability structs
- existing library manifest/curation flow
- existing docs naming
- existing command help style
```

Do not introduce a new architecture if existing internals already solve it.

## Do not break compatibility

Be careful with:

```text
.ai/lazyai.json
.ai/lock.json
adapter output paths
managed regions
existing generated files
curation manifest
MCP config shapes
beta adapter tests
```

If schema changes are needed, make them backward-compatible when possible.

---

# Required validation before final response

Run, if safe:

```bash
go test ./...
go vet ./...
```

Also run targeted tests for changed areas, for example:

```bash
go test ./packages/cli/internal/adapter/...
go test ./packages/cli/internal/compiler/...
go test ./packages/cli/internal/evals/...
go test ./packages/cli/cmd/...
```

If package paths differ, adapt them to the repo.

Check:

```bash
git status --short
```

Report all changed files.

---

# Final response format

When finished, produce the report format from `06_FINAL_REPORT_TEMPLATE.md`.

---

# Definition of done

This task is done when:

```text
- The repo has stronger fixture-based adapter conformance evidence.
- Harness/product boundary docs are consolidated.
- Spec/status drift is fixed.
- Skill/agent validation is more semantic and actionable.
- Hook lifecycle is documented and capability-mapped.
- Trace/eval improvement loop exists as templates/docs/schemas, not a runtime.
- Rubrics are first-class assets or templates.
- MCP catalog includes better examples/anti-examples.
- Adapter limitations and beta status are explicit.
- Context/handoff and multi-agent guidance are documented without orchestration.
- Tests pass or failures are clearly reported.
```

Remember:

```text
Be conservative.
Be evidence-based.
Improve trustworthiness.
Do not expand runtime scope.
Prefer fixtures, validation, docs, schemas, and adapter honesty over new execution machinery.
```
