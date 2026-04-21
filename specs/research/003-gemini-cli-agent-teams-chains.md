# Gemini CLI — Agent Teams & Agent Chains

## CRITICAL DISCOVERY
Our adapter (`gemini.ts`) says "Gemini CLI has NO agents concept" — this is OUTDATED.
Gemini CLI now has FULL subagent support with `.gemini/agents/*.md`, rich YAML frontmatter, tool isolation, plan mode, remote agents (A2A), policy engine, and extension subagents.

## Architecture
- **Subagents**: Custom `.gemini/agents/*.md` with YAML frontmatter (name, description, tools, model, max_turns, timeout_mins)
- **Built-in subagents**: codebase_investigator, cli_help, generalist_agent, browser_agent (experimental)
- **Tool wildcards**: `*`, `mcp_*`, `mcp_server_*` — unique to Gemini
- **Recursion protection**: subagents CANNOT call other subagents
- **Plan mode**: Built-in two-phase chain (plan→implement) with auto model routing (Pro→Flash)
- **Remote agents**: Agent2Agent (A2A) protocol — UNIQUE to Gemini
- **Policy engine**: policy.toml with subagent-specific tool restriction rules
- **Extension subagents**: Distributable agent bundles

## Gap Analysis
1. CRITICAL: Adapter skips ALL agent installation
2. No frontmatter generation for Gemini-native format
3. No policy.toml generation for tool restrictions
4. No settings.json generation for per-agent config
5. No plan mode guidance in GEMINI.template.md

## Implementation Alternatives

### Option A: Add Agent Installation (CRITICAL - Must Do)
- Update gemini.ts to install agents to `.gemini/agents/`
- Generate Gemini-native frontmatter with tools list, model, max_turns
- Map our agent permissions to Gemini tool lists with wildcards
- Effort: Medium, Risk: Low

### Option B: Policy Generation
- Generate `.gemini/policies/policy.toml` with per-subagent rules
- Effort: Medium, Risk: Low

### Option C: Plan Mode Integration
- Add plan mode workflow to GEMINI.template.md
- Map Scout+Planner to plan phase, Builder to implementation
- Effort: Low, Risk: Very Low

### Option D: Settings Generation
- Generate `.gemini/settings.json` with per-agent model/limit overrides
- Effort: Medium, Risk: Low

### Option E: Full Stack (A+B+C+D)
- Best coverage, highest effort
