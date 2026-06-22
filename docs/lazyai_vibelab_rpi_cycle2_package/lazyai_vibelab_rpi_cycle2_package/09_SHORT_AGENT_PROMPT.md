# Short Prompt — RPI Cycle 2

You are running locally inside `/Users/ricardo/projects/teachable/lazyai`.

Work on:

> **RPI Cycle 2 — Semantic validation for skills and agent contracts**

Cycle 1 already added adapter conformance fixtures, golden tests, harness principles docs, and fixed 029 status drift.

Your job now:

```text
Improve LazyAI validation so skills and agents are behaviorally useful, scoped, safe, evidence-driven, and aligned with vibe-lab quality principles.
```

Use RPI:

```text
Research current validation paths, skill format, agent format, and tests.
Plan a small, safe, non-breaking implementation.
Implement semantic validation with tests and actionable output.
```

Do not add:

```text
runtime/orchestration
workflow execution
autonomous task engine
judge/eval runner
mandatory trace daemon
mandatory RAG core
LangChain/LangGraph/CrewAI dependency
Codex adapter
old task/workflow/eval commands
```

Focus on:

```text
1. skill semantic validation
2. agent role contract validation
3. validation rule IDs and fix suggestions
4. docs/templates for skill and agent authors
5. tests for good/bad assets
```

Use errors only for objective breakage:

```text
invalid schema/frontmatter
broken references
unsupported adapter target
missing required structural fields
malformed or empty file
```

Use warnings for quality issues:

```text
missing trigger guidance
missing non-trigger/misuse guidance
missing evidence requirements
missing output contract
missing human gate guidance
missing handoff behavior
ambiguous role/scope
missing progressive disclosure
```

Required checks before editing:

```bash
pwd
git rev-parse --show-toplevel
git status --short
git branch --show-current
git log -1 --oneline
```

Research files:

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/compiler/
packages/cli/internal/validate/
packages/cli/cmd/validate.go
packages/cli/cmd/compile.go
packages/cli/library/skills/
packages/cli/library/canonical/agents/
packages/cli/testdata/
```

Run tests:

```bash
go test -v ./packages/cli/internal/compiler -run 'Test.*Skill|Test.*Agent|Test.*Validate'
go test -v ./packages/cli/internal/validate/...
go test -v ./packages/cli/cmd/...
go test ./...
go vet ./...
```

Final report must include:

```text
research summary
plan executed
files changed
skill validation details
agent validation details
rule IDs
tests run and results
risks
remaining work
product boundary confirmation
```

Confirm explicitly that LazyAI remains a compiler/asset manager and no runtime/orchestration surface was added.
