# Hooks

## What hooks do

Hooks are guardrail or automation scripts that run at tool lifecycle events (for example, command execution, session stop, or commit/workflow boundaries).
They can block an action, emit an advisory warning, or trigger maintenance work.
Support is adapter-specific and varies by tool.

For the lifecycle vocabulary and capability matrix, see [Hook Lifecycle & Capability Matrix](../concepts/hook-lifecycle.md).
For the neutral lifecycle catalog, see [Hook Catalog](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/hooks/catalog.md).

---

## Hook Classification

Every shipped hook is classified by lifecycle event, purpose, and behavioral
properties. The classification is grounded in the hook policy files under
`packages/cli/library/hooks/` and `packages/cli/library/canonical/hooks/`.

| Hook | Lifecycle | Purpose | blocks_actions | requires_human_approval | captures_evidence | surfaces |
|---|---|---|---|---|---|---|
| `pre-commit` | pre-commit | Enforce token-rent budget before local commits (`.githooks/pre-commit`). RPI gate attestation enforced via CI (`.github/workflows/rpi-gate-check.yml`) and `.husky/pre-commit` if configured. | true | false | true | git (`.githooks/pre-commit`, `.husky/pre-commit`) |
| `rpi-gate-check.yml` | CI/CD (pull_request, push) | Enforce RPI process gates on non-trivial PRs and main pushes | true | false | true | GitHub Actions |
| `caveman-memory-promotion` | on_compaction, after_agent | Detect reusable knowledge in caveman summaries and route to memory-promotion review | false | true | false | opencode (plugin) |
| `startup-self-heal` | before_agent | Run scoped health checks and regenerate CLI artifacts on drift | false | false | true | opencode (plugin) |
| `block-destructive-shell` | before_tool | Block destructive shell commands (rm -rf /, mkfs, dd, shutdown, etc.) | true | false | false | opencode, claude, copilot, antigravity, omp, pi (extension), kiro |
| `objective-workflow-gate` | after_model | Require verification evidence or explicit blocked reason in completion claims | true | false | true | opencode, claude, copilot, antigravity |
| `session-start` | before_agent | Session bootstrap reminder policy (objective, handoff, verification seam) | false | false | false | instruction_only (all tools) |

## Shipped hooks

Each shipped hook is defined in `packages/cli/library/hooks/`, `packages/cli/library/canonical/hooks/`, or the OpenCode runtime plugin.

### `block-destructive-shell`

**Policy source:** `packages/cli/library/hooks/block-destructive-shell.md`

- **Purpose:** Block obvious destructive commands (for example `rm -rf /`, `mkfs`, `dd if=/dev/zero ...`, `shutdown`, `reboot`).
- **Trigger event:**
  - Claude/OpenCode/Copilot/Antigravity: shell-tool pre-execution.
  - Kiro: `PreToolUse` with `matcher: "shell"` in native `.kiro/hooks/block-destructive-shell.json`.
  - Pi/OMP: extension/tool-call hook points for `bash`/`tool_call`.
- **Runtime behavior:**
  - OpenCode: plugin `event.type === "tool.execute.before"` in `.opencode/plugins/vibe-lab-hooks.js`.
  - Claude/Copilot/Antigravity: runtime JSON/stdin shell command descriptors in generated hook scripts.
  - Kiro: command hook invokes `.kiro/hooks/block-destructive-shell.sh`; exit code `2` blocks the shell tool.
  - OMP/Pi: TypeScript hook factories in `.omp/hooks/pre/` and `.pi/extensions/` match destructive patterns.

### `objective-workflow-gate`

**Policy source:** `packages/cli/library/hooks/objective-workflow-gate.md`

- **Purpose:** Require completion claims in stop/session-end output to include machine-observable verification/evidence or an explicit blocked reason.
- **Trigger event:**
  - OpenCode: `session.idle` (with stored `lastAssistantMessage` behavior).
  - Claude/Copilot: stop-style event (`Stop`/`agentStop`).
- **Runtime behavior:**
  - OpenCode and Claude/Copilot scripts detect completion keywords (e.g. `Done`, `Complete`, `Implemented`, `Changed`, `Shipped`) and block only when evidence lines such as `Verification:`/`Tests:`/`Checks:`/`Smoke:`/`Evidence:` are missing.
  - Blocked when malformed event payload is observed at fail-closed points.

### `caveman-memory-promotion`

**Policy source:** `packages/cli/library/hooks/caveman-memory-promotion.md`

