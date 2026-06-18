# TECH-SPEC: LazyAI Internal Sidecar

**Version:** 1.0.0
**Status:** Draft
**Author:** turbo-crank (Fortnite multi-agent system)
**Date:** 2026-05-24

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

4. **Commands are thin wrappers.** Each command calls into the resolver or a config writer. No business logic in command handlers.

---

## Data Structures

### `SidecarConfig` (new — `packages/cli/internal/sidecar/types.go`)

```go
package sidecar

// SidecarConfig holds the sidecar configuration for a single scope level.
type SidecarConfig struct {
    Path           string              `yaml:"path"`
    SpecsDir       string              `yaml:"specs_dir,omitempty"`
    DocsDir        string              `yaml:"docs_dir,omitempty"`
    PlansDir       string              `yaml:"plans_dir,omitempty"`
    LinkedProjects []LinkedProject     `yaml:"linked_projects,omitempty"`
}

// LinkedProject is a cross-project reference within a sidecar.
type LinkedProject struct {
    Name string `yaml:"name"`
    Path string `yaml:"path"`
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
   If present and valid, apply as base. Set ConfigLevel = "global".

3. Load project sidecar (<projectRoot>/.lazyai-sidecar.yaml).
   If present and valid, merge over global. Set ConfigLevel = "project".

4. Load workspace config (~/.lazyai/workspaces.yaml).
   If active workspace has sidecar block, merge over project. Set ConfigLevel = "workspace".

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

### Merge Semantics

- Each level provides a full or partial `SidecarConfig`.
- A field at a higher-priority level completely replaces the lower-priority value (no deep merge of individual fields).
- `linked_projects` are merged by name: higher priority wins for same-name entries, lower priority entries are appended for unique names.

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
1. Validate --path is a directory or can be created.
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
  --scope   string   workspace|project|global|all (default: all)

Flow:
1. Load sidecar configs for requested scope(s).
2. For each config:
   a. Check path exists and is a directory.
   b. Check path is writable.
   c. Check each *_dir resolves to an existing directory (or can be created).
   d. Check linked_projects paths exist.
3. Print issues with severity (ERROR/WARN).
4. Exit 0 if no errors, 1 if warnings only, 2 if errors.
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
| `linked_projects` path does not exist | `doctor` reports WARN. Other commands warn but proceed. |

---

## Testing Strategy

### Unit Tests (`packages/cli/internal/sidecar/`)

- `resolver_test.go`: Test all resolution priority combinations.
  - No sidecar at any level → defaults.
  - Global only → global paths.
  - Project overrides global.
  - Workspace overrides project.
  - Relative path resolution.
  - Absolute path passthrough.
  - Default *_dir values.
  - linked_projects merge semantics.

- `loader_test.go`: Test YAML parsing with test fixtures.
  - Valid workspace with sidecar.
  - Valid workspace without sidecar (backward compat).
  - Malformed YAML.
  - Missing files.

- `doctor_test.go`: Test validation logic.
  - All paths valid.
  - Missing sidecar path.
  - Non-writable directory.
  - File where directory expected.
  - Broken linked_projects.

### Integration Tests (CLI level)

- `sidecar init` creates correct config files.
- `sidecar status` output format.
- `sidecar attach` / `sidecar detach` round-trip.
- `sidecar doctor` exit codes.
- Backward compat: existing workspace commands still work.

---

## Migration Notes

- **No data migration needed.** Existing `workspaces.yaml` files without `sidecar` blocks parse correctly (the new `Sidecar *SidecarConfig` field defaults to `nil` via `omitempty`).
- **No new dependencies.** All YAML handling uses existing `gopkg.in/yaml.v3`.
- **No breaking changes.** The `WorkspaceEntry` struct gains an optional field. All existing commands are unaffected.