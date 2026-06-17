# TECH-SPEC: LazyAI Internal Sidecar

**Version:** 1.1.0
**Status:** Approved
**Author:** turbo-crank (Fortnite multi-agent system)
**Date:** 2026-06-17

---

## Overview

This document describes the implementation design for the LazyAI internal sidecar. It covers data structures, file layout, resolution algorithm, command wiring, and error handling. The implementation lives in `packages/cli/cmd/sidecar.go` (new file) with supporting types in `packages/cli/internal/sidecar/` (new package).

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    lazyai-cli sidecar                    │
├─────────────────────────────────────────────────────────┤
│  init    status   attach   detach   doctor              │
│    │        │        │        │        │                │
│    └────────┴────────┴────────┴────────┘                │
│                      │                                  │
│              ┌───────▼────────┐                         │
│              │  SidecarResolver│                         │
│              │  (resolution    │                         │
│              │   priority      │                         │
│              │   chain)        │                         │
│              └───────┬────────┘                         │
│                      │                                  │
│     ┌────────────────┼────────────────┐                 │
│     ▼                ▼                ▼                 │
│ ~/.lazyai/     <project>/.lazyai-  ~/.lazyai/          │
│ workspaces.yaml  sidecar.yaml     sidecar.yaml          │
│ (workspace)     (project)         (global)              │
└─────────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **No new config file format.** Sidecar extends existing YAML structures. Workspace sidecar lives in `workspaces.yaml`. Project sidecar is a standalone `.lazyai-sidecar.yaml`. Global sidecar is `~/.lazyai/sidecar.yaml`.

2. **Resolution is a pure function.** `SidecarResolver.Resolve(scope)` takes a scope enum and returns resolved paths. No global state, no caching. Callers are responsible for passing the correct scope.

3. **Backward compatible by default.** The `SidecarConfig` struct uses `omitempty` on all fields. Existing `WorkspaceEntry` gets an optional `Sidecar *SidecarConfig` field. Old YAML without `sidecar` parses to `nil`.

4. **Commands perform boundary validation and delegate.** Each command validates its inputs (project root accessibility, scope flags, path presence) and delegates config load/save/resolution/validation to `internal/sidecar`.

---

## Data Structures

### `SidecarConfig` (new — `packages/cli/internal/sidecar/types.go`)

```go
package sidecar

// SidecarConfig holds the sidecar configuration for a single scope level.
type SidecarConfig struct {
    Path     string `yaml:"path"`
    SpecsDir string `yaml:"specs_dir,omitempty"`
    DocsDir  string `yaml:"docs_dir,omitempty"`
    PlansDir string `yaml:"plans_dir,omitempty"`
}

// ResolvedPaths holds the final resolved directory paths.
type ResolvedPaths struct {
    DocsDir     string
    SpecsDir    string
    PlansDir    string
    ConfigLevel string // "workspace", "project", "global", "default"
}

// Scope represents the resolution scope.
type Scope int

const (
    ScopeWorkspace Scope = iota
    ScopeProject
    ScopeGlobal
)
```

### Extended `WorkspaceEntry` (modify — `packages/cli/cmd/workspace.go`)

```go
type WorkspaceEntry struct {
    Name    string              `yaml:"name"`
    Path    string              `yaml:"path"`
    Sidecar *sidecar.SidecarConfig `yaml:"sidecar,omitempty"`
}
```

### Global Sidecar Config (new file — `~/.lazyai/sidecar.yaml`)

```go
// GlobalSidecarConfig is the top-level structure of ~/.lazyai/sidecar.yaml
type GlobalSidecarConfig struct {
    Sidecar *SidecarConfig `yaml:"sidecar"`
}
```

### Project Sidecar Config (new file — `<project>/.lazyai-sidecar.yaml`)

```go
// ProjectSidecarConfig is the top-level structure of .lazyai-sidecar.yaml
type ProjectSidecarConfig struct {
    Sidecar *SidecarConfig `yaml:"sidecar"`
}
```

---

## File Layout

```
packages/cli/
├── cmd/
│   ├── workspace.go          # MODIFY: add Sidecar field to WorkspaceEntry
│   └── sidecar.go            # NEW: sidecar command group + subcommands
├── internal/
│   └── sidecar/              # NEW package
│       ├── types.go          # SidecarConfig, ResolvedPaths, Scope
│       ├── resolver.go       # Resolve(scope) -> ResolvedPaths
│       ├── loader.go         # LoadWorkspaceSidecar, LoadProjectSidecar, LoadGlobalSidecar
│       ├── writer.go         # WriteWorkspaceSidecar, WriteProjectSidecar, WriteGlobalSidecar
│       ├── remover.go        # RemoveWorkspaceSidecar, RemoveProjectSidecar
│       └── doctor.go         # Doctor(sidecar) -> []Issue
```

