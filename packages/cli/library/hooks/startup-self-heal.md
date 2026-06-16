# startup-self-heal Policy

## Purpose

At session start, run a scoped health check for the current CLI and regenerate only that CLI's project-local artifacts when drift or missing files are detected.

## Events

- OpenCode plugin: `session.created`
- Claude Code: policy retained; no Claude runtime hook is emitted by LazyAI today
- Copilot: no hook descriptor is emitted
- Antigravity: no startup hook is emitted
- OMP/Pi: unsupported

## Decision

- Allow when: the scoped startup check for the current CLI passes.
- Warn when: scoped doctor fails, scoped regeneration runs, the harness script is unavailable, or regeneration still leaves issues.
- Deny when: never. This is advisory automation, not a safety gate.

## Runtime

- OpenCode: generated plugin calls `bin/startup-self-heal --cli opencode --event session.created --quiet` when the repository harness script exists.
- Other generated adapters: no startup self-heal runtime hook is emitted today.

## Fail-Closed Semantics

- Advisory hook. If runtime support is unavailable or repair fails, print a warning and continue without blocking the session.

## Scope

- Scoped doctor/inject touch only the current CLI surface plus the matching generated context file.
- Full docs and cross-CLI regeneration still use plain `bin/inject`.
