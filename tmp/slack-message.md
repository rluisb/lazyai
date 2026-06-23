:rocket: *LazyAI is out* — define your AI-tool setup *once*, compile it to every AI surface

Hey team :wave: — quick share on *LazyAI* (`lazyai-cli`), a Go CLI Alisson and I have been building out. If you bounce between AI coding tools (or we standardize on one), this keeps your agents, skills, hooks, and MCP servers *consistent across every project* instead of hand-maintaining configs per tool. Alisson and I run it on *bee-gone* to keep that repo's AI setup in sync.

*The core idea*
You author one canonical source tree under `.ai/` (`.ai/lazyai.json`, `.ai/mcp.json`, asset trees). `lazyai-cli compile` then emits *tool-native* config for whichever surfaces you use. Re-compiling is idempotent, so it only touches managed regions, tracked via `.ai/lock.json`. Go-only, `go install`, no npm/npx.

*What it ships out of the box*
• *8 canonical agents* — `guide` (front door) + `researcher`, `planner`, `implementer`, `reviewer`, `deployer`, `responder`, `evidence-verifier`
• *31 skills* — engineering practices, Spec-Kit workflow, authoring/governance
• *Hooks* — destructive-shell guard, workflow gate, memory promotion, pre-commit, RPI gate, etc.
• *MCP servers* — ai-memory, filesystem, ripgrep, codegraph, obsidian (wired into each tool's native MCP config)

*Supported targets:* OpenCode, Claude Code, GitHub Copilot, *Kiro*, Pi, plus OMP and Antigravity (beta).

:star2: *Kiro is a first-class target*
Point LazyAI at Kiro and it lays down a full `.kiro/` setup, no manual wiring:
```bash
lazyai-cli init --scope project --tools kiro --preset full --name bee-gone --no-interactive
lazyai-cli compile
```

You get:
• `.kiro/agents/<name>.md` — all 8 agents as *Kiro CLI v3 custom agent profiles*
• `.kiro/skills/<name>/SKILL.md` — the full skill library
• `.kiro/prompts/*.md` — prompt templates
• `.kiro/settings/mcp.json` — MCP servers ready to go
• `AGENTS.md` — shared root instructions

So the same agents/skills your teammates use in Claude/OpenCode show up identically in Kiro. Add a server later with `lazyai-cli server add filesystem && lazyai-cli compile` and it updates the Kiro MCP config for you.

*Get started*
```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
cd bee-gone
lazyai-cli init --scope project --tools kiro,claude-code --preset full --name bee-gone --no-interactive
lazyai-cli compile && lazyai-cli status
```

*Docs:* new *Reference* section just landed — per-tool output files & structure, agent/skill/hook/MCP catalogs. Check `Reference → Tool Outputs` for exactly what gets written where.

Questions / want it set up for your repo? Ping me. :pray:
