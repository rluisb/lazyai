# startup-self-heal Policy

## Purpose

At session start, run a scoped health check for the current CLI and regenerate only that CLI's project-local artifacts when drift or missing files are detected.

## Events

- Claude Code: `SessionStart`
- OpenCode plugin: `session.created`
- OMP/Pi: unsupported

## Decision

- Allow when: the scoped startup check for the current CLI passes.
- Warn when: scoped doctor fails, scoped regeneration runs, or regeneration still leaves issues.
- Deny when: never. This is advisory automation, not a safety gate.

## Runtime

- Claude Code: generated command hook runs `bin/startup-self-heal --cli claude` for `startup|resume` sessions.
- OpenCode: generated plugin runs `bin/startup-self-heal --cli opencode` on `session.created`.
- OMP/Pi: no project-local startup hook is generated.

## Fail-Closed Semantics

- Advisory hook. If runtime is unavailable or repair fails, print a warning and continue without blocking the session.

## Scope

- Scoped doctor/inject touch only the current CLI surface plus the matching generated context file.
- Full docs and cross-CLI regeneration still use plain `bin/inject`.
