<workspace-protocol>

# Workspace Protocol — Multi-Repo Awareness

A "workspace" is a directory that contains one or more sibling repositories plus a planning repo that orchestrates work across them. AI tooling lives at the **workspace root**, not inside any single repo, so a single agent session can reason across all of them.

This protocol governs how skills and agents move between repos, where they record state, and how standards propagate.

---

## 1. Workspace topology

```
~/work/lazyai-workspace/                  ← workspace root (where AI tools live)
├── .claude/                                 ← Claude Code config (workspace-scope)
├── .github/                                 ← Copilot config (workspace-scope)
├── .specify/                                ← shared spec-kit memory + templates
│   ├── memory/
│   │   ├── constitution.md                  ← workspace-wide constitution
│   │   └── repos/
│   │       ├── creator-checkout/
│   │       │   ├── ledger.md                ← per-repo durable record
│   │       │   └── last-known-state.md      ← per-repo session pointer
│   │       ├── fedora/
│   │       └── school-plan-service/
│   ├── templates/                           ← spec/plan/tasks templates
│   └── scripts/
├── specs/                                   ← workspace-level cross-cutting specs
├── creator-checkout/                        ← code repo (Next.js)
├── fedora/                                  ← code repo (Rails)
└── school-plan-service/                     ← code repo (Go)
```

**The two repo roles.**
- **Planning repo.** Holds shared specs, the workspace constitution, ledgers, and the AI tool config. Often the workspace root itself, sometimes a dedicated `*-planning/` sibling.
- **Code repos.** Hold actual product code. They MAY have their own `.specify/` for repo-local specs; they MAY also be governed only by the workspace.

---

## 2. Where AI tools live

**Rule.** Tool configuration installs at the workspace root, not inside a code repo.

| Tool | Workspace location |
|---|---|
| Claude Code | `<workspace>/.claude/` |
| Copilot | `<workspace>/.github/` |
| OpenCode | `<workspace>/.opencode/` |

**Why.** A workspace-rooted install lets agents read and write across every sibling repo in one session without re-bootstrapping. Project-rooted installs duplicate config, drift apart, and force agents to lose state when crossing repo boundaries.

**Exception.** A code repo MAY have its own project-scope install when it leaves the workspace (e.g., open-sourced). The project-scope install is then independent.

---

## 3. Per-repo ledgers

Every code repo gets a durable record in `.specify/memory/repos/<repo-name>/`:

### `ledger.md`
Append-only log of decisions, completed work, and open follow-ups for that repo.

```markdown
## 2026-04-27 — feat(auth): rotate session tokens

**Who:** ricardo + claude
**Plan:** specs/2026-04-auth-rotation/plan.md
**Status:** completed
**Verified:** test suite green, manual smoke OK

### Decisions
- Stored tokens in encrypted column rather than separate table — simpler, single transaction.
- Rejected library `xyz` because [...].

### Follow-ups
- [ ] Add metric for rotation failure rate (next sprint).
```

### `last-known-state.md`
Single-page snapshot a future session reads first. Replace, don't append.

```markdown
# Last Known State — fedora

**Updated:** 2026-04-27 by ricardo+claude
**Branch:** feat/auth-rotation
**Dirty files:** none (committed at HEAD)
**Paused at:** session 3 of plan, ready for task-007
**Open question:** confirm rotation interval with security team
```

**Update obligations.**
- `update-memory` skill MUST append the ledger and rewrite `last-known-state.md` at the end of every workflow.
- A new session MUST read `last-known-state.md` for the active repo as its first action.

---

## 4. Standards propagation cascade

Standards live at three scopes. Lower scopes override higher ones; conflicts trigger an explicit resolution step.

```
global standards (~/.claude/, ~/.specify/)
        ↓ (default for everything the user works on)
workspace standards (<workspace>/specs/standards/)
        ↓ (overrides for this team / org)
project standards (<repo>/specs/standards/)
        ↓ (overrides for this codebase)
code (the actual files)
```

**How agents apply it.**
- Read in cascade order: global → workspace → project. Most specific wins.
- The `extract-standards` skill writes to the **most specific** scope that applies (per-repo for repo patterns, workspace for org patterns, global for universal patterns).
- A standard at one scope that contradicts a higher scope MUST be marked explicitly (`# Overrides: workspace/coding/api.md`).

---

## 5. Spec-kit structure detection

When `ai-setup compile` runs in a workspace, it detects pre-existing structures and never overwrites them.

**Detection rules.**

| If detected | Behavior |
|---|---|
| `<workspace>/.specify/` exists | Never overwrite `memory/`, `templates/`, `scripts/`. Add only missing `repos/<name>/` ledgers. |
| `<repo>/specs/###-slug/` flat dirs | Never restructure. Add per-feature `AGENTS.md` only if missing. |
| Both `.specify/` and flat `specs/` | Treat them as parallel surfaces — write to both per the user's existing choice. |
| Neither | Scaffold the spec-kit-aligned structure fresh. |

**Why this matters.** Many users adopted spec-kit before ai-setup; their handcrafted memory MUST survive a re-run of `ai-setup compile`.

---

## 6. Cross-repo workflows

When a workflow spans multiple repos (e.g., a feature that touches `fedora` + `creator-checkout`):

1. The workspace-level `specs/###-slug/` holds the cross-cutting spec.
2. Each affected repo gets a ledger entry pointing back to the workspace spec.
3. The orchestrator agent decomposes the work per-repo and assigns workers (one per repo).
4. The synthesizer agent merges results and writes a final cross-repo ledger entry.

**Permission boundary.** A worker writing to repo A MUST NOT write to repo B. Repo isolation is enforced by the agent's working directory contract — the orchestrator switches contexts explicitly.

---

## 7. Frontmatter declaration

Skills that touch workspace state declare it in frontmatter:

```yaml
workspace:
  scope: [workspace, project]    # or [project] only
  reads: [.specify/memory/repos/<active>/last-known-state.md]
  writes: [.specify/memory/repos/<active>/ledger.md]
  cross_repo: false              # true only for orchestrator-driven workflows
```

Agents executing a workflow MUST honor these declarations — no surprise writes outside the declared paths.

</workspace-protocol>
