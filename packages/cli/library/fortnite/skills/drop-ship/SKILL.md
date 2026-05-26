---
name: drop-ship
skill_path: skills/drop-ship
description: End-of-feature PR release workflow. Auto-detects build artifacts, runs impact analysis, creates PRs with commit conventions, watches for reviews, and orchestrates merge and deploy. 6-phase release pipeline.
trigger: /drop-ship
triggers:
  - "ship this"
  - "create PR"
  - "release this feature"
  - "deploy this"
  - "merge and deploy"
scripts:
  - name: detect-artifacts.sh
    description: Auto-detect build artifacts, spec files, and task completions
    path: scripts/detect-artifacts.sh
  - name: impact-scan.sh
    description: Run impact analysis on changed files and downstream dependencies
    path: scripts/impact-scan.sh
  - name: pr-create.sh
    description: Create PR with conventional commits and auto-generated description
    path: scripts/pr-create.sh
  - name: review-watch.sh
    description: Monitor PR for comments, classify feedback, and trigger fixes
    path: scripts/review-watch.sh
  - name: merge-deploy.sh
    description: Orchestrate merge, tag, and deploy with approval gates
    path: scripts/merge-deploy.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | End-of-feature PR release, merge and deploy |
| **Do not use when** | Implementation, review |
| **Primary agent** | rift-deploy |
| **Runtime risk** | Medium — PR creation, deploy orchestration |
| **Outputs** | PR description, impact analysis, deploy status |
| **Validation** | Build artifacts, approval gates |
| **Deep mode trigger** | `/drop-ship` or release workflow |

# Drop Ship — The Battle Bus Exit

> *"When the build is solid and the squad is ready — drop from the bus and claim the Victory Royale."*

## Purpose

The **Drop Ship** is the end-of-feature release pipeline. When wall-builder finishes implementation and shield-audit gives the green light, Drop Ship takes over: it detects what was built, analyzes the blast radius, runs final quality gates, creates a production-ready PR, watches the review cycle, and orchestrates merge + deploy.

**Use when:**
- A feature branch is ready for review and merge
- You need to "ship this" — create a PR from current work
- Post-implementation verification is complete and it's time to release
- A PR has been reviewed and needs feedback addressed before merge
- You want to automate the full release pipeline from branch to production

**What it automates:**
- Artifact detection (specs, tasks, tests, build outputs)
- Impact analysis (what files changed, what could break)
- Quality gate validation (lint, test, build, ledger integrity)
- PR creation with conventional commits and structured descriptions
- Review monitoring (classify comments, trigger fixes, track approvals)
- Merge orchestration (squash, tag, changelog)
- Deploy handoff to rift-deploy with preflight checks

---

## ⚠️ Human Gate Protocol

Every phase that mutates GitHub state requires explicit human approval. The agent presents a plan, the human says "go", and only then does the agent execute.

| Phase | Mutation | Gate |
|-------|----------|------|
| 4: PR Creation | `gh pr create` | Human reviews PR title, description, reviewers → approves |
| 5: Review Watch | Posting comments | Human reviews each queued response → approves individually |
| 5: Review Watch | Triggering fixes | Human reviews the fix plan → approves dispatch to wall-builder |
| 6: Merge | `gh pr merge` | Human executes merge themselves — agent never runs this |

## Phases

### Phase 1: Detect — Loot Scan

**What:** Automatically discover all artifacts produced during the feature cycle.

**Key Actions:**
1. Scan the current branch for changes against `main`
2. Detect spec files (`bee-gone/specs/<NNN-slug>/SPEC.md`)
3. Detect task files (`.specify/tasks/`, `TODO.md`, `TASKS.md`)
4. Detect test files (new/modified tests, coverage reports)
5. Detect build artifacts (`dist/`, `build/`, `.next/`, container images)
6. Detect documentation changes (`README.md`, `CHANGELOG.md`, `docs/`)
7. Detect configuration changes (`.env*`, `docker-compose*`, `k8s/`)

