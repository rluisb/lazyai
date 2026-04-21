# GitHub Copilot (VS Code) - Agent Teams & Agent Chains

## Executive Summary
GitHub Copilot has the MOST sophisticated agent orchestration of all 6 tools:
- **Handoffs**: Built-in sequential agent chains with UI buttons
- **Nested subagents**: Up to depth 5 (unique - all others block nesting)
- **Spawn restrictions**: `agents` frontmatter with list, wildcard `*`, empty `[]`
- **Subagent-only agents**: `user-invocable: false`
- **Multi-environment**: Local/CLI/Cloud/Third-Party with session handoff
- **Model fallback**: Priority-ordered model arrays
- **Organization sharing**: GitHub-level agent distribution

## Architecture
- 4 execution environments: Local (VS Code), Copilot CLI, Cloud (GitHub), Third-Party (Anthropic/OpenAI)
- 3 built-in agents: Agent (implementation), Plan (planning), Ask (questions)
- Subagents via `runSubagent` tool, nested up to depth 5 (configurable)
- Handoffs = guided sequential workflows with UI buttons after each response
- Custom agents: `.agent.md` or `.md` in `.github/agents/`, `~/.copilot/agents`, `.claude/agents/`
- Frontmatter: name, description, tools, agents, model, user-invocable, disable-model-invocation, target, mcp-servers, handoffs, hooks

## Key Mechanism: HANDOFFS
Handoffs are THE killer feature. Each handoff defines: label, agent (target), prompt, send (auto-submit), model.
Buttons appear after chat response completes. User clicks to transition to next agent.
This gives us a native, UI-supported agent chain mechanism.

## Gap Analysis
1. CRITICAL: handoffs not generated - missing native agent chains
2. HIGH: agents (spawn restrictions) not set
3. HIGH: tools per agent stripped with frontmatter
4. HIGH: user-invocable not used for subagent-only agents
5. MEDIUM: target (env scoping) not set
6. MEDIUM: disable-model-invocation not used
7. LOW: model fallback arrays not generated

## Implementation Alternatives

### Option A: Handoff Chain Generation (CRITICAL)
Generate handoffs frontmatter: Scout->Planner->Builder->Reviewer->Documenter chain.
Add Red-Team as optional branch from Planner.
Effort: Medium, Risk: Low

### Option B: Rich Copilot Frontmatter
Replace stripFrontmatterAndInjectModel with full Copilot-native frontmatter.
Include tools, agents, user-invocable, model fallback arrays.
Effort: Low-Medium, Risk: Low

### Option C: Coordinator Pattern
Create orchestrator agent with agents: [all agents], uses runSubagent for automated chains.
Effort: Low, Risk: Low

### Option D: Full Hybrid (A+B+C)
Handoffs for user-driven chains + frontmatter for isolation + coordinator for automated chains.
Best coverage.
