# Hooks

## What hooks do

Hooks are guardrail or automation scripts that run at tool lifecycle events (for example, command execution, session stop, or commit/workflow boundaries).
They can block an action, emit an advisory warning, or trigger maintenance work.
Support is adapter-specific and varies by tool.

## Shipped hooks

Each shipped hook is defined in `packages/cli/library/hooks/`, `packages/cli/library/canonical/hooks/`, or the OpenCode runtime plugin.

### `block-destructive-shell`

**Policy source:** `packages/cli/library/hooks/block-destructive-shell.md`

- **Purpose:** Block obvious destructive commands (for example `rm -rf /`, `mkfs`, `dd if=/dev/zero ...`, `shutdown`, `reboot`).
- **Trigger event:**
  - Claude/OpenCode/Copilot/Antigravity: shell-tool pre-execution.
  - Pi/OMP: extension/tool-call hook points for `bash`/`tool_call`.
- **Runtime behavior:**
  - OpenCode: plugin `event.type === "tool.execute.before"` in `.opencode/plugins/vibe-lab-hooks.js`.
  - Claude/Copilot/Antigravity: runtime JSON/stdin shell command descriptors in generated hook scripts.
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
- **Trigger event:** conceptual workflow convention (“beginning of work session”), not emitted as a dedicated per-tool runtime hook.

### `pre-commit`

**Policy source:** `packages/cli/library/canonical/hooks/pre-commit.md`

- **Purpose:** Before local commits, ensure the plan/budget rules are honored (RPI compliance and token-rent checks).
- **Trigger event:** pre-commit.
- **Runtime behavior:** repository-level hook policy asset; this policy file is retained with adapter target `[none]` in manifest.

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

| Tool | Hook surface emitted | Notes |
|---|---|---|
| opencode | `.opencode/plugins/vibe-lab-hooks.js` | `OpenCodeAdapter.Install` copies `opencode/plugins` to `.opencode/plugins`. |
| claude | `.claude/hooks/*.sh` + `.claude/settings.json` | `claude` settings wire `PreToolUse` and `Stop` commands to the generated scripts. |
| copilot | `.github/hooks/*.{json,sh}` | `copilot/hooks` JSON descriptors pair with generated shell scripts; project scope only. |
| antigravity | `.gemini/hooks/lazyai/*.sh` + `.agents/hooks.json` + `.gemini/settings.json` | `AntigravityAdapter` copies hook scripts under `.gemini/hooks` and writes hook mapping in `.agents/hooks.json`. |
| pi | `.pi/extensions/block-destructive-shell.ts` | `PiAdapter` comments “Pi safety hooks ship as extensions at `.pi/extensions/*.ts`”; no `.pi/hooks` directory emitted. |
| omp | `.omp/hooks/pre/*.ts` | `OmpAdapter` ensures `.omp/hooks/pre` and copies `omp/hooks/*` there; available as TypeScript hook factories. |
| kiro | not supported for hooks in current adapter | `KiroAdapter` emits only agents/skills/prompts and no hook directory. |

For each tool’s managed output paths and MCP outputs, see [Tool Outputs](tool-outputs.md) and [Hook Template](../canonical/hook-template.md).