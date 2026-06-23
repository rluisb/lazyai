# LazyAI / vibe-lab — Ordered Missing Improvements Backlog

This backlog contains all missing improvements identified from the LazyAI / vibe-lab comparison. Implement in this order unless repo evidence forces a safer split.

## Priority 1 — Adapter conformance fixtures and golden tests

### Problem

FR-008 expects conformance fixtures under `packages/cli/testdata/`, but `packages/cli/testdata/` is missing or insufficient. Some assertions exist inline in tests, but fixture-based coverage is incomplete.

### Goal

Create reusable test fixtures and golden output tests that prove the compiler contract across all 7 adapters.

### Implement or improve

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

### Required fixture scenarios

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

### Acceptance criteria

```text
- `go test ./...` passes.
- Fixture tests are reusable and not only inline string assertions.
- Golden outputs are deterministic.
- Beta behavior is explicitly tested.
- Pi MCP no-op is explicitly tested as intentional.
- Codex rejection is explicitly tested.
```

---

## Priority 2 — Consolidate harness principles and product boundary docs

### Problem

Harness principles exist across fragments, docs, ADRs, and assets, but may not exist as one canonical doctrine.

### Goal

Create a single source of truth for the LazyAI/vibe-lab architecture.

### Add or update

```text
docs/concepts/harness-principles.md
docs/concepts/trace-evidence-over-vibes.md
docs/concepts/no-runtime-orchestration.md
docs/concepts/lazyai-vibelab-boundary.md
```

If equivalent docs already exist, consolidate instead of duplicating.

### Required content

LazyAI principles:

```text
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
```

vibe-lab principles:

```text
- RPI workflow
- human gates
- anti-slop rules
- clean-code-for-agents
- skill/hook/rubric discipline
- handoff and memory hygiene
- evidence-verifier behavior
- trace/eval improvement thinking
```

Boundary:

```text
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

### Acceptance criteria

```text
- One canonical doctrine exists.
- README points to it.
- ADRs/specs do not contradict it.
- No doc suggests LazyAI is a general runtime/orchestrator.
```

---

## Priority 3 — Fix product/spec status drift

### Problem

`packages/cli/KNOWLEDGE_MAP.md` may say Done while `specs/029-lazyai-v2/spec.md` says Draft.

### Goal

Make status markers consistent.

### Implement

```text
- Identify all status markers for 029 LazyAI V2.
- Decide the correct status from evidence.
- Update inconsistent docs.
- Add a small test or validation check if a status convention exists.
```

### Potential files

```text
packages/cli/KNOWLEDGE_MAP.md
specs/029-lazyai-v2/spec.md
README.md
specs/adrs/
docs/lazyai-vibelab-product-spec-pack/
```

### Acceptance criteria

```text
- No contradictory Draft/Done status for the same feature.
- If the feature is partially complete, say so explicitly.
- Docs distinguish implemented, partial, beta, and planned.
```

---

## Priority 4 — Semantic skill validation

### Problem

There are many skills, but structural validation is not enough. Skills must be validated for quality, trigger clarity, misuse prevention, evidence requirements, and progressive disclosure.

### Goal

Improve `lazyai validate` / compile-time validation for skill quality.

### Research first

Inspect current skill format before changing schemas.

### Implement validation checks for

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

### Suggested severity

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

### Skill quality contract

A high-quality skill should state:

```text
- when to use it
- when not to use it
- required evidence
- expected output
- required tools, if any
- human gate, if any
- examples or anti-examples when ambiguous
- progressive-disclosure behavior
```

### Potential files

```text
packages/cli/internal/compiler/skill_validate.go
packages/cli/internal/validate/
packages/cli/internal/schema/
packages/cli/library/skills/
packages/cli/library/templates/skill-quality.md
```

### Acceptance criteria

```text
- Existing skills pass or produce actionable warnings.
- New tests cover good and bad skills.
- Validation output tells the user how to fix the skill.
- No mass breaking change unless necessary.
```

---

## Priority 5 — Agent role contract validation

### Problem

Default agents exist, but agent validation should verify behavioral contracts, not only shape.

### Goal

Validate agent quality consistently.

### For each agent, validate presence of

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

### Suggested files

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/schema/
packages/cli/library/canonical/agents/
packages/cli/library/templates/agent-contract.md
```

