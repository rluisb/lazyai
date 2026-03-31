<rule>
  <scope>always</scope>
  <description>Project overview, conventions, and context loading guide</description>
</rule>

# [YOUR_PROJECT_NAME] — AI Agent Rules

> This file is read at the start of every AI session.
> Keep it accurate. Keep it current. Treat it like code.
> Mirror content to CLAUDE.md for Claude/pi compatibility.

---

## Project Overview

<!-- One paragraph. What this project does, who uses it, why it exists. -->

[YOUR_PROJECT_OVERVIEW]

**Stack:**
- Language: [YOUR_LANGUAGE]
- Framework: [YOUR_FRAMEWORK]
- Database: [YOUR_DATABASE]
- ORM/Query: [YOUR_ORM]
- Testing: [YOUR_TEST_FRAMEWORK]
- Package manager: [YOUR_PACKAGE_MANAGER]

---

## Codebase Map

| Path | Responsibility |
|------|---------------|
| [YOUR_PATH_1] | [WHAT_IT_DOES] |
| [YOUR_PATH_2] | [WHAT_IT_DOES] |
| [YOUR_PATH_3] | [WHAT_IT_DOES] |
| [YOUR_SHARED_PATH] | Shared utilities — check all importers before editing |
| [YOUR_INFRA_PATH] | Infrastructure — read-only for AI agents |

---

## Decision Tree — What to Load

<!-- CRITICAL: Only load documents relevant to your current task.
     Do NOT load all documentation at once.
     This tree tells you exactly what to read based on what you're doing. -->

