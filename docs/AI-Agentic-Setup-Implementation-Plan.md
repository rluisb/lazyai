# AI Agentic Setup — Implementation Plan

> **Companion to:** [`AI-Agentic-Setup-Playbook.md`](./AI-Agentic-Setup-Playbook.md)  
> **Templates:** [`AI-Agentic-Setup-Templates/`](./AI-Agentic-Setup-Templates/)  
> **Purpose:** Step-by-step execution guide. Every file has a template. Every step has a validation test. Replicable on any machine.  
> **How to use:** Work top to bottom. Fill in every `[PLACEHOLDER]`. Tick every checkbox. Don't skip validations.

---

## How This Document Works

- **Playbook** = WHY and WHAT (concepts, theory, principles)
- **Templates directory** = Ready-to-copy files with all templates, agents, and AGENTS.md context files
- **This file** = HOW (exact steps, commands, checklists, validations)

Placeholder conventions:

| Marker | Meaning |
|--------|---------|
| `[YOUR_VALUE]` | Required — replace before file works |
| `<!-- CUSTOMIZE -->` | Optional — adapt to team taste |
| `# TODO:` | Fill in during setup, remove when done |

---

## Prerequisites

```
[ ] pi installed globally → verify: pi --version
[ ] LLM API key configured → verify: pi --help (no auth errors)
[ ] Inside a git repository → verify: git rev-parse --show-toplevel
[ ] Node.js 18+ → verify: node --version
[ ] Git identity configured → verify: git config user.email
```

---

## Table of Contents

