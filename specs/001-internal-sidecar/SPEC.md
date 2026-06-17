# SPEC: LazyAI Internal Sidecar

**Version:** 1.1.0
**Status:** Approved
**Author:** turbo-crank (Fortnite multi-agent system)
**Date:** 2026-06-17

---

## Goal

Provide an optional, LazyAI-native sidecar mechanism that overrides the default docs/specs/plans root directory. When no sidecar is configured, LazyAI resolves these directories to sensible defaults per scope. No external provider integration (Skeeper or otherwise).

---

## Constitution

**Problem:** LazyAI users need to organize docs, specs, and plans outside their project repository — for example, in a shared knowledge base, a monorepo docs directory, or a personal vault. Currently, all artifacts default to the project root, which couples documentation to code and prevents cross-project knowledge sharing.

**Solution:** A lightweight sidecar configuration block that redirects `docs_dir`, `specs_dir`, and `plans_dir` to user-specified paths. The sidecar is optional and layered: workspace scope takes priority, project scope provides overrides, and global scope supplies defaults.

**Non-negotiable constraints:**
- Sidecar is always optional. No sidecar = sensible defaults, not an error.
- No external provider fields (no `skeeper`, no `provider`, no `remote`).
- Workspace scope is the primary configuration surface.
- Resolution is deterministic and inspectable via `sidecar status`.
- Sidecar paths are always absolute or resolved relative to a well-defined anchor.

**Success looks like:**
- A user can run `lazyai-cli sidecar init` and get a working sidecar config.
- `lazyai-cli sidecar status` shows the resolved paths for the current scope.
- Commands that read/write docs/specs/plans use the sidecar-resolved paths transparently.
- No sidecar configured = behavior identical to today.

**Out of scope:**
- Remote sidecar providers (S3, Git repos, APIs).
- Sidecar synchronization or replication.
- Sidecar content migration tools.
- Multi-sidecar (one sidecar per scope, not multiple).
- Sidecar for anything other than docs/specs/plans.

---

## Requirements

### R1: Sidecar Configuration Block

The sidecar is a YAML block within the existing `~/.lazyai/workspaces.yaml` file, attached to individual workspace entries. A project may also carry a `.lazyai-sidecar.yaml` file in its root. Global defaults live in `~/.lazyai/sidecar.yaml`.

**Workspace-level sidecar block:**
```yaml
workspaces:
  - name: my-project
    path: /Users/me/projects/my-project
    sidecar:
      path: /Users/me/kb/my-project-docs    # anchor path
      specs_dir: specs                       # relative to path, or absolute
      docs_dir: docs                         # relative to path, or absolute
      plans_dir: plans                       # relative to path, or absolute
active: my-project
```

**Project-level sidecar (`.lazyai-sidecar.yaml`):**
```yaml
sidecar:
  path: ../shared-docs
  specs_dir: specs
  docs_dir: docs
  plans_dir: plans
```

**Global sidecar (`~/.lazyai/sidecar.yaml`):**
```yaml
sidecar:
  path: /Users/me/kb
  specs_dir: specs
  docs_dir: docs
  plans_dir: plans
```

**Acceptance criteria:**
- [ ] Workspace YAML parsing tolerates missing `sidecar` block (backward compatible).
- [ ] All `*_dir` fields default to their name (e.g., `specs_dir` defaults to `"specs"`) when omitted.
- [ ] `path` is required when any sidecar block is present. A sidecar block with missing or empty `path` is invalid: `Resolve` returns an error, `sidecar status` returns non-zero, and `sidecar doctor` reports an ERROR.
- [ ] Relative paths in `*_dir` are resolved against `sidecar.path`.
- [ ] Absolute paths in `*_dir` are used as-is.
- [ ] `linked_projects` is reserved for a future cross-project feature; this version ignores the field if present in YAML and does not validate, merge, or manage it.

### R2: Resolution Priority

When resolving docs/specs/plans directories, LazyAI applies the following priority chain:

1. **Workspace sidecar** (from `~/.lazyai/workspaces.yaml` active workspace entry) — highest priority.
2. **Project sidecar** (from `<project-root>/.lazyai-sidecar.yaml`) — overrides global, but not workspace.
3. **Global sidecar** (from `~/.lazyai/sidecar.yaml`) — lowest priority, fallback.
4. **Scope default** — if no sidecar at any level:
   - `project` scope → `<project-root>/` (current behavior)
   - `workspace` scope → active workspace `path` (current behavior)
   - `global` scope → `~/.lazyai/` (current behavior)

**Whole-config replacement (not field-level merge):** When a higher-priority sidecar level is present, that level supplies the complete docs/specs/plans tuple. Missing `docs_dir`, `specs_dir`, and `plans_dir` default under that same level's `path`; they do not inherit from lower-priority sidecars.

**Level-specific relative-path anchors:** Sidecar `path` relative values are resolved against the directory of the config file that owns the block, not the requested command scope:
- workspace sidecar `path` → resolved against `~/.lazyai/` (the block lives in `workspaces.yaml`);
- project sidecar `path` → resolved against `projectRoot`;
- global sidecar `path` → resolved against `~/.lazyai/`.