**Output:** `artifacts.json` — structured manifest of everything found.

```bash
./skills/drop-ship/scripts/detect-artifacts.sh --branch $(git branch --show-current) --base main
```

---

### Phase 2: Impact Analysis — Storm Scout

**What:** Analyze the blast radius of changes before opening the PR.

**Key Actions:**
1. Run `codegraph_impact` on all modified symbols to find downstream callers
2. Check if changes touch shared libraries, APIs, or database schemas
3. Identify services/modules that depend on changed code
4. Flag high-risk changes (auth, payment, data migration, public APIs)
5. Generate an impact report with risk rating (Low / Medium / High / Critical)

**Output:** `impact-report.json` — affected files, risk rating, recommended reviewers.

```bash
./skills/drop-ship/scripts/impact-scan.sh --diff $(git diff --name-only main)
```

**Risk Ratings:**

| Rating | Criteria | Action |
|--------|----------|--------|
| 🟢 Low | Isolated feature, no shared deps | Standard PR |
| 🟡 Medium | Touches shared utils, internal APIs | Add domain expert reviewer |
| 🟠 High | Changes public API, auth, payment | Require security review |
| 🔴 Critical | DB migration, infra change, breaking API | Require approval from tech lead + rift-deploy dry-run |

---

### Phase 3: Quality Gates — Zero Point Check

**What:** Run all pre-PR quality gates. Nothing ships without passing the Zero Point.

**Key Actions:**
1. Run repo-appropriate quality gates via `quality-gate.sh`
2. Verify ledger integrity: `ledger.sh verify`
3. Run drift-scope check: `drift-check.sh --spec <spec> --repo .`
4. Run contract-check post-conditions: `contract-check.sh --mode post`
5. Check for uncommitted changes, merge conflicts, or stale branches
6. Verify all tasks in spec are marked complete

**Gates:**

| Gate | Tool | Pass Criteria |
|------|------|---------------|
| Lint | `quality-gate.sh` | Zero new lint errors |
| Test | `quality-gate.sh` | All tests pass, coverage ≥ threshold |
| Build | `quality-gate.sh` | Production build succeeds |
| Ledger | `ledger.sh verify` | Hash chain intact |
| Drift | `drift-check.sh` | No spec violations |
| Contract | `contract-check.sh` | Post-conditions met |

**If any gate fails:** Stop. Report failures. Hand off to `reboot-van` or `wall-builder` for fixes. Do NOT create the PR.

---

### Phase 4: PR Creation — Supply Drop

**What:** Create a production-ready PR with conventional commits, structured description, and auto-assigned reviewers.

**Key Actions:**
1. Generate PR title from conventional commit format
2. Auto-generate PR description from spec, task list, and impact report
3. Link Jira ticket (extracted from branch name or spec)
4. Link related Confluence pages
5. Attach artifact manifest and impact report
6. Assign reviewers based on impact analysis (CODEOWNERS + domain experts)
7. Add labels (`feature`, `bugfix`, `breaking`, `needs-review`, risk rating)
8. Set draft status if Critical risk rating

**Commit Conventions:**

| Type | Prefix | When to Use |
|------|--------|-------------|
| Feature | `feat:` | New behavior, endpoint, UI component |
| Bugfix | `fix:` | Bug fix, regression repair |
| Refactor | `refactor:` | Code change with no behavior change |
| Test | `test:` | Adding or updating tests |
| Docs | `docs:` | Documentation only |
| Chore | `chore:` | Build, tooling, dependency updates |
| Breaking | `feat!:` or `BREAKING CHANGE:` | Incompatible API change |

**PR Description Template:**