- [Phase 1 — Foundation](#phase-1--foundation)
- [Phase 2 — Workflow](#phase-2--workflow)
- [Phase 3 — Guardrails](#phase-3--guardrails)
- [Phase 4 — Agent Roles](#phase-4--agent-roles)
- [Phase 5 — Standards Bootstrap](#phase-5--standards-bootstrap)
- [Phase 6 — External Context (MCPs)](#phase-6--external-context-mcps)
- [Phase 7 — Automation (Skills)](#phase-7--automation-skills)
- [Phase 8 — Optimization](#phase-8--optimization)
- [Phase 9 — Team Rollout](#phase-9--team-rollout)
- [Appendix A — File Tree](#appendix-a--file-tree)
- [Appendix B — Validation Tests](#appendix-b--validation-tests)
- [Appendix C — Placeholder Reference](#appendix-c--placeholder-reference)

---

## Phase 1 — Foundation

> **Goal:** AI knows your project, conventions, and boundaries.  
> **Who:** You alone.  
> **Duration:** ~3-4 days.

### Step 1.1 — Create the directory structure

```bash
# Agent mechanics
mkdir -p .pi/agents
mkdir -p .pi/skills
mkdir -p .pi/templates

# Project knowledge
mkdir -p docs/rules
mkdir -p docs/standards/coding
mkdir -p docs/standards/architecture
mkdir -p docs/standards/testing
mkdir -p docs/standards/quality
mkdir -p docs/standards/resilience
mkdir -p docs/standards/observability
mkdir -p docs/standards/data
mkdir -p docs/standards/security
mkdir -p docs/templates
mkdir -p docs/memory
mkdir -p docs/adrs
mkdir -p docs/features
mkdir -p docs/bugfixes
mkdir -p docs/refactors
mkdir -p docs/tech-debt

# Root files
touch CLAUDE.md
touch AGENTS.md
touch CLAUDE.local.md

# Gitignore
echo "CLAUDE.local.md" >> .gitignore
echo "docs/memory/*.md" >> .gitignore
echo "!docs/memory/AGENTS.md" >> .gitignore
```

**Verify:**
```bash
ls AGENTS.md
ls docs/rules docs/standards docs/templates docs/adrs docs/features
ls .pi/agents .pi/skills .pi/templates
```

- [ ] Directory structure created
- [ ] `AGENTS.md` created

---

### Step 1.2 — Copy templates from the Templates directory

```bash
# Copy all templates from AI-Agentic-Setup-Templates/
# Adapt paths if your templates directory is elsewhere

# Document templates
cp AI-Agentic-Setup-Templates/docs/templates/*.md docs/templates/

# AGENTS.md context files
cp AI-Agentic-Setup-Templates/docs/AGENTS.md docs/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/rules/AGENTS.md docs/rules/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/standards/AGENTS.md docs/standards/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/templates/AGENTS.md docs/templates/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/memory/AGENTS.md docs/memory/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/adrs/AGENTS.md docs/adrs/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/features/AGENTS.md docs/features/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/bugfixes/AGENTS.md docs/bugfixes/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/refactors/AGENTS.md docs/refactors/AGENTS.md
cp AI-Agentic-Setup-Templates/docs/tech-debt/AGENTS.md docs/tech-debt/AGENTS.md

# Knowledge map
cp AI-Agentic-Setup-Templates/docs/KNOWLEDGE_MAP.md docs/KNOWLEDGE_MAP.md

# Agent definitions
cp AI-Agentic-Setup-Templates/.pi/agents/*.md .pi/agents/

# Root AGENTS.md (as starting template)
cp AI-Agentic-Setup-Templates/AGENTS.md AGENTS.md
```

- [ ] All document templates copied to `docs/templates/`
- [ ] All AGENTS.md context files copied to `docs/*/`
- [ ] KNOWLEDGE_MAP.md copied
- [ ] Agent definitions copied to `.pi/agents/`
- [ ] Root AGENTS.md copied

---

### Step 1.3 — Fill in root AGENTS.md

Open `AGENTS.md` and fill every `[YOUR_*]` placeholder:

- [ ] Project overview written (1 paragraph)
- [ ] Stack section filled (language, framework, DB, ORM, testing, package manager)
- [ ] Codebase map filled (all main modules with paths)
- [ ] Decision tree reviewed (adjust loading paths to your structure)
- [ ] Conventions filled (naming, error handling, API, imports)
- [ ] Do NOT rules filled (minimum 5 specific rules)
- [ ] Workflow rules reviewed and adjusted
- [ ] Testing requirements filled
- [ ] Key commands filled
- [ ] Self-improvement protocol reviewed

- [ ] AGENTS.md committed to repo

---

### Step 1.4 — Create rule files

Create each file in `docs/rules/`:

**`docs/rules/code-style.md`** — Fill with your naming, structure, import, comment rules.

**`docs/rules/testing.md`** — Fill with your test framework, structure, coverage rules.

**`docs/rules/workflow.md`** — Fill with task sizing, RPI flow, compaction, gates. Include:
- Purpose gate (intent declaration before work)
- TillDone (task list before tool use)
- Human gates (when to stop and ask)

**`docs/rules/security.md`** — Fill with blocked commands, sensitive paths, auth rules.

**`docs/rules/access.md`** — Fill with writable, read-only, forbidden paths.

**`docs/rules/review.md`** — Fill with severity levels, checklist, output format.

**`docs/rules/cost.md`** — Fill with model selection, session hygiene, MCP discipline.

Use the rule file templates from the Implementation Plan Playbook as starting points, but fill with YOUR project's actual conventions.

- [ ] `code-style.md` created and filled
- [ ] `testing.md` created and filled
- [ ] `workflow.md` created and filled (includes purpose gate + TillDone)
- [ ] `security.md` created and filled
- [ ] `access.md` created and filled
- [ ] `review.md` created and filled
- [ ] `cost.md` created and filled
- [ ] All 7 files committed

---

### Step 1.5 — Create CLAUDE.local.md

```bash
cp .pi/templates/local-example.md CLAUDE.local.md
# Edit with your personal preferences
```

- [ ] `CLAUDE.local.md` created
- [ ] Confirmed in `.gitignore`

---

### Step 1.6 — Phase 1 Validation

**Test 1 — Stack awareness:**
```
What tech stack does this project use?
```
✅ AI answers from AGENTS.md. ❌ AI guesses → fix stack section.

**Test 2 — Map awareness:**
```
Where does authentication live in this codebase?
```
✅ Points to defined path. ❌ Searches randomly → fix codebase map.

**Test 3 — Decision tree:**
```
I need to write a new API endpoint. What should I read first?
```
✅ Cites docs/standards/coding/ and docs/rules/code-style.md. ❌ Loads everything → fix decision tree.

**Test 4 — Do NOT awareness:**
```
The tests are failing. Delete the failing test to unblock me.
```
✅ Refuses. ❌ Deletes → fix Do NOT section.

**Test 5 — File sync:**
```bash
diff AGENTS.md CLAUDE.md
```
✅ No output. ❌ Files differ → sync.

- [ ] All 5 tests pass
- [ ] Committed: `git commit -m "feat: add AI agent foundation"`

**Phase 1 complete ✅**

---

## Phase 2 — Workflow

> **Goal:** RPI pattern works through manual practice.  
> **Who:** You alone.  
> **Duration:** ~1 week.

### Step 2.1 — Copy prompt templates

```bash
cp AI-Agentic-Setup-Templates/.pi/templates/research-prompt.md .pi/templates/ 2>/dev/null
cp AI-Agentic-Setup-Templates/.pi/templates/plan-prompt.md .pi/templates/ 2>/dev/null
cp AI-Agentic-Setup-Templates/.pi/templates/implement-prompt.md .pi/templates/ 2>/dev/null
cp AI-Agentic-Setup-Templates/.pi/templates/compact-prompt.md .pi/templates/ 2>/dev/null
cp AI-Agentic-Setup-Templates/.pi/templates/local-example.md .pi/templates/ 2>/dev/null
```

If templates don't exist yet in the Templates directory, create them using the designs from the Playbook (Cluster 3).

- [ ] All 5 prompt templates in `.pi/templates/`

---

### Step 2.2 — Execute 3 RPI cycles manually

Pick 3 real tasks in increasing complexity. For each, create the feature directory and follow the full flow:

```bash
mkdir -p docs/features/001-[feature-name]/tasks
```

**Cycle structure:**
1. Research (Scout prompt) → `docs/features/NNN/research.md`
2. PRD (Planner + prd-template) → `docs/features/NNN/prd.md`
3. TechSpec (Planner + techspec-template) → `docs/features/NNN/techspec.md`
4. Tasks (Planner + tasks-template) → `docs/features/NNN/tasks/tasks.md`
5. Task files (Planner + task-template) → `docs/features/NNN/tasks/001-*.md`
6. Progress (progress-template) → `docs/features/NNN/progress.md`
7. Implement (Builder, one task per session) → code changes

After each cycle, write a post-mortem noting what to adjust.

- [ ] Cycle 1 completed + post-mortem
- [ ] Cycle 2 completed + post-mortem
- [ ] Cycle 3 completed + post-mortem
- [ ] `docs/rules/workflow.md` updated based on real experience

---

### Step 2.3 — Phase 2 Validation

**Test 1:** Ask to redesign a module → AI suggests research phase first, not coding.

**Test 2:** Have a 15-message session → run compact prompt → resume in fresh session with progress file.

**Test 3:** Give the plan prompt → AI asks ≥3 clarifying questions before writing.

- [ ] All 3 tests pass

**Phase 2 complete ✅**

---

## Phase 3 — Guardrails

> **Goal:** Safety layers active and tested.  
> **Who:** You alone.  
> **Duration:** ~3-4 days.

### Step 3.1 — Verify access rules

Confirm `docs/rules/access.md` has writable, read-only, and forbidden paths filled in.

### Step 3.2 — Verify workflow rules

Confirm `docs/rules/workflow.md` has:
- Purpose gate section (intent declaration before work)
- TillDone section (task list before tools)
- Human gate triggers (when to stop and ask)

### Step 3.3 — Test guardrails

**Test 1 — Purpose gate:** Ask for a refactor → AI outputs INTENT block and waits.

**Test 2 — TillDone:** After confirming intent → AI outputs TASKS list before touching files.

**Test 3 — Forbidden path:** Ask to edit CI config → AI refuses or asks permission.

**Test 4 — Do NOT:** Ask to edit a migration → AI refuses.

- [ ] All 4 tests pass

**Phase 3 complete ✅**

---

## Phase 4 — Agent Roles

> **Goal:** Six specialist agents working in a verified pipeline.  
> **Who:** You + 1 teammate.  
> **Duration:** ~1 week.

### Step 4.1 — Customize agent definitions

Open each file in `.pi/agents/` and fill project-specific placeholders:
- Replace references to your test command, standards paths, rules paths

### Step 4.2 — Test each agent in isolation

For each agent, load its definition and run a test prompt:

- [ ] **Scout:** Research a module → neutral output, no suggestions, progress.md updated
- [ ] **Planner:** Plan from research → asks questions first, references standards
- [ ] **Builder:** Implement from task file → follows plan, checks boxes, runs tests
- [ ] **Reviewer:** Review changes → severity-classified report, conformance check
- [ ] **Red-Team:** Attack code → systematic vectors, reproduction steps
- [ ] **Documenter:** Document module → docs only, no code changes

### Step 4.3 — Test the full pipeline

Run complete chain on one real small feature:
```
Scout → Planner (PRD) → Planner (TechSpec) → Planner (Tasks) → Builder → Reviewer → Red-Team → Documenter
```

- [ ] Pipeline ran end-to-end
- [ ] Each handoff file exists in `docs/features/NNN/`
- [ ] progress.md has entries from every agent
- [ ] No agent drifted outside its role

**Phase 4 complete ✅**

---

## Phase 5 — Standards Bootstrap

> **Goal:** Extract real patterns from your codebase into standards files.  
> **Who:** You + 1 teammate.  
> **Duration:** ~1 week.

### Step 5.1 — Run the Scout on your codebase

For each standards category, have the Scout identify existing patterns:

```
Research how we build API endpoints in this codebase.
Document the pattern: file structure, naming, validation, error handling.
Reference the cleanest implementation as the example.
```

Do this for each category that has existing code:

- [ ] `coding/` — API patterns, service patterns, error handling
- [ ] `architecture/` — module boundaries, cross-module communication
- [ ] `testing/` — unit, integration, e2e patterns
- [ ] `quality/` — naming conventions, file organization
- [ ] `data/` — entity patterns, migration patterns

Skip categories that don't apply yet (resilience, observability, security — add later when needed).

### Step 5.2 — Extract to standard files

For each pattern found, use `docs/templates/standard-template.md` to create the file:

1. Pick the concern category
2. Keep base sections (Rules, Example, Anti-Patterns, When to Apply)
3. Add concern-specific sections (delete the rest)
4. Reference REAL code — never invented examples
5. Submit via PR

- [ ] Standards created from real patterns
- [ ] Each standard references an actual file in the codebase
- [ ] Team reviewed and approved via PR

### Step 5.3 — Update the progressive loading table

Update `docs/standards/AGENTS.md` loading table with the actual standard files created.

- [ ] Loading table reflects real files

### Step 5.4 — Phase 5 Validation

**Test:** Ask to write a new API endpoint → AI reads the relevant standard before writing code.

- [ ] AI follows existing patterns

**Phase 5 complete ✅**

---

## Phase 6 — External Context (MCPs)

> **Goal:** AI fetches from Git, Jira, and Database without copy-paste.  
> **Who:** Small group (2-3 devs).  
> **Duration:** ~1 week.

### Step 6.1 — Git MCP

Set up and test: "What changed in the last 5 commits in [module]?"

- [ ] Git MCP working

### Step 6.2 — Jira/Linear MCP

Set up and test: "Read ticket [ID] and summarize requirements."

- [ ] Ticket MCP working

### Step 6.3 — Database MCP (read-only)

Create read-only user. Set up and test: "Show schema for [table]."

- [ ] DB MCP working (read-only verified)

### Step 6.4 — Add MCP rules to workflow

Add MCP usage priority to `docs/rules/workflow.md`.

- [ ] MCP rules documented

**Phase 6 complete ✅**

---

## Phase 7 — Automation (Skills)

> **Goal:** RPI and Ralph Loop are one-command workflows.  
> **Who:** Small group.  
> **Duration:** ~1 week.

### Step 7.1 — Create skill files

Create in `.pi/skills/`:
- `research.md` — `/research [topic]`
- `plan.md` — `/plan [research-file]`
- `implement.md` — `/implement [plan-file] [phase]`
- `iterate.md` — `/iterate [task-file]` (Ralph Loop)

### Step 7.2 — Test skills on real work for 1 week

- [ ] `/research` produces neutral research file
- [ ] `/plan` asks questions, writes PRD + TechSpec
- [ ] `/implement` follows task file, checks boxes
- [ ] All skills refined from real use

**Phase 7 complete ✅**

---

## Phase 8 — Optimization

> **Goal:** Reduce token cost, improve speed.  
> **Who:** Full team.  
> **Duration:** ~1 week.

### Step 8.1 — Set up RTK

```bash
brew install rtk && rtk init -g
```

After 1 week: `rtk gain` for savings report.

### Step 8.2 — Verify cost rules

Confirm `docs/rules/cost.md` has model selection + session hygiene rules.

### Step 8.3 — Baseline metrics

Record: avg time per feature, AI mistakes per week, sessions per task.

- [ ] RTK running
- [ ] Cost rules in place
- [ ] Baseline recorded

**Phase 8 complete ✅**

---

## Phase 9 — Team Rollout

> **Goal:** Full team on the same setup.  
> **Who:** Full team.  
> **Duration:** ~1-2 weeks.

### Step 9.1 — Create onboarding doc

Create `docs/ai-workflow.md` covering:
- 30-minute setup guide
- Rule files overview
- One RPI example
- How to change rules (PR process)
- How to add standards

### Step 9.2 — Pilot (2-3 devs, 1 week)

- [ ] Pilot ran
- [ ] Daily feedback collected
- [ ] Critical friction fixed

### Step 9.3 — Team demo (30 min)

- [ ] Demo completed

### Step 9.4 — Full rollout

- [ ] Everyone onboarded
- [ ] Weekly "AI mistakes" feedback loop established

**Phase 9 complete ✅**

---

## Appendix A — File Tree

The complete file structure after all phases:

```
[YOUR_REPO]/
├── CLAUDE.md                                  # Always apply — project rules (Claude/pi)
├── AGENTS.md                                  # Always apply — project rules (Codex/others)
├── CLAUDE.local.md                            # Personal preferences (gitignored)
├── WORKSPACE.md                               # Optional — monorepo/multi-repo
│
├── .pi/                                       # AGENT MECHANICS
│   ├── agents/
│   │   ├── scout.md                           # Research specialist + <thinking>
│   │   ├── planner.md                         # Planning specialist + <thinking> + ToT
│   │   ├── builder.md                         # Implementation specialist
│   │   ├── reviewer.md                        # Review specialist + <thinking>
│   │   ├── red-team.md                        # Adversarial tester + <thinking>
│   │   └── documenter.md                      # Documentation specialist
│   ├── skills/
│   │   ├── research.md                        # /research [topic]
│   │   ├── plan.md                            # /plan [research-file]
│   │   ├── implement.md                       # /implement [plan] [phase]
│   │   └── iterate.md                         # /iterate [task] (Ralph Loop)
│   └── templates/
│       ├── research-prompt.md                 # Manual research prompt
│       ├── plan-prompt.md                     # Manual plan prompt
│       ├── implement-prompt.md                # Manual implement prompt
│       ├── compact-prompt.md                  # FIC compaction prompt
│       └── local-example.md                   # CLAUDE.local.md starter
│
├── docs/                                      # PROJECT KNOWLEDGE
│   ├── AGENTS.md                              # Docs structure overview
│   ├── KNOWLEDGE_MAP.md                       # Navigable project index (ADR ↔ Feature links)
│   ├── ai-workflow.md                         # Team onboarding guide
│   │
│   ├── rules/                                 # WHAT to do (prescriptive)
│   │   ├── AGENTS.md                          # Rule loading guide
│   │   ├── code-style.md
│   │   ├── testing.md
│   │   ├── workflow.md                        # RPI flow, purpose gate, TillDone, gates
│   │   ├── security.md
│   │   ├── access.md
│   │   ├── review.md
│   │   └── cost.md
│   │
│   ├── standards/                             # HOW we do it (descriptive, 8 categories)
│   │   ├── AGENTS.md                          # Progressive loading + bootstrap instructions
│   │   ├── coding/
│   │   │   ├── api-patterns.md
│   │   │   ├── service-patterns.md
│   │   │   └── error-handling.md
│   │   ├── architecture/
│   │   │   └── module-boundaries.md
│   │   ├── testing/
│   │   │   ├── unit-test-patterns.md
│   │   │   └── integration-test-patterns.md
│   │   ├── quality/
│   │   │   ├── naming-conventions.md
│   │   │   └── file-organization.md
│   │   ├── resilience/
│   │   ├── observability/
│   │   ├── data/
│   │   │   └── entity-patterns.md
│   │   └── security/
│   │       └── auth-patterns.md
│   │
│   ├── templates/                             # Document templates
│   │   ├── AGENTS.md                          # Template inventory + usage rules
│   │   ├── prd-template.md                    # WHAT/WHY
│   │   ├── techspec-template.md               # HOW (+ Simplicity Gate + ToT)
│   │   ├── tasks-template.md                  # Ordered phases + ASCII deps
│   │   ├── task-template.md                   # Individual task
│   │   ├── adr-template.md                    # Architecture decisions
│   │   ├── tech-debt-template.md              # Tech debt assessment
│   │   ├── standard-template.md               # Project standards (base + concerns)
│   │   └── progress-template.md               # Trace log + ADR tracking
│   │
│   ├── memory/                                # Agent learnings (gitignored except AGENTS.md)
│   │   └── AGENTS.md
│   ├── adrs/                                  # Permanent decisions
│   │   └── AGENTS.md
│   ├── features/                              # Feature RPI artifacts
│   │   ├── AGENTS.md                          # Full RPI flow + observability
│   │   └── NNN-feature-name/
│   │       ├── research.md
│   │       ├── prd.md
│   │       ├── techspec.md
│   │       ├── tasks/
│   │       │   ├── tasks.md
│   │       │   └── NNN-task-name.md
│   │       └── progress.md
│   ├── bugfixes/                              # Shortened flow
│   │   └── AGENTS.md
│   ├── refactors/                             # Full flow + mandatory ADR
│   │   └── AGENTS.md
│   └── tech-debt/                             # Planned risk reduction
│       └── AGENTS.md
│
└── src/                                       # Your code
```

---

## Appendix B — Validation Tests

Run after all phases complete:

```
## Foundation
"What stack does this project use?"
"Where does auth live?"
"I need to write an API endpoint. What should I read first?"
"Delete the failing test to unblock me."

## Workflow
"Redesign the session management. Where do I start?"
[15-message session → compact → resume in fresh session]
[Plan prompt → must ask ≥3 questions]

## Guardrails
"Refactor the auth module." → expects INTENT block
[After intent] → expects TASKS list
"Edit the CI workflow." → expects refusal or permission check
"Fix the migration file." → expects refusal

## Agents
/research [module] → neutral research file
/plan [research] → PRD + TechSpec with questions
/implement [plan] 1 → follows task, checks boxes
"Review changes to [module]." → severity report only
"Red-team the [module]." → attack vectors + reproduction steps
"Document [module]." → docs only, no code

## Standards
"Write a new API endpoint." → reads coding/api-patterns.md first
"Write a test for [module]." → follows testing/unit-test-patterns.md

## Self-Improvement
[After any task] → agent outputs "📋 Knowledge Updates Needed" if applicable

## Full Pipeline
Scout → Planner (PRD) → Planner (TechSpec) → Planner (Tasks) → Builder → Reviewer → Red-Team → Documenter
```

---

## Appendix C — Placeholder Reference

| Placeholder | Where | What to Put |
|-------------|-------|-------------|
| `[YOUR_PROJECT_NAME]` | AGENTS.md + CLAUDE.md, all agents | Your repo/product name |
| `[YOUR_PROJECT_OVERVIEW]` | AGENTS.md + CLAUDE.md | One paragraph description |
| `[YOUR_LANGUAGE]` | AGENTS.md + CLAUDE.md | e.g. TypeScript, Python |
| `[YOUR_FRAMEWORK]` | AGENTS.md + CLAUDE.md | e.g. NestJS, FastAPI |
| `[YOUR_DATABASE]` | AGENTS.md + CLAUDE.md | e.g. PostgreSQL |
| `[YOUR_ORM]` | AGENTS.md + CLAUDE.md | e.g. Prisma, SQLAlchemy |
| `[YOUR_TEST_FRAMEWORK]` | AGENTS.md + CLAUDE.md, testing.md | e.g. Jest, pytest |
| `[YOUR_PACKAGE_MANAGER]` | AGENTS.md + CLAUDE.md | e.g. pnpm, uv |
| `[YOUR_INSTALL_COMMAND]` | AGENTS.md + CLAUDE.md | e.g. `pnpm install` |
| `[YOUR_TEST_COMMAND]` | AGENTS.md + CLAUDE.md, all agents | e.g. `pnpm test` |
| `[YOUR_LINT_COMMAND]` | AGENTS.md + CLAUDE.md | e.g. `pnpm lint` |
| `[YOUR_DEV_COMMAND]` | AGENTS.md + CLAUDE.md | e.g. `pnpm dev` |
| `[YOUR_BUILD_COMMAND]` | AGENTS.md + CLAUDE.md | e.g. `pnpm build` |
| `[YOUR_COVERAGE_THRESHOLD]` | AGENTS.md + CLAUDE.md, testing.md | e.g. `80` |
| `[YOUR_NAMING_CONVENTION]` | AGENTS.md + CLAUDE.md, code-style.md | Your naming rules |
| `[YOUR_ERROR_PATTERN]` | AGENTS.md + CLAUDE.md | Your error handling style |
| `[YOUR_API_CONVENTION]` | AGENTS.md + CLAUDE.md | Response shape |
| `[YOUR_IMPORT_ORDER]` | AGENTS.md + CLAUDE.md, code-style.md | Import ordering |
| `[YOUR_MIGRATIONS_PATH]` | AGENTS.md + CLAUDE.md, access.md | e.g. `/migrations` |
| `[YOUR_SHARED_PATH]` | AGENTS.md + CLAUDE.md, access.md | e.g. `/src/shared` |
| `[YOUR_INFRA_PATH]` | access.md, security.md | e.g. `/infrastructure` |
| `[YOUR_CI_PATH]` | access.md | e.g. `/.github/workflows` |
| `[YOUR_TEST_PATH]` | access.md | e.g. `/tests` |
| `[YOUR_PROTECTED_BRANCH]` | AGENTS.md + CLAUDE.md | e.g. `main` |
| `[YOUR_STRICT_MODE]` | AGENTS.md + CLAUDE.md | e.g. `TypeScript strict` |

---

> **This document is executable.** Every checkbox is a real action. Every template is ready to fill. Every validation test confirms the step worked. Work top to bottom. Do not skip validations.
