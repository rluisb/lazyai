---
name: worktree-manager
description: "Unified worktree + container workflow. Manages git worktrees (agent parallel, dev ticket-based, manual feature branches) and their container coupling. Lifecycle: create → use → merge → clean."
trigger: /worktree-manager
skill_path: skills/worktree-manager
scripts:
  - name: worktree-lifecycle.sh
    description: "Full worktree lifecycle: create, provision, use, merge, cleanup"
    path: scripts/worktree-lifecycle.sh
  - name: worktree-status.sh
    description: "Inventory of all worktrees with status, age, and container mapping"
    path: scripts/worktree-status.sh
---

# Worktree Manager — Unified Worktree + Container Workflow

## Purpose
Single skill for all worktree operations. Manages three distinct worktree systems and their container relationships:

1. **Dev worktrees** (`dev worktree`) — Creates git worktree + isolated container for ticket-based feature development
2. **Agent worktrees** (`worktree-manager.sh`) — Git worktrees only, for parallel agent execution
3. **Manual worktrees** (`git worktree`) — Long-lived feature branches without containers

**Use when:**
- Starting feature work on a ticket → `dev worktree`
- Running parallel agent tasks → agent worktrees
- Managing long-lived branches → manual worktrees
- Cleaning up stale worktrees → lifecycle cleanup

## Scripts

| Script | Purpose |
|--------|---------|
| `worktree-lifecycle.sh` | Full lifecycle: create, provision, use, merge, cleanup |
| `worktree-status.sh` | Inventory of all worktrees with status, age, container mapping |

Run from skill directory: `./scripts/<script>.sh [command]`

---

## Worktree Systems

### System 1: Dev Worktrees (Worktree + Container)

The **only** system that automatically couples a git worktree with an isolated container.

```bash
# Create worktree + container for a ticket
dev worktree fedora PEN-123

# Provision the container
dev update fedora --worktree-path fedora-worktrees/PEN-123

# Start the service
dev start fedora --worktree-path fedora-worktrees/PEN-123

# Run commands in the container
dev exec fedora-wt-pen-123 -- ./bin/rspec spec/

# Stop when done
dev stop fedora --worktree-path fedora-worktrees/PEN-123
```

**Naming convention:**
- Worktree path: `fedora-worktrees/<ticket>/`
- Container name: `<service>-wt-<ticket-slug>` (e.g., `fedora-wt-pen-123`)
- Branch: auto-created by `dev worktree`

**Safety rules:**
- Always `dev update` before `dev start` on a new worktree
- Never push from worktree containers — merge requires user approval
- Stop worktree containers when not actively using them

### System 2: Agent Worktrees (Git Only, No Container)

Short-lived worktrees for parallel agent execution. No containers — agents use `dev exec` separately if needed.

```bash
# Create agent worktree
./skills/build-mode/scripts/worktree-manager.sh create task-1234567
# → Creates git worktree at .worktrees/task-1234567/ on branch wt/task-1234567

# Merge back to main branch
./skills/build-mode/scripts/worktree-manager.sh merge task-1234567

# Clean up
./skills/build-mode/scripts/worktree-manager.sh clean task-1234567
```

**Naming convention:**
- Worktree path: `.worktrees/<name>/`
- Branch: `wt/<name>` (e.g., `wt/task-1234567`)

**Safety rules:**
- Agents never push from worktrees
- Merge requires user approval
- Clean up after each parallel wave

### System 3: Manual Git Worktrees (Git Only, No Container)

Long-lived feature branches with separate working directories.

```bash
# Create manual worktree
git worktree add -b judge-expansion ../opencode.judge-expansion main

# List all worktrees
git worktree list

# Remove worktree
git worktree remove ../opencode.judge-expansion
```

**Naming convention:**
- Worktree path: `opencode.<topic>/` (sibling to main repo)
- Branch: `<topic>` (e.g., `judge-expansion`)

---

## Lifecycle Commands

```bash
# Full lifecycle for a dev worktree
./scripts/worktree-lifecycle.sh create fedora PEN-123
./scripts/worktree-lifecycle.sh provision fedora PEN-123
./scripts/worktree-lifecycle.sh start fedora PEN-123
./scripts/worktree-lifecycle.sh stop fedora PEN-123
./scripts/worktree-lifecycle.sh merge fedora PEN-123
./scripts/worktree-lifecycle.sh cleanup fedora PEN-123

# Status of all worktrees
./scripts/worktree-status.sh

# Status of specific worktree
./scripts/worktree-status.sh PEN-123
```

---

## Worktree ↔ Container Mapping

| System | Worktree Location | Container | Purpose |
|--------|------------------|-----------|---------|
| `dev worktree` | `fedora-worktrees/<ticket>/` | ✅ `<service>-wt-<ticket-slug>` | Feature dev on a ticket |
| Agent worktrees | `.worktrees/<name>/` | ❌ None | Parallel agent execution |
| Manual worktrees | `opencode.<topic>/` | ❌ None | Long-lived feature branches |

---

## Integration with Other Skills

| Skill | Integration Point |
|-------|-------------------|
| **build-mode** | Uses agent worktrees for parallel execution |
| **battle-bus** | Orchestrates parallel waves with worktree isolation |
| **dev-cli** | Provides `dev worktree`, `dev exec` commands |
| **refresh-dev-containers** | Refreshes worktree containers with latest main |
| **dev-health** | Reports worktree status in health check |

---

## Agent Usage

| Agent | When |
|-------|------|
| **rift-deploy** | Primary owner — manages dev worktrees and containers |
| **wall-builder** | Uses agent worktrees for parallel implementation |
| **loop-driver** | Orchestrates worktree lifecycle in battle-bus waves |

---

## Safety Rules

1. **Never push from worktrees** — merging requires user approval
2. **Never delete worktrees with uncommitted changes** — stash or commit first
3. **Never remove running containers** — stop first
4. **Always provision before starting** — `dev update` before `dev start`
5. **Always clean up after parallel waves** — remove worktrees and branches
6. **Use `--non-interactive` for agent commands** — prevents hanging prompts
