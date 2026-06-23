# LazyAI / vibe-lab — Boundaries and Safety Rules

Use this as a required preamble for local AI agents working on LazyAI / vibe-lab.

## Repository context

- Repo root: `/Users/ricardo/projects/teachable/lazyai`
- Product: `lazyai-cli`, a Go CLI that manages a canonical `.ai/` source tree and compiles it into tool-native AI-harness surfaces.
- LazyAI is **not** an agent runtime.
- vibe-lab is the embedded quality/process/taste layer: RPI workflow, human gates, anti-slop rules, skills, hooks, rubrics, handoff discipline, clean-code-for-agents, and trace/eval thinking.
- Host tools execute the generated assets: OpenCode, Claude Code, GitHub Copilot, Pi, OMP, Gemini/Antigravity, Kiro.
- LazyAI owns canonical source, schemas, manifests, validation, compilation, adapter output, doctor/status/update/eject, and optional runtime-adjacent state.
- LazyAI must not become a workflow/orchestration/runtime engine.

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
