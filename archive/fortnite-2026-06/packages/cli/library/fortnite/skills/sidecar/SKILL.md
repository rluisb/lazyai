---
name: sidecar
description: External workspace manager — discover specs relevant to the repo you're working in. Lives outside all repos at the parent directory level.
trigger: /sidecar
skill_path: skills/sidecar
scripts:
  - name: sidecar-init.sh
    description: Interactive workspace setup — scan siblings, select repos, generate .sidecar.yml
    path: scripts/sidecar-init.sh
  - name: sidecar-add.sh
    description: Add a repo to the workspace
    path: scripts/sidecar-add.sh
  - name: sidecar-query.sh
    description: Query SQLite index for specs relevant to a repo
    path: scripts/sidecar-query.sh
  - name: sidecar-index.sh
    description: Rebuild the spec-to-repo SQLite index
    path: scripts/sidecar-index.sh
---

# Sidecar — External Workspace Manager

## Purpose
Discover which specs are relevant to the repo you're currently working in. The sidecar lives outside all repos at the parent directory level — nothing is installed in any repo.

**Use when:**
- Starting work in a repo — "what specs touch this codebase?"
- Loading context before implementation — inject relevant specs
- Cross-repo impact analysis — "which specs span fedora and creator-checkout?"
- Setting up a new workspace — `sidecar init`
- Any agent needs workspace-aware context injection

## Scripts

| Script | Purpose | Key Flags |
|--------|---------|-----------|
| `sidecar` | CLI wrapper — dispatches to the scripts below | `help` |
| `sidecar-init.sh` | Interactive workspace setup | `--sidecar <dir>` |
| `sidecar-add.sh` | Add repo to workspace | `<repo>`, `--path <path>` |
| `sidecar-query.sh` | Query specs for a repo | `<repo>`, `--json` |
| `sidecar-index.sh` | Rebuild content index; auto-recovers corrupt DB | `--force` |

## Workflow

### Setup (one-time)

If no `.sidecar.yml` exists in the workspace:

```bash
# Interactive setup — scans parent dir siblings, user multi-selects
skills/sidecar/scripts/sidecar init

# Or add repos individually
skills/sidecar/scripts/sidecar add <repo-name>
```

### Context Injection (every session)

#### Step 1: Discover Workspace
Walk up from CWD to find `.sidecar.yml`. If not found, guide user to run `sidecar init`.

```bash
# Discovery is automatic — scripts walk up from CWD
```

#### Step 2: Query Relevant Specs
```bash
# Get specs relevant to current repo as JSON
skills/sidecar/scripts/sidecar query <current-repo> --json

# Example output:
# {"repo":"fedora","specs":[{"slug":"001-creator-billing-abstraction","title":"...","confidence":1.0,"match_source":"filename","path":"specs/001-creator-billing-abstraction"}]}
```

If no specs are returned, proceed without sidecar context — do not fail.

#### Step 3: Inject Context
Load top-N spec files into agent context, prioritized by confidence:
1. Read `spec.md` from each matched spec directory
2. Optionally read `plan.md`, `research.md` if they exist
3. Format as "Sidecar Context" block in working context

#### Context Injection Format
```markdown
## Sidecar Context — [current-repo]

**Workspace:** [parent-dir] (.sidecar.yml)
**Sidecar:** [sidecar-dir]/

### Relevant Specs

| Spec | Confidence | Match |
|------|-----------|-------|
| [slug] | [score] | [source] |

### Spec Summaries
**[slug]:** [first paragraph of spec.md goal section]
```

### Maintenance

```bash
# Rebuild index after creating/editing specs
skills/sidecar/scripts/sidecar index

# Force rebuild even if fresh
skills/sidecar/scripts/sidecar index --force

# If index.db is corrupt, sidecar-index.sh warns to stderr,
# deletes/recreates the DB, and rebuilds from scratch

# Add a new repo to workspace
skills/sidecar/scripts/sidecar add <repo-name>
```

## Path Constraints

All paths stored in `.sidecar.yml` must be relative to the workspace root (where `.sidecar.yml` lives). The following fields reject absolute paths and `~` (tilde) home-directory references:

- `sidecar` — the sidecar directory path
- `repos[].path` — repo directory paths
- `settings.spec_dir` — specs directory path

Absolute paths (starting with `/`) and tilde paths (starting with `~`) are rejected at validation time.

## Integration with Other Skills

| Skill | When | How |
|-------|------|-----|
| storm-scout | Phase 0 (Grill Me) | Query for existing specs before clarifying |
| build-mode | Phase 0 (Context Load) | Load all relevant specs before implementation |
| zero-point | Pre-flight gate | Check scope matches spec claims |
| drift-scope | Step 1 (Parse Spec) | Discover specs claiming to cover current repo |
| ricochet | Step 3 (Update Spec) | Verify updated spec still linked correctly |
| compact-blueprint | Step 1 (Create Spec) | Discover related specs for cross-reference |

## Tips

- Run `sidecar query` at session start for automatic context injection
- Re-index after creating new specs: `sidecar index`
- Use `--json` flag for programmatic consumption by other scripts
- The sidecar is external — it never modifies repo files
- If `.sidecar.yml` is missing, guide user through `sidecar init`
- If query returns no specs, proceed without sidecar context — do not fail
