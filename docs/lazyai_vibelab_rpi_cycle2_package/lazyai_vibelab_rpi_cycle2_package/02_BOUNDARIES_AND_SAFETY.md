# Cycle 2 Boundaries and Safety Rules

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

## Allowed work

```text
- skill semantic validation
- agent contract validation
- validation warnings/errors
- schema or lint helpers
- tests and fixtures
- docs/templates explaining skill and agent quality
- small asset updates required to satisfy validation
```

## Severity guidance

Use **errors** only for objective issues:

```text
- invalid schema/frontmatter
- broken references
- unsupported adapter target
- missing required structural fields
- malformed file
- empty required body
```

Use **warnings** for semantic quality issues:

```text
- weak trigger guidance
- missing non-trigger guidance
- missing evidence requirements
- missing output contract
- missing human gate guidance
- ambiguous role/scope
- no examples or anti-examples
- no progressive-disclosure guidance
```

## Required safety checks before editing

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

## Safe commands

```bash
find packages/cli/library/skills -maxdepth 2 -type f | sort
find packages/cli/library/canonical/agents -maxdepth 2 -type f | sort
rg "validate|skill|agent|frontmatter|contract|evidence|trigger|when to use|handoff|human" packages/cli/internal packages/cli/library docs specs
go test ./packages/cli/internal/compiler/...
go test ./packages/cli/internal/validate/...
go test ./packages/cli/cmd/...
go test ./...
```