---

## Resolution Algorithm

```
func Resolve(scope Scope, projectRoot string) ResolvedPaths

1. Initialize result with scope defaults:
   - ScopeWorkspace → active workspace path
   - ScopeProject   → projectRoot
   - ScopeGlobal    → ~/.lazyai/

2. Load global sidecar (~/.lazyai/sidecar.yaml).
   If present and valid, apply as base. Set ConfigLevel = "global". Resolve path against `~/.lazyai/`.

3. Load project sidecar (<projectRoot>/.lazyai-sidecar.yaml).
   If present and valid, replace global. Set ConfigLevel = "project". Resolve path against `projectRoot`.

4. Load workspace config (~/.lazyai/workspaces.yaml).
   If active workspace has sidecar block, replace project. Set ConfigLevel = "workspace". Resolve path against `~/.lazyai/`.

5. For each level, resolve paths:
   a. If sidecar.path is relative, resolve against anchor:
      - workspace: relative to ~/.lazyai/
      - project:   relative to projectRoot
      - global:    relative to ~/.lazyai/
   b. For each *_dir field:
      - If absolute, use as-is.
      - If relative, resolve against sidecar.path.
      - If empty/missing, default to field name (e.g., "specs").

6. Return ResolvedPaths with final values and ConfigLevel.
```

### Replacement Semantics

- A higher-priority sidecar level replaces the lower-priority config entirely (whole-config replacement, not field-level merge).
- Missing `docs_dir`, `specs_dir`, and `plans_dir` in the winning level default under that level's `path`; they do not inherit from lower-priority sidecars.
- `linked_projects` is reserved and ignored. No merge, no validation, no management.

---

## Command Wiring

### `sidecar init`

```
Flags:
  --scope       string   workspace|project|global (default: workspace)
  --path        string   sidecar root path (required)
  --specs-dir   string   override specs directory name
  --docs-dir    string   override docs directory name
  --plans-dir   string   override plans directory name

Flow:
1. Validate --path is non-empty. The sidecar root does NOT need to exist on disk (it is created on first write); `sidecar doctor` will WARN if it is missing.
2. Based on --scope:
   - workspace: load workspaces.yaml, add sidecar block to active entry, save.
   - project:   create .lazyai-sidecar.yaml in project root.
   - global:    create/update ~/.lazyai/sidecar.yaml.
3. Print confirmation with resolved paths.
```

### `sidecar status`

```
Flags: none

Flow:
1. Determine current scope (from context: are we in a project? is there an active workspace?).
2. Call Resolve(scope, projectRoot).
3. Print table:
   Scope       Config Level   Docs Dir            Specs Dir           Plans Dir
   ─────────   ────────────   ────────────────    ────────────────    ────────────────
   workspace   workspace      /Users/me/kb/docs   /Users/me/kb/specs  /Users/me/kb/plans
```

### `sidecar attach`

```
Flags:
  --scope       string   workspace|project (default: workspace)
  --path        string   sidecar root path (required)
  --specs-dir   string   override specs directory name
  --docs-dir    string   override docs directory name
  --plans-dir   string   override plans directory name

Flow:
1. Same as init but requires an existing config target (fails if no active workspace for workspace scope).
2. Does NOT create the sidecar path directory (doctor will flag it).
```

### `sidecar detach`

```
Flags:
  --scope   string   workspace|project (default: workspace)
  --force   bool     skip confirmation prompt

Flow:
1. Based on --scope:
   - workspace: load workspaces.yaml, remove sidecar block from active entry, save.
   - project:   delete .lazyai-sidecar.yaml.
2. If no sidecar exists at that level, print message and exit 0.
3. Prompt for confirmation (unless --force).
4. Print what was removed.
```

### `sidecar doctor`

```
Flags:
  --scope   string   workspace|project|global (default: determined by `determineScope`)

Flow:
1. Load sidecar configs for the requested scope chain:
   - global: validate `[global]` if configured;
   - project: validate `[global, project]` where each configured cfg is non-nil;
   - workspace: validate `[global, project, workspace]` where each configured cfg is non-nil.
2. For each configured candidate:
   a. Check sidecar `path` is non-empty (ERROR if empty).
   b. Resolve `path` against the level-specific anchor.
   c. Check path exists and is a directory (ERROR if file; WARN if missing).
   d. Check path is writable.
   e. Check each *_dir resolves to an existing directory (or can be created).
3. Print issues with severity (ERROR/WARN).
4. Exit 0 if no issues or only WARN issues; non-zero (2) if any ERROR issues are found. WARN issues do not fail the command.
```

