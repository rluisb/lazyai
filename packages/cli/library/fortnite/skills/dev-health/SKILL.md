---
name: dev-health
description: Environment health check — colima, dev CLI services, worktrees, containers, disk usage. Single command to assess the full dev environment state.
trigger: /dev-health
skill_path: skills/dev-health
scripts:
  - name: dev-health-check.sh
    description: Full environment health assessment
    path: scripts/dev-health-check.sh
  - name: dev-cleanup.sh
    description: Clean up stale worktrees, stopped containers, orphaned branches
    path: scripts/dev-cleanup.sh
---

# Dev Health — Environment Health Check

## Purpose
Single-command assessment of the entire development environment: Docker runtime, dev CLI services, worktrees, containers, and disk usage. Produces a structured health report with cleanup recommendations.

**Use when:**
- Starting a new session — "is everything running?"
- Before a deploy — "is the environment clean?"
- After an incident — "what's broken?"
- Periodic maintenance — "what can I clean up?"

## Scripts

| Script | Purpose |
|--------|---------|
| `dev-health-check.sh` | Full environment health assessment |
| `dev-cleanup.sh` | Clean up stale worktrees, stopped containers, orphaned branches |

Run from skill directory: `./scripts/<script>.sh [command]`

---

## Health Check Levels

| Level | What it checks | Command |
|-------|---------------|---------|
| **quick** | Colima status + running services | `./scripts/dev-health-check.sh quick` |
| **full** | Everything + worktrees + containers + disk | `./scripts/dev-health-check.sh full` |
| **deep** | Full + container logs + service health endpoints | `./scripts/dev-health-check.sh deep` |

---

## Cleanup Operations

```bash
# Preview what would be cleaned (dry run)
./scripts/dev-cleanup.sh --dry-run

# Remove stopped containers older than 7 days
./scripts/dev-cleanup.sh containers --age 7

# Remove stale agent worktrees (merged branches)
./scripts/dev-cleanup.sh worktrees --stale

# Remove orphaned branches (no worktree, no remote)
./scripts/dev-cleanup.sh branches --orphaned

# Full cleanup (all of the above, with confirmation)
./scripts/dev-cleanup.sh all
```

**Safety rules:**
- Never remove running containers
- Never remove worktrees with uncommitted changes
- Never remove branches with unpushed commits
- Always dry-run first
- Always confirm before destructive operations

---

## Integration with Other Skills

| Skill | Integration Point |
|-------|-------------------|
| **refresh-dev-containers** | Pre-check: verify environment is healthy before refresh |
| **colima** | Runtime check: colima-health.sh is a subset of dev-health |
| **dev-cli** | Service status: dev list output is part of health report |
| **rift-deploy** | Pre-deploy gate: environment must be healthy before deploy |
| **respawn-crew** | Incident triage: health check is first step in diagnosis |

---

## Agent Usage

| Agent | When |
|-------|------|
| **rift-deploy** | Pre-deploy health check, environment setup verification |
| **respawn-crew** | Incident triage — first step in P1/P2 diagnosis |
| **wall-builder** | Pre-implementation — verify dev environment is ready |
| **loop-driver** | Session start — include in session-reset.sh flow |
