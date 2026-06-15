---
name: parallel-execution
description: Execute independent sub-tasks concurrently.
trigger: /parallel-execution
phase: implement
preset: full
---

# Parallel Execution

## Purpose

Enable safe parallel task execution by identifying independent work units that can run concurrently without conflicts. Reduces wall-clock time on large features.

## The Wave Model

Group tasks into waves based on dependencies. Tasks within a wave run in parallel; waves run sequentially.

```
Wave 0 (No dependencies):
  ┌─────────────┐  ┌─────────────┐
  │ Task A      │  │ Task B      │  ← Run in parallel
  └─────────────┘  └─────────────┘
         │                │
         └────────┬───────┘
                  ▼
Wave 1 (Depends on Wave 0):
  ┌─────────────────────────┐
  │ Task C (needs A + B)    │  ← Must wait for Wave 0
  └─────────────────────────┘
                  │
                  ▼
Wave 2 (Depends on Wave 1):
  ┌─────────────┐  ┌─────────────┐
  │ Task D      │  │ Task E      │  ← Parallel again
  └─────────────┘  └─────────────┘
```

## Dependency Analysis

### File Touch-Map

Before parallelizing, create a file touch-map — which files each task modifies:

```
Task A: src/api/handlers.ts, src/api/routes.ts
Task B: src/db/queries.ts, src/db/schema.sql
Task C: src/api/handlers.ts, src/db/queries.ts   ← conflicts with A and B
Task D: src/services/user.ts
Task E: src/services/order.ts
```

### Conflict Detection

Two tasks **conflict** if they modify the same file:
- Task A and Task C both touch `src/api/handlers.ts` → **CONFLICT** — different waves
- Task A and Task B have no overlap → **SAFE** to parallelize in same wave

### Wave Assignment Rules

1. Tasks with no file conflicts can be in the same wave
2. If Task X depends on Task Y's output, X must be in a later wave
3. Minimize total waves while respecting all dependencies

## Execution by Orchestrator

```
Wave 0: dispatch agent(A) + agent(B) concurrently → wait for both
Wave 1: dispatch agent(C) → wait for completion
Wave 2: dispatch agent(D) + agent(E) concurrently → wait for both
```

## Safety Guarantees

| Guarantee | How Enforced |
|-----------|--------------|
| No file conflicts | File touch-map analysis before dispatch |
| No race conditions | Each agent works in isolated worktree or branch |
| Deterministic merge | Waves complete fully before next wave starts |
| Rollback safety | Each wave's commits are atomic |
| Approval gates | User approves wave plan before execution |

## Worktree Isolation

Each parallel task gets its own worktree or branch:

```
.worktrees/
├── feature-A-{task-hash}/   ← Task A workspace
├── feature-B-{task-hash}/   ← Task B workspace
└── ...
```

**CRITICAL**: Agents never push from worktrees. Merging and remote operations require explicit user approval.

## Merge Strategy

1. Complete Wave 0 → user merges all wave worktrees to main
2. Rebase Wave 1 worktrees on updated main
3. Complete Wave 1 → user merges
4. Repeat until all waves complete

## Example

```
Feature: Add user reviews to products

Wave 0 (Independent):
├── Task A: Add review data model + migrations
│           Files: src/models/review.ts, migrations/
└── Task B: Add product rating calculation
            Files: src/services/rating.ts

Wave 1 (Depends on A + B):
└── Task C: Add review API endpoints
            Files: src/api/reviews.ts, src/api/routes.ts

Wave 2 (Depends on C):
├── Task D: Add review UI components
│           Files: src/ui/review_form.tsx
└── Task E: Add review notification emails
            Files: src/services/notification.ts
```

## Anti-Patterns

- Parallelizing tasks that share mutable files
- Starting Wave 1 before Wave 0 fully completes
- Forgetting to rebase on updated main between waves
- Multiple agents working in the same worktree
- Agents pushing or merging without user approval

## Output Format

