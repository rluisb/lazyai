# Suggested PR Sequence — LazyAI / vibe-lab Improvements

Use this sequence to keep the work reviewable.

## PR 1 — Adapter conformance fixtures

Title:

```text
Add adapter conformance fixtures and golden output tests
```

Scope:

```text
packages/cli/testdata/projects/*
packages/cli/testdata/golden/*
adapter golden tests
compile contract regression tests
```

Success criteria:

```text
- all 7 adapters covered
- beta adapters covered
- Codex rejection covered
- drift refusal covered
- legacy MCP migration covered
- dry-run behavior covered
```

## PR 2 — Harness principles and boundary docs

Title:

```text
Consolidate LazyAI harness principles and runtime boundary
```

Scope:

```text
docs/concepts/harness-principles.md
docs/concepts/trace-evidence-over-vibes.md
docs/concepts/no-runtime-orchestration.md
README links
ADR cross-links
status drift cleanup if small enough
```

Success criteria:

```text
- one canonical doctrine exists
- LazyAI/vibe-lab/host-tool boundary is unambiguous
- optional runtime-adjacent extras are explained
```

## PR 3 — Skill and agent quality validation

Title:

```text
Add semantic validation for skills and agent contracts
```

Scope:

```text
internal/compiler/skill_validate.go
internal/compiler/agent_validate.go
skill schema or Markdown linting
agent contract schema or Markdown linting
validation tests
```

Checks:

```text
- trigger guidance
- non-trigger guidance
- progressive disclosure
- required evidence
- misuse risks
- missing references
- unsupported adapter claims
```

## PR 4 — Hook lifecycle catalog

Title:

```text
Introduce neutral hook lifecycle catalog and adapter capability mapping
```

Scope:

```text
internal/hooks/lifecycle.go
internal/hooks/capabilities.go
packages/cli/library/hooks/catalog.md
hook validation
adapter support table
```

Lifecycle:

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

## PR 5 — Trace/eval improvement workflow

Title:

```text
Add trace taxonomy and harness improvement templates
```

Scope:

```text
packages/cli/library/templates/trace-failure.md
packages/cli/library/templates/harness-change-report.md
packages/cli/library/templates/eval-promotion-checklist.md
packages/cli/library/templates/holdout-review.md
docs/concepts/trace-eval-improvement-loop.md
```

Success criteria:

```text
- no scoring engine added
- no judge runtime added
- no orchestration added
- trace-to-eval workflow documented and validated structurally
```

## PR 6 — Rubrics and MCP catalog examples

Title:

```text
Add reusable rubrics and improve MCP catalog guidance
```

Scope:

```text
packages/cli/library/rubrics/* or templates/rubrics/*
packages/cli/library/mcp/catalog.json
schema/tests if needed
```

Success criteria:

```text
- rubrics are discoverable
- MCP tool examples/anti-examples improve tool choice
- generated MCP output remains compatible
```

## PR 7 — Adapter docs and compliance matrix

Title:

```text
Document adapter capability matrix and beta graduation path
```

Scope:

```text
docs/adapters/capability-matrix.md
docs/adapters/opencode.md
docs/adapters/claude.md
docs/adapters/copilot.md
docs/adapters/pi.md
docs/adapters/omp.md
docs/adapters/antigravity.md
docs/adapters/kiro.md
docs/lazyai-vibelab-product-spec-pack/03_OFFICIAL_TOOL_COMPLIANCE_MATRIX*
```

Success criteria:

```text
- beta status visible and justified
- Pi MCP no-op explicit
- Kiro limitations explicit
- matrix matches code
```

## PR 8 — Context, multi-agent boundary, minimality, and CLI clarity

Title:

```text
Clarify context discipline, multi-agent boundaries, token rent, init, and server capability levels
```

Scope:

```text
context/handoff templates
multi-agent boundary docs
token-rent/minimality docs
init headless populate help/docs
server L3 handshake help/docs
```

Success criteria:

```text
- no orchestration added
- context compaction is event-driven guidance
- token-rent/minimality is documented
- init/server limitations are honest and clear
```