---

## Error Handling

| Error Condition | Behavior |
|---|---|
| `workspaces.yaml` missing | Treated as empty config (no workspaces). Not an error. |
| `workspaces.yaml` malformed YAML | Command fails with parse error including file path and line. |
| `.lazyai-sidecar.yaml` missing | Treated as no project sidecar. Not an error. |
| `.lazyai-sidecar.yaml` malformed YAML | Command fails with parse error. |
| `~/.lazyai/sidecar.yaml` missing | Treated as no global sidecar. Not an error. |
| Sidecar `path` is a file | `init`/`attach` reject. `doctor` reports ERROR. |
| Sidecar `path` does not exist | `init`/`attach` accept (directories created on first write). `doctor` reports WARN. |
| No active workspace for workspace-scoped command | Command fails with message "no active workspace set; use 'lazyai-cli workspace switch' first". |
| Sidecar `path` empty or missing in a present block | `Resolve` returns an error; `doctor` reports ERROR; `status` exits non-zero. |
| `linked_projects` present in YAML | Field is reserved and ignored. No error, no warning, no validation. |

---

## Testing Strategy

### Unit Tests (`packages/cli/internal/sidecar/`)

- `resolver_test.go`: Test all resolution priority combinations.
  - No sidecar at any level → defaults.
  - Global only → global paths.
  - Project overrides global (whole-config replacement).
  - Workspace overrides project (whole-config replacement).
  - Relative path resolution with level-specific anchors.
  - Absolute path passthrough.
  - Default *_dir values.
  - Empty sidecar path rejection.
  - Global relative path anchors to `~/.lazyai/` even at project scope.
  - Project relative path anchors to `projectRoot` even at workspace scope.

- `loader_test.go`: Test YAML parsing with test fixtures.
  - Valid workspace with sidecar.
  - Valid workspace without sidecar (backward compat).
  - Malformed YAML.
  - Missing files.

- `doctor_test.go`: Test scope-chain validation.
  - No sidecars at any level → no issues.
  - Global + project validated for project scope.
  - Global + project + workspace validated for workspace scope.
  - Level-specific anchors used for relative paths.
  - Empty path is ERROR.
  - Missing sidecar root is WARN.
  - Path is a file is ERROR.
  - LinkedProjects ignored (no issues emitted).

- `writer_test.go`: Test atomic write and durability.
  - Project sidecar rejects missing/file project root.
  - Project sidecar writes atomically and backs up existing file.
  - Global sidecar writes atomically and backs up existing file.
  - SaveWorkspaceConfig writes single-slot backup.
  - UpdateWorkspaceConfig preserves concurrent mutations.

- `remover_test.go`: Test sidecar removal.
  - Remove active workspace sidecar preserves other entries.
  - No active workspace is no-op.
  - No sidecar is no-op.
  - Active workspace missing returns error.
  - Remove project sidecar removes file.
  - Remove project sidecar missing file is no-op.
  - Workspace removal writes backup through SaveWorkspaceConfig.

### Files Helper Tests (`packages/cli/internal/files/`)

- `files_test.go`: Test atomic write and lock primitives.
  - AtomicWriteFile creates file without backup when target does not exist.
  - AtomicWriteFile replaces target and writes single-slot backup.
  - AtomicWriteFile overwrites single-slot backup on second write.
  - WithFileLock serializes concurrent critical sections.
  - WithFileLock times out when lock is held.
  - WithFileLock removes stale locks.

### Integration Tests (CLI level)

- `sidecar init` creates correct config files.
- `sidecar status` output format and root-error propagation.
- `sidecar attach` / `sidecar detach` round-trip.
- `sidecar doctor` exit codes and scope-chain validation.
- `sidecar attach` rejects missing project path without creating it.
- `determineScope` returns workspace config load errors instead of silent fallback.
- Backward compat: existing workspace commands still work via sidecar store.
- Workspace add/switch/status/list use the sidecar workspace config store.

---

## Migration Notes

- **No data migration needed.** Existing `workspaces.yaml` files without `sidecar` blocks parse correctly (the new `Sidecar *SidecarConfig` field defaults to `nil` via `omitempty`).
- **No new dependencies.** All YAML handling uses existing `gopkg.in/yaml.v3`.
- **No breaking changes.** The `WorkspaceEntry` struct gains an optional field. All existing commands are unaffected.