# Pi - Agent Teams & Agent Chains

## Executive Summary
Pi (`@mariozechner/pi-coding-agent`, v0.66.0, 33k+ stars) is a MINIMAL terminal coding harness with an aggressive extensibility philosophy. The core deliberately has NO built-in sub-agents, plan mode, or MCP — everything is extensions. The `pi-subagents` third-party extension (686 stars) provides the MOST SOPHISTICATED chain system of all 6 tools: native `.chain.md` files with sequential/parallel/fan-out execution, worktree isolation, background execution, and a full TUI agent manager.

⚠️ Our codebase incorrectly labels Pi as a "Claude Code wrapper" — it is an independent ecosystem.

## Architecture

### Pi Core (Extension-First)
- Core deliberately provides ZERO multi-agent features
- Philosophy: "Spawn pi instances via tmux, or build your own with extensions"
- 4 modes: Interactive, Print/JSON, RPC, SDK
- Extensions are first-class: tools, skills, prompt templates, themes

### pi-subagents Extension (Third-Party)
- Provides the richest chain system of ALL 6 tools
- Agent files: `.pi/agents/{name}.md` with YAML frontmatter
- Chain files: `.pi/agents/{name}.chain.md` with chain/parallel/fan-out specs
- Built-in agents: scout, planner, worker, reviewer, context-builder, researcher, delegate

## Chain System (UNIQUE TO PI)

### Sequential Chains
```yaml
chain:
  - agent: scout
    task: "Research the codebase for {task}"
  - agent: planner
    task: "Create plan based on: {previous}"
  - agent: worker
    task: "Implement the plan: {previous}"
  - agent: reviewer
    task: "Review the changes: {previous}"
```
Variables: `{task}` (original), `{previous}` (last result), `{chain_dir}` (artifacts)

### Parallel Execution
```yaml
tasks:
  - agent: scout
    task: "Research auth module"
  - agent: scout
    task: "Research database layer"
  - agent: scout
    task: "Research API endpoints"
```

### Fan-out/Fan-in (Parallel within Chain)
```yaml
chain:
  - agent: planner
    task: "Break down {task} into sub-tasks"
  - parallel:
      - agent: worker
        task: "Implement sub-task 1: {previous}"
      - agent: worker
        task: "Implement sub-task 2: {previous}"
      - agent: worker
        task: "Implement sub-task 3: {previous}"
  - agent: reviewer
    task: "Review all implementations: {previous}"
```

### Per-Step Overrides
```yaml
chain:
  - agent: scout
    task: "Research {task}"
    output: context.md
    model: claude-sonnet-4
    reads: ["src/**/*.ts"]
  - agent: worker
    task: "Implement based on: {previous}"
    skills: ["tdd-loop"]
    progress: true
```
Inline syntax: `scout[output=context.md,model=claude-sonnet-4]`

## Slash Commands
- `/run <agent> <task>` — single agent dispatch
- `/chain agent1 "task1" -> agent2 "task2"` — sequential chain
- `/parallel agent1 "task1" -> agent2 "task2"` — parallel execution
- `/subagents-status` — async status overlay
- `/agents` or Ctrl+Shift+A — Agents Manager TUI

## Execution Modes
- **Foreground** (default): blocks until complete
- **Background** (`--bg`): async, trackable
- **Forked** (`--fork`): branched session from parent's current leaf
- **Worktree** (`worktree: true`): each parallel agent gets own git worktree
- **Concurrency**: configurable parallel thread count
- **failFast**: stop all on first failure

## Agent Frontmatter
```yaml
name: scout
description: Read-only codebase researcher
tools: [Read, Grep, Glob, "mcp:codegraph"]
extensions: [pi-subagents]
model: claude-sonnet-4
thinking: medium  # off/minimal/low/medium/high/xhigh
skill: research
output: findings.md
defaultReads: ["src/**/*"]
defaultProgress: true
interactive: false
maxSubagentDepth: 2
```

## Gap Analysis

1. CRITICAL: Pi labeled as "Claude Code wrapper" in our codebase — it is NOT
2. CRITICAL: `stripYamlFrontmatter` removes ALL frontmatter including tools, extensions, model, thinking, output
3. CRITICAL: No chain files generated (`.chain.md`) — Pi's MOST UNIQUE feature
4. HIGH: No extension dependency declared (pi-subagents must be installed separately)
5. HIGH: No root template exists for Pi — it has no root instruction file guidance
6. HIGH: No execution mode guidance (background, fork, worktree)
7. MEDIUM: No slash command documentation in generated files
8. MEDIUM: No per-step override guidance in chains
9. LOW: No TUI agent manager guidance

## Implementation Alternatives

### Option A: Chain File Generation (CRITICAL)
Generate `.chain.md` files alongside agent `.md` files:
- `pipeline.chain.md`: Scout → Planner → Builder → Reviewer → Documenter
- `research.chain.md`: Scout parallel fan-out
- `review.chain.md`: Reviewer + Red-Team parallel
Effort: Medium, Risk: Low

### Option B: Rich Frontmatter Preservation (CRITICAL)
STOP stripping frontmatter. Preserve/enhance:
- tools, extensions, model, thinking, output, defaultReads, maxSubagentDepth
Map our agents:
- Scout: tools=[Read,Grep,Glob], thinking=medium, output=findings.md
- Planner: tools=[Read,Grep,Glob], thinking=high, output=plan.md
- Builder: tools=[Read,Grep,Glob,Edit,Write,Bash], thinking=medium
- Reviewer: tools=[Read,Grep,Glob], thinking=high, output=review.md
- Red-Team: tools=[Read,Grep,Glob,Bash], thinking=high, output=security-review.md
- Documenter: tools=[Read,Grep,Glob,Edit,Write], thinking=low
Effort: Medium, Risk: Low

### Option C: Extension Dependency Management
Add pi-subagents as a required extension:
- Check if installed during `ai-setup install`
- Provide install instructions if missing
- Configure extension settings
Effort: Medium, Risk: Medium

### Option D: Root Template Creation
Create Pi-specific root template (`INSTRUCTIONS.template.md` or similar):
- Include chain/parallel dispatch guidance
- Slash command reference
- Extension configuration
Effort: Low, Risk: Low

### Option E: Full Stack (A+B+C+D)
All of the above for complete Pi support.
