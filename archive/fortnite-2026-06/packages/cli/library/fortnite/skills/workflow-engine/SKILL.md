---
name: workflow-engine
description: Runtime workflow orchestration. Reads YAML configs from .opencode/workflows/, syncs to SQLite, executes phases with human gates and feedforward context.
trigger: /workflow-engine
triggers:
  - "run workflow"
  - "execute workflow"
  - "start workflow"
  - "define team"
  - "create workflow"
  - "orchestrate agents"
skill_path: skills/workflow-engine
scripts:
  - name: workflow-run.sh
    description: Workflow orchestrator — sync YAML configs, execute phases, enforce gates
    path: scripts/workflow-run.sh
database_schema:
  tables: [workflows, workflow_runs, workflow_phases]
  foreign_keys: [session_id → sessions, workflow_id → workflows, run_id → workflow_runs]
---

## Quick Reference

| | |
|---|---|
| **Use when** | Runtime workflow orchestration, YAML config execution |
| **Do not use when** | Direct agent dispatch, simple single-step tasks |
| **Primary agent** | engine-control |
| **Runtime risk** | Medium — step sequencing, gate enforcement |
| **Outputs** | Workflow instances, phase metrics, gate results |
| **Validation** | Schema validation, phase completion |
| **Deep mode trigger** | `/workflow-engine` or complex workflow definition |

# Workflow Engine — The Drop Zone Commander

**Script**: `scripts/workflow-run.sh`
**Configs**: `.opencode/workflows/*.yaml`
**Database**: `.specify/session.db` — tables `workflows`, `workflow_runs`, `workflow_phases`

You coordinate multi-agent workflows at runtime. You read YAML workflow configs, sync them to SQLite, and execute phases with human gates and feedforward context.

## Available Workflows

| Workflow | Trigger | Phases | Modes |
|----------|---------|--------|-------|
| **rpi** | `/rpi` | clarify → research → plan → implement → verify | simple, complex, fast |
| **bugfix** | `/bugfix` | diagnose → fix → verify | investigate, known, fast |
| **hotfix** | `/hotfix` | triage → fix → verify | emergency (no gates) |
| **refactor** | `/refactor` | plan → implement → verify | standard, fast |
| **spike** | `/spike` | explore → document → validate | standard, deep |

## Commands

```bash
# Sync YAML configs to SQLite
workflow-run.sh sync

# List available workflows
workflow-run.sh list

# Run a workflow
workflow-run.sh run rpi --mode complex "Add OAuth support"
workflow-run.sh run bugfix --mode known "Fix null pointer in auth"
workflow-run.sh run hotfix "Production timeout on /api/users"
workflow-run.sh run refactor --dry-run --mode standard "Clean up auth module"

# Check workflow run status
workflow-run.sh status
workflow-run.sh status wfr_abc123
```

## Workflow Config Format

YAML configs live in `.opencode/workflows/`. Each config defines:

```yaml
name: rpi
trigger: /rpi
description: Full feature workflow
modes:
  simple:
    phases: [plan, implement, verify]
    skip_gates: false
  complex:
    phases: [clarify, research, plan, implement, verify]
    skip_gates: false
default_mode: complex
phases:
  - name: clarify
    agent: turbo-crank
    skill: storm-scout
    mode: clarify
    feedforward: "Task: {GOAL}. Resolve critical unknowns."
    gate: human_confirms
    gate_prompt: "Clarification complete. Proceed?"
```

### Phase Fields

| Field | Purpose | Example |
|-------|---------|---------|
| `name` | Phase identifier | `clarify`, `research`, `plan` |
| `agent` | Target agent | `turbo-crank`, `wall-builder`, `shield-audit` |
| `skill` | Skill to load | `storm-scout`, `build-mode`, `zero-point` |
| `mode` | Agent/skill mode | `clarify`, `senior`, `review` |
| `feedforward` | Context from previous phase | `"Spec: {SPEC}. Task: {TASK}."` |
| `gate` | Human approval gate | `human_confirms`, `human_approves_spec`, `null` |
| `gate_prompt` | Prompt shown at gate | `"Spec ready. Approve?"` |

### Variable Interpolation

Feedforward strings support `{VARIABLE}` placeholders and ternary expressions:

- `{GOAL}` — The workflow goal passed at run time
- `{COMPLEXITY}` — The mode (simple/complex/etc.)
- `${COMPLEXITY == 'complex' ? 'senior' : 'standard'}` — Ternary mode selection

## Integration with Other Skills

- **battle-bus**: Workflow configs replace inline templates; battle-bus references YAML
- **truth-chain**: Each phase start/complete logged to immutable ledger
- **session-db**: Workflow runs tracked in SQLite with phase-level detail
- **storm-scout**: Used by clarify, research, and plan phases
- **build-mode**: Used by implement phases
- **zero-point**: Used by verify phases
- **reboot-van**: Used by bugfix diagnose phase
- **task-queue**: Enqueue workflow steps by topic/agent using `skills/workflow-engine/scripts/workflow-exec.sh enqueue-step <instance-id>`; agents claim by topic and use task-scoped chat; inspect queue status using `skills/workflow-engine/scripts/workflow-exec.sh queue-status`; duplicate enqueue prevention is implemented via deterministic dedupe keys using the format `workflow:<wiid>:step:<step_order>`

  After enqueue, agents use these task-queue commands:
  - `scripts/task-queue.sh claim <topic> <agent>` — Claim a task from the queue
  - `scripts/task-queue.sh msg-send <task_id> <from_agent> "<body>" [to_agent]` — Send a task-scoped message
  - `scripts/task-queue.sh msg-poll <task_id> [last_seen_id]` — Poll for task-scoped messages
  - `scripts/task-queue.sh complete <task_id>` — Mark task as complete
  - `scripts/task-queue.sh fail <task_id> <agent> "<error>" "<context>"` — Mark task as failed with error context

## Database Schema

```sql
workflows: id, name, description, trigger_cmd, config_json, version, created_at, updated_at
workflow_runs: id, workflow_id, session_id, mode, status, current_phase, phases_completed, phases_total, started_at, completed_at, feedforward_context
workflow_phases: id, run_id, phase_name, agent, skill, mode, status, input_context, output, gate_passed, started_at, completed_at, ledger_seq
```

## Rules

- Never skip human gates unless `--skip-gates` flag is used
- Always log phase start/complete to truth-chain ledger
- Feedforward context flows from phase N to phase N+1
- Workflow configs are source of truth — sync to SQLite before running
- Hotfix workflows have no gates by design (emergency path)
