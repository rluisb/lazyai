# Research: Claude Code — Agent Teams & Agent Chains

**Status:** ✅ Complete  
**Date:** 2026-04-08  
**Tool:** `claude-code`  
**Adapter:** `src/adapters/claude-code.ts`

---

## Executive Summary

Claude Code has the **richest native multi-agent support** of any tool we support, with TWO distinct mechanisms:

1. **Subagents** (Agent Chains) — sequential dispatch via `Task` tool, stable API
2. **Agent Teams** (Parallel Coordination) — experimental, multi-instance coordination

---

## Architecture Overview

```mermaid
flowchart TB
    subgraph "Claude Code Multi-Agent Architecture"
        direction TB
        
        subgraph SUBAGENTS["🔗 SUBAGENTS (Agent Chains)"]
            direction LR
            MAIN[Main Session<br/>Full Context] -->|"Task tool"| SA1[Subagent 1<br/>Own Context]
            MAIN -->|"Task tool"| SA2[Subagent 2<br/>Own Context]
            SA1 -.->|"Result"| MAIN
            SA2 -.->|"Result"| MAIN
            
            note1[" ❌ Subagents CANNOT<br/>spawn subagents<br/>⛓️ Chain via main only"]
        end
        
        subgraph TEAMS["👥 AGENT TEAMS (Parallel Coordination)"]
            direction TB
            LEAD[Team Lead<br/>Orchestrator] -->|"CreateTask"| TL[Shared Task List]
            TL -->|"Assign"| T1[Teammate 1<br/>Own Context + MCP]
            TL -->|"Assign"| T2[Teammate 2<br/>Own Context + MCP]
            TL -->|"Assign"| T3[Teammate 3<br/>Own Context + MCP]
            
            T1 -->|"SendMessage"| LEAD
            T2 -->|"SendMessage"| LEAD
            LEAD -->|"Broadcast"| T1
            LEAD -->|"Broadcast"| T2
            LEAD -->|"Broadcast"| T3
            
            note2[" 🧪 EXPERIMENTAL<br/>Env: CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1"]
        end
    end
```

---

## 1. Subagents (Agent Chains)

### How It Works

```mermaid
flowchart LR
    subgraph "Subagent Dispatch Flow"
        U[User Request] --> M[Main Agent]
        M -->|"1. Task(scout)"| SC[Scout Agent<br/>🔍 Read-only]
        SC -->|"Findings"| M
        M -->|"2. Task(planner)"| PL[Planner Agent<br/>📋 Read-only]
        PL -->|"Plan"| M
        M -->|"3. Task(builder)"| BL[Builder Agent<br/>🔨 Full tools]
        BL -->|"Changes"| M
        M -->|"4. Task(reviewer)"| RV[Reviewer Agent<br/>✅ Read-only]
        RV -->|"Feedback"| M
    end
```

### Configuration

Subagents are defined as `.md` files with YAML frontmatter in `.claude/agents/`:

```yaml
---
name: scout
description: "Read-only codebase researcher"
model: sonnet
tools:
  - Read
  - Grep
  - Glob
  - Bash(read-only)
  - ListFiles
disallowedTools:
  - Write
  - Edit
  - MultiEdit
permissionMode: plan
maxTurns: 50
skills:
  - explore
  - code-standards
isolation: worktree  # optional: run in git worktree
background: false
---

# Scout Agent

Your identity and instructions here...
```

### Scopes Priority (highest → lowest)

1. Managed settings (organization)
2. CLI `--agents` flag
3. Project `.claude/agents/`
4. User `~/.claude/agents/`
5. Plugin agents

### Key Frontmatter Fields

| Field | Purpose | Type |
|-------|---------|------|
| `name` | Display name | string |
| `description` | When to use this agent | string |
| `tools` | Allowed tools | string[] |
| `disallowedTools` | Blocked tools | string[] |
| `model` | Model to use | string |
| `permissionMode` | plan/auto/manual | string |
| `maxTurns` | Max conversation turns | number |
| `skills` | Auto-loaded skills | string[] |
| `mcpServers` | Agent-specific MCP servers | string[] |
| `hooks` | Lifecycle hooks | object |
| `memory` | Memory types (user/project/local) | object |
| `background` | Run in background | boolean |
| `effort` | Reasoning effort level | string |
| `isolation` | worktree isolation | string |
| `color` | Terminal color | string |
| `initialPrompt` | Auto-start prompt | string |

### Critical Constraints

- ❌ **Subagents CANNOT spawn other subagents** — chain must always go through main agent
- ✅ **Chaining via prompt instruction** — "Use scout first, then planner with findings"
- ✅ **SendMessage** can resume a subagent by ID (requires agent teams enabled)
- 🔄 **Auto-compaction** at ~95% context capacity

### Spawn Restriction

Control which subagents the main agent can dispatch:

```yaml
tools:
  - Agent(scout, planner, builder, reviewer, red-team, documenter)
```

---

## 2. Agent Teams (Parallel Coordination)

### How It Works

```mermaid
flowchart TB
    subgraph "Agent Teams Architecture"
        direction TB
        
        LEAD[🎯 Team Lead] -->|"CreateTask"| TASKS[(Shared Task List)]
        
        TASKS -->|"task: research auth"| W1[👷 Teammate: Scout<br/>worktree: auth-research]
        TASKS -->|"task: write tests"| W2[👷 Teammate: Builder<br/>worktree: test-branch]  
        TASKS -->|"task: review PR"| W3[👷 Teammate: Reviewer<br/>worktree: review-branch]
        
        W1 -->|"✉️ SendMessage"| LEAD
        W2 -->|"✉️ SendMessage"| LEAD
        W3 -->|"✉️ SendMessage"| LEAD
        
        LEAD -->|"📢 Broadcast"| W1
        LEAD -->|"📢 Broadcast"| W2
        LEAD -->|"📢 Broadcast"| W3
        
        subgraph "Task States"
            P[⏳ Pending] --> IP[🔄 In Progress] --> C[✅ Completed]
        end
    end
```

