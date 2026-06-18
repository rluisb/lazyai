---
name: catalog-manage
description: Manage the orchestration catalog lifecycle — version, diff, promote, deprecate, and remove chains, teams, and workflows safely.
argument-hint: "[list | diff | promote | deprecate | remove] [kind/name]"
trigger: /catalog-manage
phase: meta
---

# Catalog Manage Skill

Manage orchestration catalog entries (chains, teams, workflows) through their full lifecycle: creation, versioning, promotion from ephemeral to permanent, diffing between versions, deprecation, and removal.

## When to Use

- Promoting an ephemeral `dynamic-compose` definition to a permanent catalog entry
- Versioning an existing chain/team/workflow after a meaningful change
- Comparing two versions of a definition (what changed and why)
- Deprecating or removing an obsolete definition
- Listing all versions of a definition to understand its evolution

## When NOT to Use

- Running a chain/team/workflow → use `orchestrate` skill
- Composing a new ad-hoc definition → use `dynamic-compose` skill
- Overriding agent/domain/mode at runtime → use `chain-customize` skill

## The Catalog Lifecycle

```
[Ephemeral]                [Registered]              [Active]            [Deprecated]
dynamic-compose            catalog_create_version     catalog_set_active   catalog_deactivate
creates in-memory           writes to catalog DB      pins to version     clears active pointer
     │                           │                        │                     │
     └─── success? ──────────────┘                        │                     │
           │                                                │                     │
           └─── 3+ successful runs? ──── recommend promotion ──┘              │
           │                                                                │
           └─── no longer needed? ──────────────────────────────────────────┘
```

### States

| State | Meaning | How it got here | What can happen next |
|-------|--------|-----------------|---------------------|
| **Ephemeral** | Exists in memory only, not in catalog | `dynamic-compose` created it | Register it or discard it |
| **Registered** | Exists in catalog with one or more versions, no active pointer | `catalog_create_version` | Set active or remove |
| **Active** | Has an active version pinned; will be discovered by `list_catalog` | `catalog_set_active` | Version it, deactivate it |
| **Deprecated** | Active pointer cleared; no longer surfaced by default | `catalog_deactivate` | Remove it, or reactivate it |
| **Removed** | Deleted entirely from catalog | `catalog_remove` | Gone (irreversible) |

## Procedures

### Promote Ephemeral → Registered → Active

When a dynamic composition has proven useful and should be persisted:

1. **Verify the definition is sound.** It should have:
   - All steps with valid agent names (agents that actually exist)
   - All transitions pointing to real step IDs or `done`/`handoff`
   - A clear description
   - Version `0.1.0` (ephemeral-originated) or `1.0.0` (manually crafted)

2. **Call `catalog_create_version`** with the full definition:
   ```
   catalog_create_version({
     kind: "chain" | "team" | "workflow",
     name: "<slug>",
     frontmatter: { version: 1 },
     body: "<JSON-stringified-definition>"
   })
   ```
   This creates version 1 (immutable). Checksum-deduplication means if the same body already exists, it's a no-op.

3. **Call `catalog_set_active`** to make it discoverable:
   ```
   catalog_set_active({
     kind: "chain",
     name: "<slug>",
     version: 1
   })
   ```

4. **Verify** — call `list_catalog` to confirm the new entry appears.

### Promote after 3+ Successful Runs

When the same composition has been used successfully 3+ times:

1. Check `.ai/orchestration/state/` for past run records matching the composition's shape.
2. If ≥3 successful runs found, propose promotion to the user:
   > "The chain `websocket-feature` has completed successfully 4 times. Recommend promoting to a permanent catalog entry. Approve?"
3. On approval, follow the **Promote Ephemeral → Active** procedure above.
4. Set version to `1.0.0` (proven, not experimental).

### Version an Existing Definition

When a chain/team/workflow needs a meaningful change:

1. **Read the current definition** using `catalog_get_version` (or read the file directly).
2. **Make the change** — modify the body (add steps, change agents, adjust transitions).
3. **Call `catalog_create_version`** with the modified body. This creates a new immutable version.
4. **Diff the versions** to confirm the change is correct:
   ```
   catalog_diff({
     kind: "chain",
     name: "<slug>",
     fromVersion: 1,
     toVersion: 2
   })
   ```
5. **If the diff looks correct**, call `catalog_set_active` to pin to the new version.
6. **If the diff reveals unintended changes**, do NOT set active. Fix and create a new version.

