# startup-self-heal Policy

## Purpose

Preserve scoped startup-repair guidance as maintainer/reference policy without advertising it as an emitted runtime hook for supported adapters.

## Events

- Generated adapters: none
- Repository harness only: `bin/startup-self-heal`
- OMP/Pi: unsupported

## Decision

- Allow when: maintainers explicitly run the repository harness for a local repair workflow.
- Warn when: documentation or generated assets imply that startup self-heal is a shipped adapter behavior.
- Deny when: never. This is documentation/harness guidance, not an active runtime gate.

## Runtime

- Claude Code: no project-local startup hook is emitted.
- OpenCode: no startup self-heal plugin event is emitted.
- Copilot: no startup self-heal hook descriptor is emitted.
- Antigravity: no startup self-heal hook is emitted.
- OMP/Pi: no project-local startup hook is generated.

## Fail-Closed Semantics

- No generated runtime behavior. If maintainers need scoped repair, they run the repository harness explicitly.

## Scope

- The repository harness may still use scoped doctor/inject internally for local maintenance of this repository.
- Generated setups must not depend on `bin/startup-self-heal`.
