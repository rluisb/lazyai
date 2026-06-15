# Cross-Agent Dispatch Matrix

Every agent has `permission.task: allow` and can dispatch to other agents. Delegation follows role boundaries.

## Dispatch Rules

| From Agent | Can Dispatch To | When | Parameters |
|---|---|---|---|
| **loop-driver** | Any | Routing / escalation | Full dispatch params |
| **loop-driver** | shield-audit | After parallel fork wave barrier resolves | `MODE=judge INPUTS=<pathA>,<pathB>` CONTEXT=<spec> |
| **engine-control** | Any agent in workflow | Step execution | `MODE=<step-mode> THINK=true WORKFLOW_INSTANCE=<iid>` |
| **engine-control** | loop-driver | Escalation | — |
| **loot-hawk** | turbo-crank | Research → spec needed | `MODE=plan` |
| **loot-hawk** | shield-audit | Findings need verification | `MODE=review` |
| **loot-hawk** | loop-driver | Escalation | — |
| **turbo-crank** | loot-hawk | Spec needs more research | `MODE=deep DOMAIN=<topic>` |
| **turbo-crank** | wall-builder | Plan approved → implement | `MODE=standard` (human approval) |
| **turbo-crank** | shield-audit | Pre-implementation review | `MODE=review FOCUS=spec` |
| **turbo-crank** | loop-driver | Escalation | — |
| **wall-builder** | shield-audit | Self-review after impl | `MODE=review FOCUS=spec` |
| **wall-builder** | shield-audit | Pre-PR review request | `MODE=review FOCUS=all` |
| **wall-builder** | shield-audit | Feedback review on PR comments | `MODE=review FOCUS=spec` |
| **wall-builder** | loot-hawk | Needs code exploration | `MODE=deep DOMAIN=<topic>` |
| **wall-builder** | turbo-crank | Spec ambiguous | `MODE=clarify` |
| **wall-builder** | loop-driver | Escalation / blocker | — |
| **shield-audit** | wall-builder | Findings need fixing | `MODE=standard` (human approval) |
| **shield-audit** | wall-builder | Judge verdict recommends fix for loser | `MODE=standard FOCUS=<recommendation>` (human approval) |
| **shield-audit** | wall-builder | PR review findings need fixing (pr-review) | `MODE=standard FOCUS=<finding>` (human approval) |
| **shield-audit** | wall-builder | Feedback review action plan needs fixes | `MODE=standard FOCUS=<action-item>` (human approval) |
| **shield-audit** | loot-hawk | Deeper code exploration | `MODE=exhaustive DOMAIN=<topic>` |
| **shield-audit** | loop-driver | Critical findings | — |
| **rift-deploy** | shield-audit | Pre-deploy verification | `MODE=quick` |
| **rift-deploy** | respawn-crew | Post-deploy health | `MODE=triage SEVERITY=P3` |
| **rift-deploy** | wall-builder | Deploy blocked by code | `MODE=standard` (human approval) |
| **rift-deploy** | loop-driver | Deploy failure | — |
| **respawn-crew** | wall-builder | Incident needs code fix | `MODE=senior` (human approval) |
| **respawn-crew** | rift-deploy | Rollback needed | `MODE=staging` |
| **respawn-crew** | loot-hawk | Code investigation | `MODE=exhaustive DOMAIN=<topic>` |
| **respawn-crew** | shield-audit | Post-incident security | `MODE=security FOCUS=security` |
| **respawn-crew** | loop-driver | P1 beyond threshold | — |
| **turbo-crank** | shield-audit | Architecture fork needs verdict | `MODE=judge INPUTS=<archA>,<archB> CONTEXT=<spec>` |
| **wall-builder** | shield-audit | Implementation fork needs verdict | `MODE=judge INPUTS=<implA>,<implB> CONTEXT=<spec>` |
| **respawn-crew** | shield-audit | Mitigation fork needs verdict | `MODE=judge INPUTS=<mitigationA>,<mitigationB>` |
| **rift-deploy** | shield-audit | Deploy strategy fork needs verdict | `MODE=judge INPUTS=<strategyA>,<strategyB> CONTEXT=<spec>` |
| **loop-driver** | shield-audit | Model fallback tiebreaker | `MODE=judge INPUTS=<modelOutputA>,<modelOutputB> CONTEXT=<spec>` |
| **loot-hawk** | shield-audit | Research synthesis fork needs verdict | `MODE=judge INPUTS=<findingsA>,<findingsB>` |

## Delegation Rules

1. **Human approval required** when delegating from verify→implement or plan→implement
2. **Never delegate across phase boundaries** without loop-driver mediation (e.g., deploy→plan)
3. **All agents can escalate** to loop-driver at any time
4. **Read-only agents** (loot-hawk, shield-audit) never delegate to write agents without approval

## Parallel Execution Protocol

Independent tasks run in parallel waves. Each wave completes before the next begins.

```
## Dispatch Parameters
WAVE: 1
TASKS:
  - agent: wall-builder, mode: fork, text: "Implement auth middleware"
  - agent: wall-builder, mode: fork, text: "Implement payment handler"
BARRIER: impl-sync (count: 2)
```

### Parallel Task Lifecycle

1. **Register**: `session-db.sh ptask <sid> <agent> <task> <wave_id>`
2. **Start**: `session-db.sh ptask-start <ptask-id>`
3. **Complete**: `session-db.sh ptask-done <ptask-id> <result> [output_path]`
4. **Fail**: `session-db.sh ptask-fail <ptask-id> <error>`
5. **Sync**: `task-barrier.sh arrive <barrier-id>` → wait for all arrivals

### Barrier Synchronization

```bash
# Create barrier expecting 3 parallel tasks
task-barrier.sh create "wave-1-sync" 3

# Each parallel task arrives when done
task-barrier.sh arrive "wave-1-sync"

# Coordinator waits for all
task-barrier.sh wait "wave-1-sync" 120
```

### Exclusive Locks

```bash
# Acquire lock before writing to shared spec
task-lock.sh acquire "spec-write" "turbo-crank"

# Release when done
task-lock.sh release "spec-write"

# Try with timeout (returns 0 on success, 1 on timeout)
task-lock.sh try "deploy-lock" "rift-deploy" 30
```

### Parallel Safety Rules

1. **Different file paths** — parallel writers must target non-overlapping files
2. **Lock shared resources** — use `task-lock.sh` for any shared state
3. **Wave isolation** — wave N+1 starts only after wave N barrier resolves
4. **No circular dispatch** — agent A → B → A creates infinite loop; escalate to loop-driver
5. **Max parallelism** — limit concurrent forks to 4 per wave to avoid model rate limits
