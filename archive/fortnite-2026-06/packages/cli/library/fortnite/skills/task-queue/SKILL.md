---
name: task-queue
description: SQLite-backed multi-agent task queue with atomic claiming, dead-letter queue, zombie sweeper, and task-scoped chat. Enables parallel execution with guaranteed no duplicate claims.
trigger: /task-queue
triggers:
  - "task queue"
  - "enqueue task"
  - "claim task"
  - "task status"
  - "task chat"
skill_path: skills/task-queue
scripts:
  - name: task-queue.sh
    description: Task queue CLI — init, add, claim, list, claims, complete, fail, sweep, dlq, msg-send, msg-poll
    path: scripts/task-queue.sh
database_schema:
  tables: [task_queue, task_claims, task_dlq, task_messages]
  foreign_keys: [task_id → task_queue, session_id → sessions]
---

## Quick Reference

| | |
|---|---|
| **Use when** | Multi-agent parallel execution, atomic task claiming |
| **Do not use when** | Single-agent sequential tasks |
| **Primary agent** | All agents |
| **Runtime risk** | Medium — concurrency, zombie tasks |
| **Outputs** | Task queue state, claim records, DLQ entries |
| **Validation** | Atomic claiming, barrier sync |
| **Deep mode trigger** | `/task-queue` or parallel wave coordination |

# Task Queue — The Squad Coordinator

**Script**: `scripts/task-queue.sh`
**Database**: `.specify/session.db` — tables `task_queue`, `task_claims`, `task_dlq`, `task_messages`

You coordinate parallel task execution across multiple agents. The task queue provides atomic claiming, dead-letter queue (DLQ), zombie sweeper, and task-scoped chat for safe multi-agent collaboration.

## Purpose

SQLite-backed task queue for multi-agent parallel execution with:
- **Topic/subject routing** — Agents claim tasks by topic (e.g., "auth", "payment", "verify")
- **Multi-agent claiming** — `max_agents` controls how many agents can work on the same task
- **DLQ/state preservation** — Failed tasks move to DLQ with full context for triage
- **Zombie sweeper** — Visibility timeout via `sweep` removes stale claims
- **Task-scoped chat** — Agents collaborate via `msg-send`/`msg-poll` without global message bus

## CLI Commands

```bash
# Enqueue a task
task-queue.sh add <sid> <topic> <task> [max] [dedupe_key]

# Atomically claim an open task
task-queue.sh claim <topic> <agent>

# List tasks (optionally filter by topic)
task-queue.sh list [topic]

# List claims (optionally filter by task_id)
task-queue.sh claims [task_id]

# Mark task as completed
task-queue.sh complete <task_id>

# Mark task as failed and send to DLQ
task-queue.sh fail <task_id> <agent> [error_message] [context_dump]

# Remove stale claims (zombie sweep)
task-queue.sh sweep <timeout_seconds>

# List dead-letter queue entries
task-queue.sh dlq

# Send a message on a task
task-queue.sh msg-send <task_id> <from_agent> <body> [to_agent]

# Poll messages for a task
task-queue.sh msg-poll <task_id> [last_seen_id]
```

## Safe Operational Rules

1. **No duplicate claims by same agent** — The claim CTE prevents an agent from claiming the same task twice
2. **Claims must be completed or failed** — Never leave claims dangling; use `sweep` for stale claims
3. **Use task chat for handoffs** — When a task requires collaboration, use `msg-send`/`msg-poll` instead of global message bus
4. **DLQ is for triage** — Failed tasks preserve full context (error, agent, timestamp) for root-cause analysis
5. **Zombie sweep before claiming** — Run `sweep` periodically to clean up stale claims from crashed agents

## Example Flows

### Queue Producer (engine-control / loop-driver)

```bash
# Enqueue a workflow step
task-queue.sh add ses_abc123 "auth" "Implement login middleware" 2

# Monitor queue health
task-queue.sh list auth
task-queue.sh claims
```

### Agent Claimer (wall-builder / shield-audit)

```bash
# Claim a task by topic
TASK_ID=$(task-queue.sh claim "auth" "wall-builder" | grep -oP 'task \K\d+')

# Do work...

# Complete the task
task-queue.sh complete "$TASK_ID"
```

### Collaboration / Chat

```bash
# Send a message to another agent about this task
task-queue.sh msg-send 42 "wall-builder" "I'm starting on the HTML" "shield-audit"

# Poll for messages
task-queue.sh msg-poll 42 0
```

### Fail → DLQ

```bash
# Mark task as failed with context
task-queue.sh fail 42 "wall-builder" "Database connection timeout" "$(cat error.log)"

# Triage DLQ
task-queue.sh dlq
```

## Database Schema

```sql
task_queue: id, session_id, topic, task, status, max_agents, dedupe_key, created_at
task_claims: task_id, agent, claimed_at
task_dlq: id, task_id, failed_agent, error_message, context_dump, failed_at
task_messages: id, task_id, from_agent, to_agent, body, created_at
```

The `task_queue` table includes a partial unique index `idx_task_queue_active_dedupe` on `(dedupe_key)` where status is active/open/claimed. This ensures idempotent enqueueing for active tasks while allowing completed tasks to be re-enqueued.

## Integration with Workflow Engine

The task queue integrates with `workflow-engine` via:
- `skills/workflow-engine/scripts/workflow-exec.sh enqueue-step <instance-id>` — Queues the current workflow step by target agent/topic
- `skills/workflow-engine/scripts/workflow-exec.sh queue-status` — Shows queue state for monitoring

Downstream agents claim by topic using `task-queue.sh claim <topic> <agent>`, use task chat for blockers/handoffs, and complete/fail with context.

Note: `enqueue-step` and `queue-status` are implemented by `workflow-exec.sh`, not by the task-queue CLI itself.

## Dedupe Key Format

For deterministic workflow deduplication, use the format `workflow:<wiid>:step:<step_order>`. When a task with an active status (open/claimed) and matching dedupe_key is enqueued, the existing active task is returned and no duplicate is created. Completed tasks are not considered active duplicates and can be re-enqueued.

## Configuration

- **DB_PATH override**: Set `DB_PATH` environment variable to use a custom database path for testing or temporary queues.
- **Numeric validation**: The `max_agents`, `task_id`, and sweep timeout parameters are validated as positive integers.

## Rules

- Always `sweep` before claiming to avoid zombie claims
- Use `msg-send` for task-specific collaboration
- Complete or fail every claimed task
- DLQ entries require human triage before re-queueing
- Topic names should be lowercase, hyphenated, and descriptive (e.g., "auth-middleware", "payment-handler")
