# Compact Prompt — LazyAI / vibe-lab RPI Improvements

You are running locally inside `/Users/ricardo/projects/teachable/lazyai`.

Use RPI:

```text
R = Research current implementation.
P = Plan safest changes.
I = Implement small, test-backed improvements.
```

LazyAI is a Go CLI harness asset manager/compiler. It owns canonical `.ai/`, schemas, manifests, validation, compilation, adapters, doctor/status/update/eject, and optional runtime-adjacent state.

vibe-lab is the quality/process asset layer: RPI workflow, human gates, anti-slop rules, skills, hooks, rubrics, handoff discipline, clean-code-for-agents, and trace/eval thinking.

Host tools execute generated assets: OpenCode, Claude Code, GitHub Copilot, Pi, OMP, Gemini/Antigravity, Kiro.

Do not turn LazyAI into a runtime.

Do **not** implement:

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

Before edits:

```bash
pwd
git rev-parse --show-toplevel
git status --short
git branch --show-current
git log -1 --oneline
```

Research:

```text
- compile contract: manifest, lockfile, writer, compiler, compile cmd
- adapters: opencode, claude, copilot, pi, omp, kiro, antigravity
- testdata/golden fixture state
- docs/spec status drift
- embedded library assets
- skill/agent validation
- hooks/middleware lifecycle
- trace/eval/rubric state
```

Implement in this order:

```text
1. Adapter conformance fixtures and golden tests
2. Harness principles and product boundary docs
3. Product/spec status drift cleanup
4. Semantic skill validation
5. Agent role contract validation
6. Hook/middleware lifecycle catalog
7. Trace/eval improvement loop templates, no runtime
8. Rubric assets and validation
9. MCP examples and anti-examples
10. Adapter capability docs and beta graduation path
11. Context compaction and handoff templates
12. Multi-agent decision guide without orchestration
13. Official tool compliance matrix refresh
14. Minimality/token-rent docs/validation
15. Init headless populate clarity
16. Server L3 handshake clarity
```

Validation:

```bash
go test ./...
go vet ./...
git status --short
```

Final report must include:

```text
- research summary
- plan executed
- files changed
- tests run and results
- risks
- remaining work
- explicit confirmation that no runtime/orchestration/judge/daemon/RAG core was added
```
