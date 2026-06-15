---
name: refresh-dev-containers
description: "Refresh local branch and dev container service safely. Handles git stash/pull, dev update, Node version management, and Fedora precompiled-assets mode. Use when syncing local dev environment with main."
trigger: /refresh-dev-containers
phase: implement
---

# /refresh-dev-containers — Safe Dev Environment Refresh

Refreshes local code + dev container state for a given service. Handles git operations, container updates, and service-specific paths.

## When to Use

- Syncing local service with `origin/main`
- Updating dev container after pulling new dependencies
- Refreshing Fedora (with confirmation gate)
- Setting up mono-frontend with correct Node version

## Inputs

- `service` (required): service name used by `dev` commands (e.g., `fedora`, `creator-checkout`, `mono-frontend`)

## Workflow

### 0. Running-Service Guard (Required)

Before any operations, check if the target service is running:

```bash
dev list --json 2>/dev/null | grep -q '"status":"running"'
```

- If running and likely in active use → **ask user** whether to skip refresh
- For `fedora` → **hard gate**: never refresh while actively used without explicit confirmation

### 1. Check Branch + Working Tree

```bash
git status -sb
```

- If on merged PR branch → `git checkout main` first
- If dirty → `git stash push -u -m "refresh-dev-containers: <service>"`
- If clean → continue

### 2. Update Branch from Main

```bash
git pull origin main
```

- If merge conflicts → **STOP**, ask user to resolve
- Never auto-resolve conflicts

### 3. Restore Stashed Changes

```bash
git stash list
git stash pop  # for the stash created in step 1
```

- If stash pop conflicts → **STOP**, ask user to resolve

### 4. Service-Specific Refresh

#### mono-frontend (Not Containerized)

```bash
# Skip dev list and dev update
nvm install 22 && nvm use 22  # Preferred Node version
yarn install
```

#### Containerized Services (fedora, creator-checkout, etc.)

```bash
# Fedora confirmation gate (required)
# Ask user: "Run full Fedora refresh? It is slow."

dev --non-interactive list
# If Dev CLI update prompt → respond 'y'
# Verify service exists and is cloned

dev --non-interactive update <service>
```

**Fedora Precompiled-Assets Mode** (optional, backend-focused):

```bash
dev update fedora --precompile-assets    # Precompile assets once
dev start fedora --use-precompiled-assets  # Start without webpack
```

### 5. Report Outcome

Use the output format below.

## Output Format

```markdown
## Refresh Dev Containers Report
- Service: <service>
- Branch status: <clean | stashed changes>
- Pull from main: <success | conflict>
- Stash restore: <not needed | success | conflict>
- Service found in dev list: <yes | no | n/a (mono-frontend)>
- Refresh command: <dev update <service> | yarn install>
- Refresh result: <success | failed>
- Dev CLI auto-update during refresh: <yes | no>
- Notes: <errors or next action>
```

## Safety Rules

1. **Never drop user changes** — always stash before pull when dirty
2. **Never auto-resolve conflicts** — git or stash conflicts require user
3. **Never interrupt running service** — ask before refreshing active services
4. **Always use `--non-interactive`** — prevents hanging prompts
5. **Only update cloned services** — verify via `dev list` first
6. **Fedora requires explicit confirmation** — full refresh is slow
7. **mono-frontend skips dev update** — uses `yarn install` instead
8. **Node 22 preferred for mono-frontend** — use nvm to switch

## Scripts

| Script | Purpose |
|--------|---------|
| `scripts/refresh-service.sh` | Automated refresh with safety guards |

## Integration with Other Skills

- **dev-cli** — Uses `dev` commands for container management
- **colima** — Requires Docker runtime (colima must be running)
- **build-mode** — Can be called before implementation to sync environment

## Agent Ownership

Primary: **rift-deploy** (dev environment setup)
Secondary: **wall-builder** (pre-implementation sync)