- **Purpose:** Detect caveman summaries that look like reusable knowledge and propose a review workflow to `memory-promotion`.
- **Trigger event:**
  - OpenCode plugin event `experimental.session.compacting` and `session.idle`.
  - Manual workflow after caveman/diagnostic/triage/handoff summaries.
- **Runtime behavior:** advisory non-blocking prints in OpenCode only; no current runtime hooks for Claude Code, Copilot, Antigravity, Pi, OMP, or Kiro.

### `startup-self-heal`

**Policy source:** `packages/cli/library/hooks/startup-self-heal.md`

- **Purpose:** Run scoped health checks and regenerate CLI artifacts when drift/missing files are detected.
- **Trigger event:**
  - OpenCode plugin session startup event `session.created`.
- **Runtime behavior:**
  - OpenCode plugin executes `bin/startup-self-heal --cli opencode --event session.created --quiet`.
  - Advisory only; non-blocking when unavailable.
  - No generated startup hook runtime for Claude/Copilot/Antigravity/OMP/Pi today.

### `session-start`

**Policy source:** `packages/cli/library/canonical/hooks/session-start.md`

- **Purpose:** Session bootstrap reminder policy:
  - confirm active objective and current gate status,
  - load latest handoff,
  - identify next verification seam,
  - surface blockers early.
- **Trigger event:** conceptual workflow convention ("beginning of work session"), not emitted as a dedicated per-tool runtime hook.

### `pre-commit`

**Policy source:** `packages/cli/library/canonical/hooks/pre-commit.md`

- **Purpose:** Before local commits, enforce token-rent budget checks.
- **Trigger event:** pre-commit.
- **Runtime behavior:**
  - `.githooks/pre-commit` (active local hook): runs `token-rent-check` only. Skips if `go` is not available.
  - `.husky/pre-commit` (optional, if configured): enforces RPI gate attestation (human gate approval check) locally.
  - CI (`.github/workflows/rpi-gate-check.yml`): enforces RPI gate attestation on pull_request and push to main/master.
- **Note:** RPI gate attestation is primarily enforced via CI. The `.husky/pre-commit` hook provides optional local enforcement; the `.githooks/pre-commit` hook handles token-rent only.

### `rpi-gate-check`

**Runtime source:** `packages/cli/library/hooks/rpi-gate-check.yml`

- **Purpose:** Enforce RPI process gates on non-trivial change sets (research + approved plan evidence).
- **Trigger event:** GitHub Actions `pull_request` (`opened|synchronize|reopened`) and `push` to `main`/`master`.
- **Runtime behavior:** checks changed lines, recent `plan.md`/`research.md`, and human-authored `Human Gate: APPROVED` marker in a plan.

### OpenCode hook runtime plugin

**Runtime source:** `packages/cli/library/opencode/plugins/vibe-lab-hooks.js`

- **Purpose:** single plugin implementation point for OpenCode hook policy.
- **Trigger events implemented in runtime:**
  - `session.created` (startup-self-heal),
  - `tool.execute.before` (destructive shell block),
  - `session.idle` (objective completion gate checks).
- **Supported policy surface:** OpenCode receives all shipped hook behaviors in one plugin file.

## Per-tool hook surfaces

| Tool | Capability | Hook surface emitted | Notes |
|---|---|---|---|
| opencode | **supported** | `.opencode/plugins/vibe-lab-hooks.js` | Full plugin runtime; events: `session.created`, `tool.execute.before`, `session.idle`, `experimental.session.compacting` |
| claude | **supported** | `.claude/hooks/*.sh` + `.claude/settings.json` | Shell command hooks wired via `PreToolUse` and `Stop` settings |
| copilot | **partial** | `.github/hooks/*.{json,sh}` | Project scope only; limited to pre-exec and stop events |
| antigravity | **partial** | `.gemini/hooks/lazyai/*.sh` + `.agents/hooks.json` + `.gemini/settings.json` | Beta support level; limited to pre-exec and stop events |
| omp | **partial** | `.omp/hooks/pre/*.ts` | Beta support level; only `before_tool` surface (pre hooks) |
| pi | **instruction_only** | `.pi/extensions/block-destructive-shell.ts` | No `.pi/hooks` directory emitted; only one extension runtime; all other hooks are markdown-only |
| kiro | **supported** | `.kiro/hooks/*.json` + referenced hook scripts | Native Kiro v3 hook JSON; currently `block-destructive-shell` only, emitted for source-verified triggers only |

For each tool's managed output paths and MCP outputs, see [Tool Outputs](tool-outputs.md) and [Hook Template](../canonical/hook-template.md).
