# Codex CLI - Agent Teams & Agent Chains

## Executive Summary
Codex CLI has full subagent support with TOML-based agent configuration (unique among all tools), configurable parallelism (max_threads up to 6), sandbox modes (read-only vs workspace-write), and experimental CSV batch fan-out. Our adapter says "Agents inline in AGENTS.md" but Codex fully supports custom agents as `.codex/agents/*.toml` files.

## Architecture
- 3 execution environments: IDE, CLI, Cloud
- Built-in agents: default (general-purpose), worker (execution-focused), explorer (read-heavy)
- Custom agents: TOML files in `.codex/agents/*.toml` or `~/.codex/agents/*.toml`
- TOML schema: name, description, developer_instructions, sandbox_mode, model, model_reasoning_effort, mcp_servers, skills.config
- Global config: `.codex/config.toml` with agents.max_threads (default 6), agents.max_depth (default 1)
- Sandbox modes: read-only (no writes) vs workspace-write (full access within workspace)
- CSV batch fan-out: spawn_agents_on_csv for parallel batch processing (experimental)
- Plan-then-implement workflow: $plan skill -> cloud delegation -> /review

## Gap Analysis
1. CRITICAL: Custom agent files not generated (.codex/agents/*.toml)
2. CRITICAL: No TOML generation from our Markdown agent definitions
3. HIGH: sandbox_mode not set per agent
4. HIGH: No config.toml generation for max_threads/max_depth
5. HIGH: developer_instructions not extracted from agent body
6. MEDIUM: No workflow guidance in AGENTS.md template

## Implementation Alternatives

### Option A: TOML Agent Generation (CRITICAL)
New transform function markdownAgentToToml() to generate .codex/agents/*.toml.
Map sandbox_mode per agent (read-only for scout/planner/reviewer/red-team, workspace-write for builder/documenter).
Effort: Medium, Risk: Low

### Option B: Config.toml Generation
Generate .codex/config.toml with agents.max_threads and agents.max_depth.
Effort: Low, Risk: Very Low

### Option C: Workflow Templates
Add plan-then-implement and parallel dispatch guidance to AGENTS.md template.
Effort: Low, Risk: Very Low

### Option D: Full Stack (A+B+C)
TOML agents + config.toml + workflow templates for complete coverage.