```markdown
## Parallel Execution Report
- **Waves completed:** N/N
- **Files touched:** [list]
- **Conflicts detected:** [none | list]
- **Merge status:** [clean | required manual resolution]
```

## Integration

- **Planner agent**: Creates file touch-map and wave assignments during planning
- **Builder agent**: Executes tasks within assigned worktrees
- **Reviewer agent**: Reviews each wave's output before merge approval

## MCP Tool Integration

The parallel-execution skill uses the orchestrator MCP tools to dispatch and monitor wave execution.

### Tool Usage Map

| Operation | Tool | Usage |
|-----------|-----|-------|
| Wave dispatch | `enqueue_job` | Enqueue each task in a wave as a background job |
| Wait for completion | `get_job` | Poll job status until all tasks in a wave are completed |
| Monitor progress | `subscribe_run` | Subscribe to real-time SSE events for all running jobs |
| Handle failures | `retry_step` | Retry a failed task within the wave (if retries remain) |
| Handle failures | `escalate_step` | Escalate a persistently failing task to a different agent |

### Wave Coordination Protocol

#### Step 1: Dispatch Wave 0 Tasks

```
// Enqueue Task A
enqueue_job({ jobType: "parallel-task", payload: { taskId: "A", wave: 0 } })
→ returns jobId: "job-A"

// Enqueue Task B
enqueue_job({ jobType: "parallel-task", payload: { taskId: "B", wave: 0 } })
→ returns jobId: "job-B"
```

#### Step 2: Wait for Wave 0 Completion

```
// Poll until both jobs complete
get_job({ jobId: "job-A" })  // status: "completed"
get_job({ jobId: "job-B" })  // status: "completed"
```

#### Step 3: Subscribe to Progress (optional real-time monitoring)

```
subscribe_run({ runId: "wave-0-run" })
// Receives SSE events: job-started, job-progress, job-completed, job-failed
```

#### Step 4: Handle Failures

```
// If Task A fails:
retry_step({ runId: "wave-0-run", kind: "chain", stepId: "task-A", reason: "transient-timeout" })

// If retries exhausted:
escalate_step({ runId: "wave-0-run", kind: "chain", stepId: "task-A", targetAgent: "implement-v2", reason: "previous attempts failed" })
```

#### Step 5: Proceed to Next Wave

Only after all tasks in current wave are `completed`:
```
// Wave 1 dispatch
enqueue_job({ jobType: "parallel-task", payload: { taskId: "C", wave: 1 } })
```

### Concrete Example: 2-Wave Parallel Execution

```
// WAVE 0: Two independent tasks
enqueue_job({ jobType: "task", payload: { taskId: "build-model", files: ["src/models/review.ts"] } })
enqueue_job({ jobType: "task", payload: { taskId: "build-rating-svc", files: ["src/services/rating.ts"] } })

// Wait for both
get_job({ jobId: "job-model" })  → "completed"
get_job({ jobId: "job-rating" }) → "completed"

// WAVE 1: One task depends on Wave 0 outputs
enqueue_job({ jobType: "task", payload: { taskId: "build-api", dependsOn: ["build-model", "build-rating"], files: ["src/api/reviews.ts"] } })
get_job({ jobId: "job-api" })  → "completed"

// WAVE 2: Two independent tasks
enqueue_job({ jobType: "task", payload: { taskId: "build-ui", files: ["src/ui/review_form.tsx"] } })
enqueue_job({ jobType: "task", payload: { taskId: "build-email", files: ["src/services/notification.ts"] } })

// Wait for Wave 2
get_job({ jobId: "job-ui" })    → "completed"
get_job({ jobId: "job-email" }) → "completed"

// All waves complete
```

### Error Handling Flow

```
Task fails → check retry_limit
  ├── retries_remain: retry_step({ reason: "transient-error" })
  └── retries_exhausted:
        ├── escalate_step({ targetAgent: "alternative-agent" })
        └── if no alternative: handoff({ summary: "wave-N task taskId failed after N retries" })
```