### Writing code for a feature
- Read: `docs/features/NNN-*/tasks/NNN-current-task.md` (your task)
- Read: `docs/features/NNN-*/techspec.md` (architecture)
- Read: `docs/standards/` relevant pattern file
- Read: `docs/rules/code-style.md`
- Do NOT read: PRD, research, ADRs (Builder follows the plan, doesn't re-evaluate it)

### Researching a topic
- Read: `docs/KNOWLEDGE_MAP.md` (orientation)
- Read: `docs/standards/` (existing patterns to be aware of)
- Do NOT read: any feature/bugfix/refactor artifacts (avoid bias — discover, don't confirm)

### Writing a PRD
- Read: `docs/features/NNN-*/research.md`
- Read: `docs/templates/prd-template.md`
- Do NOT read: `docs/standards/` (PRD is WHAT/WHY, not HOW)

### Writing a TechSpec
- Read: `docs/features/NNN-*/research.md` + `prd.md`
- Read: `docs/templates/techspec-template.md`
- Read: `docs/standards/` relevant patterns
- Read: `docs/rules/` relevant rules
- Read: `docs/adrs/` related past decisions

### Writing or modifying tests
- Read: `docs/standards/test-patterns.md`
- Read: `docs/rules/testing.md`
- Read: the implementation file being tested

### Reviewing code
- Read: `docs/rules/review.md`
- Read: `docs/rules/code-style.md`
- Read: `docs/standards/` relevant pattern
- Do NOT read: PRD or research (review the code, not the plan)

### Fixing a bug
- Read: `docs/bugfixes/NNN-*/research.md`
- Read: `docs/rules/testing.md`
- Read: `docs/standards/` relevant pattern

### Handling tech debt
- Read: `docs/tech-debt/NNN-*/techspec.md`
- Read: `docs/adrs/` related decisions
- Read: `docs/standards/` relevant pattern

### Making an architecture decision
- Read: `docs/adrs/` existing ADRs (understand past decisions)
- Read: `docs/templates/adr-template.md`
- Read: `docs/standards/` (understand current patterns)

### Don't know yet
- Read: `docs/KNOWLEDGE_MAP.md` (orient yourself)
- Ask the human what you're doing before loading more context

---

## Conventions

### Naming
- [YOUR_NAMING_CONVENTION]

### Error Handling
- [YOUR_ERROR_PATTERN]

### API Responses
- [YOUR_API_CONVENTION]

### Imports
- [YOUR_IMPORT_ORDER]

---

## Do NOT

- Never modify `[YOUR_MIGRATIONS_PATH]` without explicit human approval
- Never commit `.env` or any file containing secrets
- Never disable or delete a test to make the suite pass
- Never bypass `[YOUR_STRICT_MODE]`
- Never change `[YOUR_SHARED_PATH]` without checking all importers first
- Never run destructive database commands without explicit instruction
- Never push directly to `[YOUR_PROTECTED_BRANCH]`

---

## Workflow Rules

### Task Sizing
- Under 20 lines → implement directly
- 20–100 lines → list affected files, wait for confirmation
- Over 100 lines → write a plan, wait for approval

### Before Every Non-Trivial Task
1. State the goal in one sentence
2. List files you expect to touch
3. List what you will NOT touch
4. Wait for confirmation

### Session Management
- New task = new session
- Compact after 15–20 exchanges
- One task = one session = clean context

---

## Testing

- All new code requires tests
- Test location: `[YOUR_TEST_PATH]`
- Run: `[YOUR_TEST_COMMAND]`
- Minimum coverage: `[YOUR_COVERAGE_THRESHOLD]`%

---

## Key Commands

```bash
[YOUR_INSTALL_COMMAND]     # Install dependencies
[YOUR_TEST_COMMAND]        # Run tests
[YOUR_LINT_COMMAND]        # Run linter
[YOUR_DEV_COMMAND]         # Start dev server
[YOUR_BUILD_COMMAND]       # Build
```

---

## Session Start Checks

<!-- Run these at the start of EVERY session. Non-negotiable. -->

Before doing any work:
1. **Sync check:** If both AGENTS.md and CLAUDE.md exist, verify they are identical. If they differ → flag immediately. Do not proceed until resolved.
2. **Context check:** Read this file's Decision Tree. Load ONLY what your task needs.
3. **Standards check:** If you're about to write code, check if a relevant standard exists in `docs/standards/`. Read it before writing.

---

## Recovery Procedures

If AI-generated code causes issues after merging:

1. **Revert** the commit. Atomic commits (one task = one commit) make this safe.
2. **Create** a bugfix entry in `docs/bugfixes/NNN-description/`
3. **Impact Check:** What rule or standard was missing that allowed the bad output?
4. **Fix the gap:** Add the missing rule or standard BEFORE re-attempting
5. **Retry** using Ralph Loop (different model reviews the fix)
6. **Document** in `docs/memory/` what went wrong for future prevention

If AGENTS.md or rules are corrupted:
```bash
git checkout main -- AGENTS.md CLAUDE.md docs/rules/ docs/standards/
```

---

## Self-Improvement Protocol

<!-- CRITICAL: This section defines how the AI keeps project knowledge current.
     Every agent MUST follow this after completing any work. -->

### After Every Task — Impact Check

Before ending any session, ask yourself:

```
Did my work change any of the following?
├── Project structure (new modules, moved files)     → update Codebase Map above
├── API contracts (new/changed endpoints)             → update docs/standards/coding/
├── Architecture decisions                            → create ADR in docs/adrs/
├── Testing patterns (new test type, new fixture)     → update docs/standards/testing/
├── Dependencies (added/removed/upgraded)             → update Stack section above
├── Build/test/lint commands                          → update Key Commands above
├── Security patterns (auth, validation)              → update docs/standards/security/
├── Error handling approach                           → update docs/standards/coding/
├── New code pattern not in standards                 → create new standard
├── Existing standard's reference file changed        → update the standard
├── Feature completed/status changed                  → update docs/KNOWLEDGE_MAP.md
└── Workflow process changed                          → update docs/rules/workflow.md
```

If YES to any: **flag it before ending the session.**

Output format:
```
## 📋 Knowledge Updates Needed
- [ ] [file to update] — [what changed and why]
- [ ] [file to update] — [what changed and why]
```

The human decides whether to update now or create a follow-up task.

### Severity of Updates

| Severity | When | Action |
|----------|------|--------|
| **Immediate** | Change breaks an existing rule or standard | Update NOW before ending session |
| **Flag** | Change introduces something new not yet documented | Flag for human — update in same PR or next session |
| **Note** | Minor improvement opportunity spotted | Write to docs/memory/ for future consideration |

### What Gets Updated Where

| Change Type | Update Target |
|-------------|--------------|
| New module or directory | Root AGENTS.md (codebase map) + KNOWLEDGE_MAP.md |
| New API pattern | docs/standards/coding/api-patterns.md |
| New architecture pattern | docs/standards/architecture/ + ADR if non-obvious |
| New test pattern | docs/standards/testing/ |
| Changed conventions | docs/rules/ relevant file |
| New feature started/completed | docs/KNOWLEDGE_MAP.md |
| Architecture decision made | docs/adrs/NNN-*.md |
| Bug revealed missing rule | docs/rules/ + docs/memory/ |
| Refactor changed structure | Root AGENTS.md + docs/standards/ + KNOWLEDGE_MAP.md |