### Requirements

- **Environment variable:** `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
- **Display:** in-process (Shift+Down) or split-panes (tmux/iTerm2)

### Features

| Feature | Description |
|---------|-------------|
| Task dependencies | Tasks can depend on other tasks |
| Plan approval mode | Teammate works read-only until lead approves |
| Communication | message (one) or broadcast (all) |
| Hooks | TeammateIdle, TaskCreated, TaskCompleted |
| Context isolation | Each teammate gets own context window |
| Project context | Teammates load CLAUDE.md, MCP, skills automatically |

### Storage

```
~/.claude/teams/{name}/config.json   # Team definition
~/.claude/tasks/{name}/              # Task persistence
```

### Best Practices

- 3-5 teammates per team
- 5-6 tasks per teammate
- Avoid same-file edits across teammates
- Use worktree isolation to prevent conflicts

### Limitations

- No session resumption for in-process mode
- One team per session
- No nested teams
- Team lead is fixed (cannot change)
- Permissions set at spawn (cannot change mid-session)
- Split-pane mode requires tmux or iTerm2

---

## 3. What We Install Today

### Current Adapter Behavior (`src/adapters/claude-code.ts`)

```
.claude/
├── agents/
│   ├── builder.md       # frontmatter preserved as-is
│   ├── documenter.md
│   ├── planner.md
│   ├── red-team.md
│   ├── reviewer.md
│   └── scout.md
├── skills/
│   └── <name>/SKILL.md
└── CLAUDE.md             # from CLAUDE.template.md
```

### What's Missing

| Feature | Status | Impact |
|---------|--------|--------|
| `tools` frontmatter | ❌ Not set | All agents get all tools |
| `disallowedTools` | ❌ Not set | No tool restrictions |
| `permissionMode` | ❌ Not set | Default permissions |
| `isolation: worktree` | ❌ Not set | No worktree isolation |
| `skills` frontmatter | ❌ Not set | Skills not auto-loaded |
| `mcpServers` per agent | ❌ Not set | All agents share MCP |
| `hooks` | ❌ Not set | No lifecycle hooks |
| `Agent()` restrictions | ❌ Not set | Can spawn any agent |
| `background` mode | ❌ Not set | No background agents |
| `initialPrompt` | ❌ Not set | No auto-start |
| Agent Teams config | ❌ Not generated | No team support |
| Team task templates | ❌ Not generated | No workflow templates |

---

## 4. Implementation Alternatives

### Option A: Enhanced Subagent Chains ⭐ RECOMMENDED FIRST STEP

```mermaid
flowchart TD
    subgraph "Option A Changes"
        A1[Add tools/permissions<br/>to agent frontmatter] --> A2[Add chain instructions<br/>to CLAUDE.md template]
        A2 --> A3[Add Agent restrictions<br/>to control spawn graph]
    end
```

**Changes Required:**

1. **Enrich agent frontmatter** with `tools`, `disallowedTools`, `permissionMode`
2. **Add chain protocol** to CLAUDE.md template
3. **Add `Agent()` restrictions** to main agent

**Effort:** Low — mostly template changes  
**Risk:** Low — uses stable API, backward compatible

### Option B: Agent Teams Config Generation

```mermaid
flowchart TD
    subgraph "Option B Changes"
        B1[New generator:<br/>src/generators/team.ts] --> B2[Map agents to<br/>teammate roles]
        B2 --> B3[Generate task templates<br/>for workflows]
        B3 --> B4[CLI flag: --enable-teams]
    end
```

**Changes Required:**

1. New `src/generators/team.ts` for `config.json`
2. Agent-to-teammate role mapping
3. Task templates for feature/bugfix/review workflows
4. CLI flag `--enable-teams`

**Effort:** Medium — new code paths, new file format  
**Risk:** Medium — experimental API may change

### Option C: Hybrid Approach (Best Coverage)

```mermaid
flowchart TD
    subgraph "Option C: Progressive Enhancement"
        C1[Install chains by default<br/>Option A] --> C2{--teams flag?}
        C2 -->|Yes| C3[Also generate<br/>team configs<br/>Option B]
        C2 -->|No| C4[Chains only<br/>Stable API]
    end
```

```bash
ai-setup install claude-code              # chains only (Option A)
ai-setup install claude-code --teams      # chains + teams (Option A+B)
```

---

## 5. Constraints & Gotchas

| Constraint | Impact |
|-----------|--------|
| Subagents cannot spawn subagents | Chain must go through main agent |
| One team per session | Cannot have competing teams |
| No nested teams | Team lead cannot create sub-teams |
| Experimental flag required | Users must opt in |
| Split-pane needs tmux/iTerm2 | Not all terminals supported |
| 3-5 teammates recommended | Don't create 6 teammates for 6 agents |
| No session resumption (in-process) | Teams are ephemeral |

---

## References

- [Claude Code Subagents Docs](https://docs.anthropic.com/en/docs/claude-code/sub-agents)
- [Claude Code Agent Teams Docs](https://docs.anthropic.com/en/docs/claude-code/agent-teams)
- Adapter: `src/adapters/claude-code.ts`
- Template: `library/root/CLAUDE.template.md`
- Agent definitions: `library/agents/*.md`