### Acceptance criteria

```text
- Default agents have explicit contracts.
- Tests cover missing role, missing evidence, missing handoff, bad references.
- Validation warnings are actionable.
```

---

## Priority 6 — Hook / middleware lifecycle catalog

### Problem

Hook assets exist, but there may not be a neutral LazyAI hook lifecycle abstraction.

### Goal

Define a surface-neutral hook lifecycle and adapter capability mapping.

### Add or update

```text
packages/cli/library/hooks/catalog.md
packages/cli/library/hooks/schema.json
packages/cli/internal/hooks/lifecycle.go
packages/cli/internal/hooks/capabilities.go
```

### Neutral lifecycle

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

### For each hook, record

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

### Required hooks to classify

```text
pre-commit
rpi-gate-check
caveman-memory-promotion
startup-self-heal
block-destructive-shell
objective-workflow-gate
```

### Acceptance criteria

```text
- Hook lifecycle is documented.
- Adapter capability is honest.
- Weak surfaces use instruction-only fallbacks instead of fake hooks.
- Validation catches unsupported hook claims.
- Tests cover at least one supported, partial, instruction-only, and unsupported mapping.
```

---

## Priority 7 — Trace/eval improvement loop without runtime expansion

### Problem

Eval schema exists, but trace/eval workflow is schema-only and missing trace taxonomy plus harness-change workflow.

### Goal

Add file-based, inspectable trace/eval improvement assets without adding a scoring engine or runtime.

### Add templates/docs

```text
packages/cli/library/templates/trace-failure.md
packages/cli/library/templates/harness-change-report.md
packages/cli/library/templates/eval-promotion-checklist.md
packages/cli/library/templates/holdout-review.md
packages/cli/library/templates/evidence-report.md
docs/concepts/trace-eval-improvement-loop.md
```

### Trace taxonomy

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

### Required workflow

```text
trace/failure
→ tagged eval case
→ one targeted harness change
→ holdout check
→ human review
→ promote asset update
```

### Acceptance criteria

```text
- No LLM judge engine added.
- No scoring runtime added.
- No orchestration added.
- Templates are included in library manifest/curation if required.
- Existing eval schemas are reused or extended conservatively.
- Validation checks schema/shape only.
```

---

## Priority 8 — Rubric assets and validation

### Problem

Rubric schema exists, but agent/skill/reviewer rubrics may not be first-class assets.

### Goal

Create reusable LazyAI/vibe-lab rubrics.

### Add if missing

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

### Rubrics should cover

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

### Acceptance criteria

```text
- Rubric assets are discoverable.
- Rubric schema validation passes.
- Reviewer and evidence-verifier agents reference appropriate rubrics if project conventions allow.
- No LLM-as-judge runtime is introduced.
```

---

## Priority 9 — MCP catalog examples and anti-examples

### Problem

MCP catalog exists, but tool selection can improve through examples and anti-examples.

### Goal

Improve MCP/tool descriptions so generated agents select tools better.

### Inspect

```text
packages/cli/library/mcp/catalog.json
packages/cli/internal/adapter/mcp_compiler.go
packages/cli/internal/scaffold/mcp.go
```

### For each canonical MCP server, add or validate

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

### Example expectation

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

### Acceptance criteria

```text
- Catalog remains valid.
- Generated MCP output remains compatible.
- Tool descriptions improve without bloating always-loaded instructions.
- Tests updated if schema changes.
```

---

## Priority 10 — Adapter capability documentation and beta graduation path

### Problem

OMP and Antigravity are beta. Pi MCP is a deliberate no-op. Kiro omits agents/skills by design. Capability differences must be explicit.

### Goal

Document and test adapter capability truth.

### Add/update

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

### Each adapter doc should include

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

### Acceptance criteria

```text
- Beta status is visible and justified.
- Pi MCP no-op is explicit and intentional.
- Kiro limitations are explicit.
- Capability matrix matches code.
- Tests fail if beta set unexpectedly changes.
```

---

## Priority 11 — Context compaction and handoff templates

### Problem

Memory/handoff assets exist, but context compaction should be explicitly event-driven.

### Goal

Add or consolidate templates for context discipline.

### Add/update

