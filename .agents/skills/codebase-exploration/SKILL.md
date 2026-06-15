---
name: codebase-exploration
description: Use when entering an unfamiliar repository or subsystem and you need a disciplined search-and-read strategy before changing code.
---

# Codebase Exploration

## When to Use

Use this skill before editing an unfamiliar codebase, package, or subsystem.

## Exploration Order

1. Read `README.md` or the nearest project overview.
2. Read package manifests such as `package.json`, `Cargo.toml`, `go.mod`, or equivalent.
3. Identify entry points: commands, routes, exported modules, or app bootstrap files.
4. Search for the target symbol, domain term, error message, or file owner.
5. Read only the sections needed for the next decision.
6. Trace callers and callees before changing exported behavior.

## File-Type Heuristics

- CLI task: inspect command registration before command implementation.
- Web route: inspect route declaration, request validation, handler, and tests.
- Library API: inspect exported types before private helpers.
- Build/tooling issue: inspect config, script entry, then generated outputs.
- Skill change: inspect `.agents/skills/`, symlinks, `bin/inject`, and `bin/doctor`.

## Stop-and-Ask Trigger

Stop and ask for direction when:

- two plausible subsystems own the same behavior;
- the codebase is too large to map without choosing a boundary;
- the validation signal is unknown and cannot be inferred;
- the task requires changing public behavior without a clear caller contract.

## Constraints

- Do not open files randomly.
- Do not read whole files when a targeted section is enough.
- Do not introduce a parallel convention when an existing pattern is visible.