```markdown
## 🎯 Feature Drop — <Title>

**Spec:** [NNN-slug](link-to-spec)
**Jira:** [PROJ-123](link-to-ticket)
**Risk:** 🟢 Low / 🟡 Medium / 🟠 High / 🔴 Critical

### What Changed
- <Summary from spec>

### Impact Analysis
- Files changed: N
- Risk rating: <rating>
- Affected services: <list>

### Quality Gates
- [x] Lint pass
- [x] Tests pass
- [x] Build pass
- [x] Ledger verified
- [x] Drift check pass

### Artifacts
- Spec: `bee-gone/specs/NNN-slug/SPEC.md`
- Tasks: `.specify/tasks/`
- Tests: `<test files>`

### Reviewers
- @<primary> (domain expert)
- @<secondary> (impact area)
```

```bash
./skills/drop-ship/scripts/pr-create.sh --branch $(git branch --show-current) --base main --spec bee-gone/specs/NNN-slug/SPEC.md
```

---

### Phase 5: Review Watch — Comment Queue (NOT Auto-Post)

**What:** Monitor the PR for review comments, classify feedback, and queue responses for human approval.

**Key Actions:**
1. Poll for new comments and review submissions
2. Classify each comment (Question / Change Request / Nit / Already Fixed / Out of Scope)
3. Draft a response for each using the pr-review tone guide
4. **Queue all responses — do NOT post**
5. Present the queue to the human: "I have N responses ready. Review and approve?"
6. Human approves each response individually
7. Human posts the responses (or tells agent to post with explicit per-response approval)
8. Track approval status (pending → changes requested → approved)
9. Re-run quality gates after fixes
10. Notify when PR is ready for merge

**Classification & Routing:**

| Comment Type | Action | Agent |
|--------------|--------|-------|
| Question | Draft answer with spec evidence, queue for human | drop-ship |
| Change Request (spec-aligned) | Draft fix plan, queue for human → if approved, dispatch wall-builder | drop-ship → wall-builder |
| Change Request (quality) | Draft fix plan, queue for human → if approved, dispatch wall-builder | drop-ship → wall-builder |
| Nit | Draft acknowledgment, queue for human | drop-ship |
| Out of Scope | Draft deferral with justification, queue for human | drop-ship |

**Watch Loop:**

```bash
# Start monitoring a PR (read-only polling)
./skills/drop-ship/scripts/review-watch.sh --pr <NUMBER> --repo <OWNER/REPO> --poll-interval 300

# One-time check
./skills/drop-ship/scripts/review-watch.sh --pr <NUMBER> --once
```

**Approval Gate:**
- PR must have at least 1 approval
- No unresolved change requests
- All CI checks green
- All quality gates re-passed after fixes

---

### Phase 6: Merge & Deploy — Victory Royale

**What:** Merge the PR, tag the release, and hand off to rift-deploy.

**Key Actions:**
1. Verify all approval gates met
2. Squash-merge with conventional commit message
3. Generate changelog entry from PR description
4. Create git tag (`vX.Y.Z` or `release/YYYY-MM-DD`)
5. Update `CHANGELOG.md`
6. Hand off to `rift-deploy` with deploy context
7. Monitor deploy status
8. Post-deploy health check via `respawn-crew`

**Merge Checklist:**

- [ ] All reviewers approved
- [ ] No unresolved conversations
- [ ] CI checks green
- [ ] Quality gates re-passed
- [ ] Ledger entry created for merge
- [ ] Changelog updated
- [ ] Tag created

**Deploy Handoff:**

```bash
./skills/drop-ship/scripts/merge-deploy.sh --pr <NUMBER> --tag v1.2.3 --mode staging
```

This triggers:
1. Merge PR
2. Create tag
3. Dispatch to rift-deploy: `MODE=staging PREFLIGHT=true`
4. Await deploy confirmation
5. Post-deploy health check

---

## Auto-Detection

### Artifact Detection Table