**Acceptance criteria:**
- [ ] `sidecar status` shows the resolved path and which level provided it.
- [ ] A workspace sidecar with `specs_dir: my-specs` overrides a project sidecar with `specs_dir: other-specs`.
- [ ] A project sidecar overrides a global sidecar when no workspace sidecar exists.
- [ ] When no sidecar exists at any level, the scope default is used.
- [ ] Resolution is computed once per command invocation (no caching across commands).
- [ ] A project sidecar with `path: ../project-kb` and only `docs_dir: custom-docs` resolves docs to `../project-kb/custom-docs`, specs to `../project-kb/specs`, and plans to `../project-kb/plans` — not to global paths.
- [ ] A global sidecar with relative `path: kb` resolves docs/specs/plans under `~/.lazyai/kb`, even when the requested scope is `project`.

### R3: Commands

Five new subcommands under `lazyai-cli sidecar`:

| Command | Behavior |
|---|---|
| `sidecar init` | Interactive or flag-driven creation of a sidecar config at the appropriate scope. Prompts for scope (workspace/project/global) and path. |
| `sidecar status` | Displays resolved docs/specs/plans paths for the current scope, including which config level provided each value. |
| `sidecar attach` | Adds a sidecar block to the active workspace or creates a project-level sidecar. Requires `--path`. |
| `sidecar detach` | Removes the sidecar block from the active workspace or deletes the project-level sidecar file. Requires confirmation. |
| `sidecar doctor` | Validates all configured sidecar paths in the requested scope chain exist and are writable. Reports issues with exit codes. |

**Acceptance criteria:**
- [ ] `sidecar init --scope workspace --path /tmp/kb` creates a sidecar block in the active workspace entry.
- [ ] `sidecar init --scope project --path ../kb` creates `.lazyai-sidecar.yaml` in the project root.
- [ ] `sidecar init --scope global --path ~/kb` creates/updates `~/.lazyai/sidecar.yaml`.
- [ ] `sidecar status` shows a table with columns: Scope, Config Level, Docs Dir, Specs Dir, Plans Dir.
- [ ] `sidecar attach --path /tmp/kb` without `--scope` defaults to workspace scope.
- [ ] `sidecar detach` prompts for confirmation and shows what will be removed before acting.
- [ ] `sidecar doctor` exits 0 when all paths are valid; non-zero when ERROR issues are found; WARN issues do not fail the command.
- [ ] `sidecar doctor` validates every configured sidecar level in the requested scope chain: `--scope global` validates global only; `--scope project` validates global + project; `--scope workspace` validates global + project + workspace. Missing sidecar at any level is not an issue.
- [ ] `sidecar doctor` reports: missing directories, non-writable directories, invalid YAML. (linked_projects is reserved and not validated.)

### R4: Edge Cases
| Scenario | Expected Behavior |
|---|---|
| Sidecar `path` is a relative path in workspace config | Resolved relative to `~/.lazyai/` (the config file location) |
| Sidecar `path` is a relative path in project config | Resolved relative to `projectRoot` |
| Sidecar `path` is a relative path in global config | Resolved relative to `~/.lazyai/` |
| Sidecar `path` does not exist on disk | `sidecar doctor` reports WARN; `sidecar init` and `sidecar attach` accept (directories are created on first write) |
| Workspace has sidecar but workspace is not active | Sidecar is ignored; resolution falls through to project/global/default |
| User runs `sidecar detach` with no sidecar configured | Command reports "no sidecar configured" and exits 0 |
| Sidecar `path` is a file, not a directory | Validation rejects it; `sidecar doctor` reports ERROR; `sidecar attach` and `sidecar init` refuse |
| Sidecar `path` is empty or missing in a present block | `Resolve` returns an error; `sidecar doctor` reports ERROR; `sidecar status` exits non-zero |
| YAML is malformed | Command fails with a parse error pointing to the file and line |
| `.bak` backup file | Single-slot backup: each write overwrites the previous `.bak`; no timestamped history is kept |

### R5: Non-Goals (Explicit Exclusions)

- **No Skeeper integration.** The sidecar is purely local. No `skeeper` field, no provider abstraction, no remote sync.
- **No content migration.** `sidecar init` does not move existing docs/specs/plans to the new location.
- **No multi-sidecar.** One sidecar per scope level. No chaining or fallback lists.
- **No sidecar for other artifact types.** Only docs, specs, and plans. Not for code, config, or cache.
- **No automatic sidecar discovery.** Sidecars are explicitly configured, not auto-detected from parent directories or environment variables.
- **`linked_projects` is reserved.** This version ignores the field if present in YAML and does not validate, merge, or manage it. A future cross-project feature may define it.

---

## Durability Policy

All workspace registry mutations use a locked read-modify-write path and atomic file replacement. Project and global sidecar file writes use atomic replacement. The lock file sits beside `workspaces.yaml` (as `workspaces.yaml.lock`) and protects concurrent mutating CLI invocations from lost updates. Atomic replacement means: write to a temp file in the same directory, `fsync` the temp file, then `os.Rename` over the target. Existing files are backed up to a single-slot `.bak` before replacement.

## Dependencies

- Existing `~/.lazyai/workspaces.yaml` structure (extend `WorkspaceEntry` with optional `SidecarConfig`).
- Existing CLI framework (cobra + fang).
- No new external dependencies.

---

## Acceptance Criteria (Summary)

- [ ] AC1: Workspace YAML with no sidecar block parses without error (backward compat).
- [ ] AC2: `sidecar status` shows correct resolved paths for all three scope levels.
- [ ] AC3: Resolution priority (workspace > project > global > default) is correct.
- [ ] AC4: All five commands (`init`, `status`, `attach`, `detach`, `doctor`) exist and behave as specified.
- [ ] AC5: `sidecar doctor` catches all documented edge cases.
- [ ] AC6: No sidecar configured = behavior identical to current release.
- [ ] AC7: No Skeeper or provider fields anywhere in config or code.