```text
packages/cli/library/fragments/context-compaction.md
packages/cli/library/templates/context-handoff.md
packages/cli/library/templates/session-compaction.md
packages/cli/library/templates/recovery-handoff.md
```

### Rules

Compact after:

```text
- research phase
- planning phase
- implementation phase
- failed verification
- agent/tool handoff
- before context gets too large
```

Never compact away:

```text
- user constraints
- decisions
- failed attempts
- commands run
- files touched
- unresolved risks
- human approval state
```

### Acceptance criteria

```text
- Handoff templates require evidence.
- Context compaction is framed as guidance/assets, not runtime automation.
- Existing memory/handoff skills reference these templates if appropriate.
```

---

## Priority 12 — Multi-agent decision guide without orchestration

### Problem

Specialist agents and parallel-execution guidance exist, but the decision boundary should be explicit.

### Goal

Document when to use skills vs agents vs host-tool subagents vs parallel research.

### Add

```text
docs/concepts/multi-agent-boundary.md
packages/cli/library/templates/multi-agent-decision.md
packages/cli/library/fragments/specialist-agent-guidance.md
```

### Required guidance

Use one agent + skills when:

```text
- task is local
- context fits
- one person/agent can reason through it
- no isolation needed
```

Use specialist agents when:

```text
- role separation improves quality
- review should be independent
- domain expertise differs
```

Use parallel agents when:

```text
- work is read-heavy
- outputs are easy to merge
- boundaries are explicit
```

Avoid multi-agent when:

```text
- agents need the same full context
- parallel writes would conflict
- merge cost exceeds benefit
- host tool cannot show evidence
```

### Acceptance criteria

```text
- No orchestration command added.
- No runtime scheduler added.
- Guidance is compiled as assets/templates where appropriate.
```

---

## Priority 13 — Official tool compliance matrix refresh

### Problem

The product targets multiple host tools. Adapter behavior must stay aligned with official docs and current local assumptions.

### Goal

Refresh or validate the compliance matrix against current implemented behavior.

### Inspect existing

```text
docs/lazyai-vibelab-product-spec-pack/03_OFFICIAL_TOOL_COMPLIANCE_MATRIX*
docs/adapters/
specs/029-lazyai-v2/
packages/cli/internal/adapter/capabilities.go
packages/cli/internal/adapter/output_mapping.go
```

### Update the matrix to include

```text
OpenCode
Claude Code
GitHub Copilot
Pi
OMP
Gemini/Antigravity
Kiro
```

### For each

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

### Acceptance criteria

```text
- Matrix matches code.
- Gaps are explicit.
- Beta adapters have documented graduation criteria.
```

---

## Priority 14 — Minimality / token-rent integration

### Problem

`internal/tokenrent` and `internal/minimality` exist but are lightly documented.

### Goal

Make token-rent/minimality visible in validation and docs if already implemented.

### Research

```text
packages/cli/internal/tokenrent/
packages/cli/internal/minimality/
packages/cli/library/manifests/curation.yaml
```

### Improve

```text
- docs explaining token rent
- validation warning for large always-loaded assets
- curation manifest guidance
- minimality report docs
```

### Acceptance criteria

```text
- No breaking changes.
- Users understand why some assets are excluded or progressive-disclosure only.
- Validation output is actionable.
```

---

## Priority 15 — Init headless populate clarity

### Problem

`init` headless populate is skipped because host AI tool is needed.

### Goal

Make this behavior explicit and safe.

### Inspect

```text
packages/cli/cmd/init.go
```

### Improve docs/help text around

```text
- why AGENTS.md placeholder fill may be skipped
- what host tool should do next
- how to complete setup manually
- how to validate after populate
```

### Acceptance criteria

```text
- No hidden runtime behavior added.
- CLI explains what happened and next steps.
- Tests updated if output/help changes.
```

---

## Priority 16 — Server L3 handshake clarity

### Problem

Server L3 handshake is not implemented in Go because it requires Node.js MCP SDK; only L1 config checks exist.

### Goal

Make server capability levels explicit.

### Inspect

```text
packages/cli/cmd/server.go
```

### Improve

```text
- docs
- help output
- doctor/status messaging
- tests if available
```

### Acceptance criteria

```text
- Users can distinguish L1 config check vs L3 handshake.
- No fake support is claimed.
- Missing L3 implementation is documented as intentional/current limitation.
```