| Artifact Type | Detection Method | Location Patterns |
|---------------|------------------|-------------------|
| **Spec files** | File existence + git diff | `bee-gone/specs/<NNN-slug>/SPEC.md` |
| **Task files** | File existence | `.specify/tasks/`, `TODO.md`, `TASKS.md` |
| **QA reports** | File existence + CI artifact | `coverage/`, `test-results/`, `.rspec` |
| **Build outputs** | Directory existence | `dist/`, `build/`, `.next/`, `out/` |
| **Container images** | Docker image list | `docker images --filter reference=<repo>` |
| **PR templates** | File existence | `.github/pull_request_template.md` |
| **Changelog** | File existence | `CHANGELOG.md`, `HISTORY.md` |
| **Config changes** | Git diff | `.env*`, `docker-compose*`, `k8s/`, `terraform/` |
| **Documentation** | Git diff | `README.md`, `docs/`, `*.md` |
| **Database migrations** | File pattern | `db/migrate/`, `migrations/`, `*.sql` |

---

## Quality Gates

### Pre-PR Gate Execution

Run these in sequence. Any failure blocks PR creation.

```bash
# 1. Ledger integrity
./skills/truth-chain/scripts/ledger.sh verify

# 2. Repo quality gates
./skills/zero-point/scripts/quality-gate.sh --repo-profile <profile>

# 3. Drift check
./skills/drift-scope/scripts/drift-check.sh --spec <spec-file> --repo .

# 4. Contract post-conditions
./skills/zero-point/scripts/contract-check.sh --mode post --spec-dir <dir> --repo-profile <profile>

# 5. Impact scan
./skills/drop-ship/scripts/impact-scan.sh --diff $(git diff --name-only main)
```

### Gate Results

All results are recorded to the ledger:

```bash
./skills/truth-chain/scripts/ledger.sh append quality_gate \
  '{"gate_type":"pre-pr","repo":"<repo>","passed":true,"duration_ms":4500,"errors":0,"warnings":0}'
```

---

## PR Creation

### Commit Convention Enforcement

Drop Ship enforces conventional commits for all PRs:

```
<type>[(<scope>)][!]: <description>

[body]

[footer]
```

**Types:** `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`, `ci`

**Scopes:** Derived from changed files (e.g., `auth`, `api`, `ui`, `db`)

**Breaking changes:** Marked with `!` after type/scope or `BREAKING CHANGE:` in footer

### Reviewer Assignment

| Impact Rating | Reviewers Required |
|---------------|-------------------|
| 🟢 Low | 1 (CODEOWNERS match) |
| 🟡 Medium | 2 (domain expert + CODEOWNERS) |
| 🟠 High | 3 (security + domain + tech lead) |
| 🔴 Critical | 3 + rift-deploy dry-run approval |

---

## Review Watch

### Monitoring Modes

| Mode | Use Case | Frequency |
|------|----------|-----------|
| **Active** | High-priority PR, waiting for review | Every 5 min |
| **Passive** | Normal PR, background monitoring | Every 30 min |
| **On-demand** | Check now, then exit | One-time |

### Feedback Loop

```
New comment detected
    ↓
Classify (question / change / nit / out-of-scope)
    ↓
If change request:
    Evaluate against spec → must fix / should fix / discuss / decline
    ↓
If must fix:
    Dispatch wall-builder (MODE=standard)
    ↓
Fix committed → re-run quality gates
    ↓
All green → notify reviewers
```

---

## Integration

### With rift-deploy
- Drop Ship creates the PR and monitors review
- On approval, Drop Ship merges and dispatches to rift-deploy for deploy
- rift-deploy runs preflight and executes deploy
- Post-deploy, rift-deploy reports status back to Drop Ship

### With pr-review
- Drop Ship uses pr-review logic for initial PR description generation
- Review watch phase leverages pr-review's rating scale for classifying findings
- When reviewing OTHER squads' PRs, use pr-review directly (not drop-ship)

### With feedback-review
- Review watch phase uses feedback-review's classification system
- Change requests are evaluated against spec before triggering fixes
- Action plans from feedback-review feed into wall-builder dispatches

