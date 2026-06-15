---
name: dev-cli
description: "Teachable dev tool — start/stop services, shell, exec, rebase, review, worktree, webhook forwarding. Use when managing local development environments, running tests in containers, or working with git worktrees."
trigger: /dev-cli
---

# /dev-cli — Teachable Development Tool

Teachable's internal CLI for managing local development environments. Handles service lifecycle, container execution, git operations, and AI-powered code review.

## When to Use

- Starting/stopping local development services
- Running tests inside service containers
- Opening shells in running containers
- Rebasing branches without checking out main
- AI-powered code review of staged changes
- Creating git worktrees with isolated containers
- Forwarding webhooks to local services
- Setting up new developer environments

## Quick Reference

### Service Management
```bash
dev list                          # List all services and status
dev list --json                   # JSON output for scripting
dev start fedora                  # Start a service (hot-reload enabled)
dev start fedora auth-service     # Start multiple services
dev stop fedora                   # Stop a service
dev stop --all                    # Stop all services
dev start fedora --sleep          # Start in sleep mode (for exec)
dev update fedora                 # Update service dependencies
dev uninstall fedora              # Remove service and resources
```

### Container Execution
```bash
dev exec fedora -- ./bin/rspec spec/models/user_spec.rb
dev exec checkout-service -- go test ./...
dev exec fedora --update-dependencies -- bundle install
dev exec fedora --worktree-path /path/to/worktree -- ./bin/rspec spec/...
```

### Interactive Shell
```bash
dev shell fedora                  # Open shell in container
dev shell fedora --worktree-path /path/to/worktree
```

### Git Operations
```bash
dev rebase fedora                 # Rebase onto origin/main
dev rebase fedora --reset         # Reset to origin/main (collapse commits)
dev rebase fedora --reset --soft  # Reset with changes staged
dev rebase fedora --target-branch develop
dev lock fedora                   # Lock main branch (block merges)
dev unlock fedora                 # Unlock main branch
```

### Code Review
```bash
dev review fedora                 # AI review of staged changes (Copilot CLI)
```

### Worktrees
```bash
dev worktree fedora PEN-123       # Create worktree + isolated container
dev start fedora --worktree-path /path/to/fedora-worktrees/PEN-123
dev exec fedora-wt-pen-123 -- ./bin/rspec spec/...
dev stop fedora --worktree-path /path/to/fedora-worktrees/PEN-123
```

### Webhooks
```bash
dev webhook                       # Forward SQS-backed webhooks to local
```

### Setup
```bash
dev setup --gh-token TOKEN --gh-username USER --project-name v0
```

## Key Flags

| Flag | Description |
|------|-------------|
| `--non-interactive` | Disable prompts (required for AI agents/CI) |
| `--worktree-path` | Target a worktree container |
| `--sleep` | Start container in sleep mode (for exec) |
| `--update-dependencies` | Run container_update.sh before exec |
| `--use-precompiled-assets` | Serve built assets instead of watchers |

## Safety Rules

1. **Always use `--non-interactive` in agent contexts** — prevents hanging prompts
2. **Never run `dev rebase --reset` without human approval** — collapses commits
3. **Verify service is running before `dev exec`** — use `dev list` first
4. **Use `dev stop --all` carefully** — stops all running services
5. **Worktree containers need `dev update` before `dev start`** — provision first

## Scripts

| Script | Purpose |
|--------|---------|
| `scripts/dev-status.sh` | Quick service status check |
| `scripts/dev-exec-test.sh` | Run tests in container with dependency update |
| `scripts/dev-rebase-safe.sh` | Safe rebase with stash protection |

## Integration with Other Skills

- **rift-deploy** — Uses dev for dev environment setup and service management
- **wall-builder** — Uses dev exec for running tests during implementation
- **build-mode** — Uses dev exec for quality gates (tests, lint, build)
- **zero-point** — Uses dev exec for verification test runs
- **refresh-dev-containers** — Workflow skill that wraps dev CLI for safe environment refresh

## Agent Ownership

Primary: **rift-deploy** (dev environment setup)
Secondary: **wall-builder** (test execution), **respawn-crew** (service health checks)
