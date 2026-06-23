# RPI Cycle 5 — MCP examples, adapter capability docs, official compliance matrix, and token-rent/minimality

You are running locally inside the LazyAI repository.

Repo context:

```text
Repo root: /Users/ricardo/projects/teachable/lazyai
Product: lazyai-cli, a Go CLI harness asset manager/compiler.
LazyAI owns: .ai source, manifests, schemas, validation, adapter compilation, generated output.
vibe-lab owns: process/quality/taste assets, RPI workflow, human gates, skills, hooks, rubrics, handoff discipline.
Host tools execute generated assets.
LazyAI must not become a runtime or orchestrator.
```

Cycle goal:

```text
Improve MCP/tool-selection guidance, adapter capability documentation, official compliance tracking, and token-rent/minimality visibility.
```

Use RPI:

```text
R = Research the current implementation before changing anything.
P = Plan a small, safe, non-breaking implementation.
I = Implement with tests and clear final reporting.
```

---

# Shared LazyAI / vibe-lab Boundaries

These apply to every remaining RPI cycle.

## Product boundary

LazyAI must remain:

```text
- a canonical .ai source manager
- a validation layer
- a manifest/schema/lockfile owner
- a compiler into host-tool-native AI harness surfaces
- an optional runtime-adjacent state manager
```

vibe-lab remains:

```text
- the process and quality layer
- the source of RPI workflow guidance
- the source of human gates, anti-slop rules, skills, hooks, rubrics, and handoff discipline
```

Host tools execute generated assets:

```text
OpenCode
Claude Code
GitHub Copilot
Pi
OMP
Gemini / Antigravity
Kiro
```

## Do not add or reintroduce

```text
lazyai run workflow
lazyai autonomous task
lazyai orchestrate
lazyai judge
lazyai eval runner
mandatory trace daemon
mandatory RAG core
background subagent runtime
LangChain/LangGraph/CrewAI dependency in core
old task/workflow/orchestration/eval commands
Codex adapter
```

## Safety checks before editing

```bash
pwd
git rev-parse --show-toplevel
git status --short
git branch --show-current
git log -1 --oneline
```

## Never run

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

## Final boundary confirmation required

Every cycle report must confirm:

```text
- No runtime/orchestration surface was added.
- No old task/workflow/eval command was reintroduced.
- No mandatory judge/scoring engine was added.
- No mandatory trace daemon was added.
- No mandatory RAG core was added.
- LazyAI remains a harness asset manager/compiler.
- vibe-lab remains the process/quality asset layer.
```


---

# Research Phase — Cycle 5

Inspect MCP, adapter docs, compliance matrix, and minimality/token-rent.

## Required inspection

```text
packages/cli/library/mcp/catalog.json
packages/cli/internal/adapter/mcp_compiler.go
packages/cli/internal/scaffold/mcp.go
packages/cli/internal/adapter/capabilities.go
packages/cli/internal/adapter/output_mapping.go
packages/cli/internal/tokenrent/
packages/cli/internal/minimality/
packages/cli/library/manifests/curation.yaml
docs/lazyai-vibelab-product-spec-pack/
docs/adapters/
specs/029-lazyai-v2/
```

Answer:

```text
- Does MCP catalog include examples and anti-examples?
- How does each adapter emit MCP config?
- Which adapters are beta?
- What does Pi intentionally not emit?
- What does Kiro intentionally omit?
- Is there an adapter capability matrix?
- Is there an official tool compliance matrix?
- How are token-rent and minimality documented?
```


---

# Plan Phase — Cycle 5

Create a small plan.

Include:

```text
1. MCP catalog changes.
2. Adapter docs/capability matrix changes.
3. Official compliance matrix changes.
4. Token-rent/minimality docs changes.
5. Tests to add/update.
6. Risk of schema changes.
```

Prefer backward-compatible catalog updates.


---

# Implementation Scope — Cycle 5

## Priority A — MCP catalog examples and anti-examples

For canonical MCP servers:

```text
ai-memory
filesystem
ripgrep
codegraph
obsidian
```

Add or validate:

```text
purpose
preferred_use
CLI-first vs MCP-first guidance
examples
anti_examples
security_notes
adapter_output_differences
```

Example:

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

Acceptance:

```text
- Catalog remains valid.
- Generated MCP output remains compatible.
- Tool descriptions improve without bloating always-loaded instructions.
```

## Priority B — Adapter capability documentation

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
status: stable/beta/partial
generated files
supported asset types
unsupported asset types
MCP behavior
hook behavior
skill behavior
agent behavior
known limitations
tests/fixtures proving behavior
```

Acceptance:

```text
- OMP and Antigravity beta status is visible and justified.
- Pi MCP no-op is explicit and intentional.
- Kiro limitations are explicit.
- Capability matrix matches code.
```

## Priority C — Official compliance matrix refresh

Update existing compliance matrix if present:

```text
docs/lazyai-vibelab-product-spec-pack/03_OFFICIAL_TOOL_COMPLIANCE_MATRIX*
```

Include:

```text
OpenCode
Claude Code
GitHub Copilot
Pi
OMP
Gemini / Antigravity
Kiro
```

If external web access is unavailable, mark:

```text
requires external docs refresh
```

Do not invent official requirements.

Acceptance:

```text
- Matrix matches current code.
- Gaps are explicit.
- Beta graduation criteria are documented.
```

## Priority D — Token-rent and minimality docs

Document:

```text
internal/tokenrent
internal/minimality
curation token_rent flags
minimality report behavior
```

Add or update:

```text
docs/concepts/token-rent.md
docs/concepts/minimality.md
```

Acceptance:

```text
- Users understand why some assets are excluded or progressive-disclosure only.
- Validation/report output is actionable.
- No breaking changes.
```


---

# Required validation

Run targeted tests for the changed area first.

Then run if safe:

```bash
go test ./packages/cli/...
go test ./...
go vet ./...
```

If broad tests fail for unrelated existing reasons, report the exact failure and whether targeted tests passed.

Check:

```bash
git status --short
```

---

# Final Report Template

Use this final report format for the cycle.

```markdown
# LazyAI / vibe-lab RPI Cycle <N> Report — <Title>

## 1. Research summary

- Repo state:
- Existing implementation confirmed:
- Missing pieces confirmed:
- Important constraints:

## 2. Plan executed

| Item | Status | Notes |
|---|---|---|

## 3. Changes made

| Area | Files changed | Summary |
|---|---|---|

## 4. Details by priority

### Priority A

- What changed:
- Tests:
- Remaining gaps:

### Priority B

- What changed:
- Tests:
- Remaining gaps:

### Priority C

- What changed:
- Tests:
- Remaining gaps:

## 5. Tests and validation

Commands run:

```bash
...
```

Results:

```text
...
```

## 6. Changed files

```text
...
```

## 7. Risks

| Risk | Impact | Mitigation |
|---|---|---|

## 8. Remaining work

| Item | Why not completed | Suggested next step |
|---|---|---|

## 9. Product boundary confirmation

- No runtime/orchestration surface was added.
- No old task/workflow/eval command was reintroduced.
- No mandatory judge/scoring engine was added.
- No mandatory trace daemon was added.
- No mandatory RAG core was added.
- LazyAI remains a harness asset manager/compiler.
- vibe-lab remains the process/quality asset layer.
```