### With truth-chain
- Every phase logs to the ledger: artifact detection, gate results, PR creation, merge, deploy
- Ledger integrity is a mandatory gate before PR creation
- Merge events create immutable audit trail

### With zero-point
- Quality gates are executed via zero-point's `quality-gate.sh`
- Drift check runs via zero-point's contract-check and drift-scope
- Pre-flight YAGNI check prevents shipping speculative code

### With shield-audit
- Impact analysis can dispatch shield-audit for security review on High/Critical changes
- Pre-merge, shield-audit can run a final verification pass

---

## Rules

1. **Never create a PR without passing all quality gates** — the Zero Point is absolute
2. **Never merge without at least one approval** — no solo merges, ever
3. **Always use conventional commits** — PR titles and squash messages must follow the convention
4. **Always link the spec** — every PR must reference its source spec file
5. **Always run impact analysis** — know the blast radius before opening the PR
6. **Never auto-merge** — human approval required for every merge, even with green CI
7. **Always log to ledger** — every phase creates an immutable ledger entry
8. **Never skip review watch on High/Critical PRs** — monitor until merged or explicitly cancelled
9. **Always re-run gates after fixes** — a fix that breaks a gate is not a fix
10. **Deploy requires rift-deploy** — Drop Ship merges; rift-deploy deploys. Never bypass.
11. **Never post to GitHub without human approval** — all comments, reviews, and PR mutations are queued for human review first. See Human Gate Protocol above.

---

## Agent Assignment

- **Primary**: rift-deploy (orchestrates merge and deploy handoff)
- **Secondary**: shield-audit (pre-merge verification, security review)
- **Support**: wall-builder (addressing review feedback), pr-review (reviewing others), feedback-review (processing our PR comments)

## When NOT to Use

- **Hotfixes** — Use `respawn-crew` + `reboot-van` for P1 incidents (skip full pipeline)
- **Draft PRs** — Use `gh pr create --draft` manually for early WIP sharing
- **Dependency updates** — Simple `chore:` PRs may skip impact analysis
- **Documentation only** — `docs:` PRs can skip build/test gates if no code changed

---

## Scripts

| Script | Purpose | Key Flags |
|--------|---------|-----------|
| `detect-artifacts.sh` | Auto-detect build artifacts and spec files | `--branch`, `--base`, `--output` |
| `impact-scan.sh` | Run impact analysis on changed files | `--diff`, `--repo`, `--output` |
| `pr-create.sh` | Create PR with conventional commits | `--branch`, `--base`, `--spec`, `--draft` |
| `review-watch.sh` | Monitor PR for comments and feedback | `--pr`, `--repo`, `--poll-interval`, `--once` |
| `merge-deploy.sh` | Merge, tag, and hand off to deploy | `--pr`, `--tag`, `--mode` |

---

## Example: Full Drop Ship Run

```bash
# Phase 1: Detect
./skills/drop-ship/scripts/detect-artifacts.sh --branch feature/auth-v2 --base main

# Phase 2: Impact
./skills/drop-ship/scripts/impact-scan.sh --diff $(git diff --name-only main)
# → Risk: 🟠 High (touches auth middleware)

# Phase 3: Quality Gates
./skills/zero-point/scripts/quality-gate.sh --repo-profile fedora
./skills/truth-chain/scripts/ledger.sh verify
./skills/drift-scope/scripts/drift-check.sh --spec bee-gone/specs/042-auth-v2/SPEC.md --repo .

# Phase 4: Create PR
./skills/drop-ship/scripts/pr-create.sh \
  --branch feature/auth-v2 \
  --base main \
  --spec bee-gone/specs/042-auth-v2/SPEC.md \
  --reviewer @security-lead,@backend-lead

# Phase 5: Watch
./skills/drop-ship/scripts/review-watch.sh --pr 234 --repo myorg/fedora --poll-interval 300

# Phase 6: Merge & Deploy (after approval)
./skills/drop-ship/scripts/merge-deploy.sh --pr 234 --tag v2.3.0 --mode staging
```
