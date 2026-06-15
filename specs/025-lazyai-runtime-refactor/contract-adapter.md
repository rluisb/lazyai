# P0-6: Adapter Test Contract

**Status:** Complete — registered-adapter contract defined; implementation pending
**Owner:** Ricardo Conceicao  
**Date:** 2026-06-13  
**Linked from:** `plan.md` Phase 0, P0-6

---

## Purpose

Neutral contract for each registered adapter mode (OpenCode, Claude Code, Copilot). Gemini and Codex are out of scope — no adapter implementations exist for them in the current codebase.


Current code status: adapters are not yet compliant. `opencode.go` still defaults through `FortniteMode`, and runtime defaults still reference `orchestrator`/`loop-driver` until Phase 1 rewrites them.

## Contract Per Adapter Mode

### OpenCode

| Field | Expected Value |
|---|---|
| Expected files | `.opencode.json`, `AGENTS.md`, agent/skill/hook/command files |
| Agents | `primary-agent` (canonical), plus any surviving canonical agents |
| Defaults | `default_agent: "primary-agent"` |
| Generated commands | Per canonical command inventory |
| Config merge semantics | `.opencode.json` merges with user config; no Fortnite overrides |
| Validation behavior | Validates agent/skill/hook/command presence against canonical library |
| Exclusions | No `loop-driver`, no `orchestrator`, no Fortnite agents, no `STARTUP.md` |

### Claude Code

| Field | Expected Value |
|---|---|
| Expected files | `.claude/CLAUDE.md`, `.claude/settings.json`, agent/skill/hook/command files |
| Agents | `primary-agent` (canonical), plus any surviving canonical agents |
| Defaults | `default_agent: "primary-agent"` |
| Generated commands | Per canonical command inventory |
| Config merge semantics | `.claude/settings.json` merges; `.claude/settings.local.json` for secrets |
| Validation behavior | Validates against canonical library |
| Exclusions | No Fortnite agents, no orchestrator references |

### Copilot

| Field | Expected Value |
|---|---|
| Expected files | `.github/copilot-instructions.md`, agent/skill/hook/command files |
| Agents | `primary-agent` (canonical), plus any surviving canonical agents |
| Defaults | `default_agent: "primary-agent"` |
| Generated commands | Per canonical command inventory |
| Config merge semantics | `.github/copilot-instructions.md` merges with user config |
| Validation behavior | Validates against canonical library |
| Exclusions | No Fortnite agents, no orchestrator references |

## Test Strategy

Table-driven Go tests. Each adapter mode gets a test fixture; each test asserts canonical output, not Fortnite library behavior.

```go
// Example fixture structure
var adapterContractTests = []struct {
    mode    types.ToolId
    want    ContractExpectation
}{
    {mode: "opencode", want: ContractExpectation{
        Files:       []string{".opencode.json", "AGENTS.md", "agents/primary-agent.md"},
        Agents:      []string{"primary-agent"},
        DefaultAgent: "primary-agent",
        NoFortnite:  true,
    }},
    // ... Claude Code, Copilot
}
```

## Gate

⛔ Human must approve this contract before Phase 1 adapter test rewrite begins.
