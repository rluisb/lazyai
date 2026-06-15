# Model Routing Policy

## Agent Model Assignments

| Agent | Primary Model | Fallback Models | Tier |
|-------|--------------|-----------------|------|
| orchestrator | Opus or equivalent | GPT-4, Claude-3 | Tier 1 (Router) |
| builder | GPT-4 | GPT-3.5, Claude-3 | Tier 3 (Implement) |
| documenter | GPT-4 | GPT-3.5, Claude-3 | Tier 3 (Implement) |
| implementor | GPT-4 | GPT-3.5, Claude-3 | Tier 3 (Implement) |
| planner | GPT-4 | GPT-3.5, Claude-3 | Tier 3 (Implement) |
| red-team | GPT-4 | Claude-3, Opus | Tier 2 (Review) |
| reviewer | GPT-4 | Claude-3, Opus | Tier 2 (Review) |
| scout | GPT-3.5 | GPT-4, Claude-3 | Tier 2 (Read-only) |

## Tier Definitions

### Tier 0 — Deterministic
- **Agents:** None (scripts only)
- **Models:** None
- **Can do:** Health checks, file operations, git commands
- **Needs approval for:** None

### Tier 1 — Router
- **Agents:** orchestrator
- **Models:** Opus, GPT-4
- **Can do:** Dispatch, routing, status
- **Needs approval for:** Implementation, deploy

### Tier 2 — Read-only
- **Agents:** scout, reviewer, red-team
- **Models:** GPT-4, Claude-3
- **Can do:** Research, review, verify
- **Needs approval for:** Writes, edits, commits

### Tier 3 — Implement
- **Agents:** builder, documenter, implementor, planner
- **Models:** GPT-4, GPT-3.5
- **Can do:** Code changes, workflow steps
- **Needs approval for:** Deploy, merge, branch ops

## Fallback Chains

### orchestrator
```
Opus → GPT-4 → Claude-3 → escalate
```

### builder / implementor / planner
```
GPT-4 → GPT-3.5 → Claude-3 → escalate
```

### reviewer / red-team
```
GPT-4 → Claude-3 → Opus → escalate
```

### scout
```
GPT-3.5 → GPT-4 → Claude-3 → escalate
```

## Fallback Decision Tree

1. **Model error** (rate limit, timeout, provider) → next model in chain
2. **Agent stuck / doom loop** → inject recovery prompt → still stuck → kill + escalate
3. **Agent error / unexpected** → route to reviewer → can't diagnose → escalate
4. **Agent timeout** → wait 2 min → no response → kill + escalate

## Safety Rules

- **No model shared between opposing roles** (implement ≠ review ≠ plan)
- **Critical agents have ≥2 providers** in fallback chain
- **Tier 4 (Sensitive) agents** not yet defined — requires human approval for all actions
- **Do not change production model assignments** until validated by eval data

## Cost Tracking

| Model | Input $/1K tokens | Output $/1K tokens |
|-------|-------------------|--------------------|
| GPT-4 | $0.03 | $0.06 |
| GPT-3.5 | $0.015 | $0.03 |
| Claude-3 | $0.03 | $0.06 |
| Opus | $0.05 | $0.10 |

## Usage Notes

- Default to GPT-4 for most agents (good balance of capability/cost)
- Use Opus only for orchestrator (complex reasoning required)
- Use GPT-3.5 for scout (cost-effective for research)
- Monitor costs via `lazyai-cli doctor` output