**Important:** Old versions are never modified. They remain in the catalog as immutable history. The active pointer simply moves forward.

### Diff Two Versions

To understand what changed between versions:

```
catalog_diff({
  kind: "chain",
  name: "<slug>",
  fromVersion: <N>,
  toVersion: <M>
})
```

Returns both versions side-by-side. Use this before promoting a new version to active.

### Deprecate a Definition

When a chain/team/workflow is no longer needed but should remain in history:

1. **Confirm no active runs** — check `list_jobs` and `.ai/orchestration/state/` for in-progress runs using this definition.
2. **Call `catalog_deactivate`:**
   ```
   catalog_deactivate({
     kind: "chain",
     name: "<slug>"
   })
   ```
3. This clears the active pointer. The definition won't appear in `list_catalog` results anymore, but all versions are preserved.
4. **Record why** — add an observation in the knowledge graph or memory noting why it was deprecated.

### Remove a Definition (Destructive)

**Use with extreme caution.** This deletes the definition and all its version history.

1. **Confirm it's safe:** No active runs, no downstream workflows reference it, user has explicitly approved.
2. **Check for references** — search workflows for `ref: "<name>"` to ensure no workflow depends on this definition.
3. **Call `catalog_remove`:**
   ```
   catalog_remove({
     kind: "chain",
     name: "<slug>"
   })
   ```

**⚠️ This is irreversible.** Prefer `catalog_deactivate` in almost all cases.

### Export a Version to File

To write a specific version to disk (e.g., for review, backup, or manual editing):

```
catalog_export_version({
  kind: "chain",
  name: "<slug>",
  version: <N>,
  targetPath: "/path/to/output.json"
})
```

This is the only orchestrator-initiated write to host files. Always user-driven.

## Promotion Criteria

Not every definition deserves to be permanent. Use these criteria:

| Criterion | Must Meet | Check |
|-----------|-----------|-------|
| **Proven** | ≥1 successful run (≥3 for auto-promotion) | `.ai/orchestration/state/` records |
| **Complete** | All steps have valid agents, transitions, descriptions | Read the definition |
| **Reusable** | The shape applies to more than one specific task | Ask: "could another task use this same flow?" |
| **Distinct** | No existing catalog entry covers the same need | `list_catalog` comparison |
| **Described** | Has a clear, specific description | Read the `description` field |

If any criterion fails, keep the definition ephemeral or fix it before promoting.

## Version Numbering Convention

| Pattern | Meaning | When |
|---------|---------|------|
| `0.1.0` | Ephemeral-originated, unproven | First registration from `dynamic-compose` |
| `1.0.0` | Proven, stable | After ≥3 successful runs OR manually crafted |
| `1.1.0` | Minor change (new step, updated description) | Non-breaking augmentation |
| `2.0.0` | Major change (different agent assignments, restructured flow) | Breaking structural change |

Note: The catalog uses integer version numbers (`1`, `2`, `3`), not semver. The `version` field in the JSON body can use semver for documentation purposes, but catalog versioning is a simple incrementing integer.

## Hard Rules

1. **Never remove without checking references.** Search all workflows for `ref: "<name>"` first.
2. **Never set active without reviewing the diff.** When moving to a new version, always diff first.
3. **Never modify an old version.** Versions are immutable. Create a new version instead.
4. **Prefer deactivation over removal.** Only remove when the user explicitly demands it and references are clear.
5. **Always record promotion rationale.** Why was this promoted? What evidence supports it?
6. **Always verify agents exist** before registering. Invalid agent names in a catalog entry will cause runtime failures.

## Anti-Patterns

- Promoting every ephemeral composition → catalog bloat; only promote proven, reusable patterns
- Skipping the diff before setting active → unintended changes become the default
- Removing a definition that workflows reference → broken workflows
- Creating a new version for cosmetic changes (whitespace, comments) → version noise (checksum-dedup handles this)
- Setting a definition active without listing the catalog first to check for conflicts → name collisions

## Integration

- **Primary agent:** Orchestrator (also self-improve for catalog cleanup)
- **MCP tools:** `catalog_create_version`, `catalog_set_active`, `catalog_get_version`, `catalog_list_versions`, `catalog_diff`, `catalog_export_version`, `catalog_deactivate`, `catalog_remove`
- **Triggered by:** `dynamic-compose` (post-run promotion), manual user request, `self-improve` (catalog hygiene)
- **Output:** catalog entries (runtime DB + optional file export)
- **See also:** `dynamic-compose` skill, `orchestrate` skill