# OpenCode — Agent Teams & Agent Chains

## Architecture Overview
OpenCode has a **two-tier agent system** (Primary Agents + Subagents) with a unique **`session()` tool** that provides 4 collaboration modes. It has **NO native "agent teams" feature** and **NO native chain enforcement**. 

- **Primary Agents**: Switchable via Tab (Build, Plan). Persist across conversation.
- **Subagents**: Invoked via Task tool (General, Explore, Custom). Own context window.
- **session() tool**: 4 Pillars - Collaborate (message), Handoff (new), Compress (compact), Parallelize (fork). 

## Gap Analysis (Current ai-setup implementation)
1. **Frontmatter Stripping**: Current adapter (`stripFrontmatterAndInjectModel`) removes `mode`, `permissions`, `steps`, and `task` permission fields.
2. **Missing Isolation**: No permissions set per agent - all agents have full tool access.
3. **Missing Spawn Control**: No task permission control - any agent can spawn any other.
4. **Missing Coordination**: No session() guidance in templates for structured handoffs.
5. **No custom tools**: Not leveraging `.opencode/tools/*.ts` for orchestrating chains.

## Implementation Alternatives

### Option A: Rich Frontmatter Generation (Recommended)
Replace `stripFrontmatterAndInjectModel` with OpenCode-native frontmatter generation.
- Set mode: subagent for scout/reviewer/documenter, primary for planner/builder.
- Set granular permissions per agent (read-only vs full).
- Set task permissions (glob patterns) to control the spawn graph.
*Effort: Low-Medium, Risk: Low*

### Option B: session() Chain Protocol
Add `session()` usage guidance to `AGENTS.template.md`.
- Define sequential handoff protocol: plan->build->plan
- Define parallel exploration protocol using fork mode
*Effort: Low, Risk: Very Low*

### Option C: Custom Tool Orchestrator
Generate `.opencode/tools/pipeline.ts` with a programmatic state machine pipeline.
*Effort: High, Risk: High*

### Option D: Hybrid (Option A + B)
Best coverage: Rich frontmatter for isolation + session() guidance for coordination.
