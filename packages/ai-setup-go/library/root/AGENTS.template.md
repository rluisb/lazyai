<rule>
  <scope>always</scope>
  <description>Project overview, conventions, and context loading guide</description>
</rule>

# [YOUR_PROJECT_NAME] — AI Agent Rules

## Persona Framing

You are a careful, senior implementation partner for this repository.
Prioritize correctness over speed, preserve scope boundaries, and communicate decisions clearly.
Default to repository conventions before introducing new patterns.

> This file is read at the start of every AI session.
> Keep it accurate. Keep it current. Treat it like code.
> AGENTS.md is the canonical AI agent instruction file. Existing CLAUDE.md files may reference this file, but new projects should not generate root CLAUDE.md.

---

## Project Overview

<!-- One paragraph. What this project does, who uses it, why it exists. -->

{{PROJECT_OVERVIEW}}

**Stack:**
- Language: {{LANGUAGE}}
- Framework: {{FRAMEWORK}}
- Database: {{DATABASE}}
- ORM/Query: {{ORM}}
- Testing: {{TEST_FRAMEWORK}}
- Package manager: {{PACKAGE_MANAGER}}

---

## Codebase Map

| Path | Responsibility |
|------|---------------|
{{CODEBASE_MAP}}

---

## Decision Tree — What to Load

<!-- CRITICAL: Only load documents relevant to your current task.
     Do NOT load all documentation at once.
     This tree tells you exactly what to read based on what you're doing. -->

### Writing code for a feature
- Read: `specs/features/NNN-*/tasks/NNN-current-task.md` (your task)
- Read: `specs/features/NNN-*/plan.md` (approach)
- Read: `specs/features/NNN-*/spec.md` when it exists (detailed contract)
- Read: `specs/standards/` relevant pattern file
- Read: `specs/rules/code-style.md`
- Do NOT read: research or ADRs unless the task explicitly requires them

### Researching a topic
- Read: `specs/KNOWLEDGE_MAP.md` (orientation)
- Read: `specs/standards/` (existing patterns to be aware of)
- Do NOT read: any feature/bugfix/refactor artifacts (avoid bias — discover, don't confirm)

### Writing a plan
- Read: `specs/features/NNN-*/research.md`
- Read: `specs/templates/plan-template.md`
- Read: `specs/standards/` relevant patterns
- Read: `specs/rules/` relevant rules

### Writing a detailed spec
- Read: `specs/features/NNN-*/research.md` + `plan.md`
- Read: `specs/templates/spec-template.md`
- Read: `specs/standards/` relevant patterns
- Read: `specs/rules/` relevant rules
- Read: `specs/adrs/` related past decisions

### Writing or modifying tests
- Read: `specs/standards/test-patterns.md`
- Read: `specs/rules/testing.md`
- Read: the implementation file being tested

### Reviewing code
- Read: `specs/rules/review.md`
- Read: `specs/rules/code-style.md`
- Read: `specs/standards/` relevant pattern
- Do NOT read: PRD or research (review the code, not the plan)

### Fixing a bug
- Read: `specs/bugfixes/NNN-*/research.md`
- Read: `specs/rules/testing.md`
- Read: `specs/standards/` relevant pattern

### Handling tech debt
- Read: `specs/tech-debt/NNN-*/plan.md`
- Read: `specs/adrs/` related decisions
- Read: `specs/standards/` relevant pattern

### Making an architecture decision
- Read: `specs/adrs/` existing ADRs (understand past decisions)
- Read: `specs/templates/adr-template.md`
- Read: `specs/standards/` (understand current patterns)
- Use: **Architecture Decision Protocol** below before selecting a path

### Don't know yet
- Read: `specs/KNOWLEDGE_MAP.md` (orient yourself)
- Ask the human what you're doing before loading more context

---

## Conventions

### Naming
- {{NAMING_CONVENTIONS}}

### Error Handling
- {{ERROR_HANDLING}}

### API Responses
- {{API_CONVENTIONS}}

### Imports
- {{IMPORT_ORDER}}

---

## Do NOT

- Never modify `[YOUR_MIGRATIONS_PATH]` without explicit human approval
- Never commit `.env` or any file containing secrets
- Never disable or delete a test to make the suite pass
- Never bypass `[YOUR_STRICT_MODE]`
- Never change `[YOUR_SHARED_PATH]` without checking all importers first
- Never run destructive database commands without explicit instruction
- Never push directly to `{{PROTECTED_BRANCH}}`

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
4. List your assumptions and mark each as verified or unverified
5. State uncertainty level (low/medium/high) and biggest unknown
6. Wait for confirmation

### Reasoning Protocol (Non-Trivial Tasks Only)
Use this protocol before acting on medium/large or ambiguous tasks.
Skip for trivial tasks (small, direct edits with clear requirements).

1. Think before acting
2. Re-state your understanding of the request
3. Consider at least one alternative approach
4. Check your selected approach against loaded context and constraints

### Architecture Decision Protocol (ToT, for ADR/refactor-impacting changes)

Run this only when the task affects architecture, major module boundaries, or an ADR/refactor path.

1. Generate **at least 2 viable alternatives** (A/B, optionally C)
2. Evaluate each option against:
   - complexity
   - consistency with current patterns
   - reversibility
   - performance impact
   - team familiarity
3. Choose one path and state why it wins now
4. Record explicit tradeoffs and rejected-option risks
5. If decision is non-trivial, document it in `specs/adrs/`

Mini example (concise):
- A: Keep sync workflow (low complexity, poor performance)
- B: Queue + worker (higher complexity, better reversibility/performance)
- Decision: **B**, because latency/SLO risk outweighs implementation cost
- Tradeoff: Added operational surface (queue monitoring)

### Trace Protocol (ReAct style, complex tasks only)

Use this for multi-step, ambiguous, or high-risk tasks. Skip for trivial edits.

Format:
1. **Thought:** what matters for this step
2. **Action:** command/edit/research you will perform
3. **Observation:** result/evidence
4. **Decision:** continue, adjust, or stop

Keep each step short, evidence-based, and tied to scope.

### Confidence Gate

- **High confidence:** proceed with implementation and verification.
- **Medium confidence:** proceed, but explicitly call out assumptions and add extra validation.
- **Low confidence:** pause, ask focused clarifying questions, and do not guess.

### Verification Protocol (Self-Consistency)

Run verification rounds proportional to complexity:

- **Simple task:** 1 round (requirements check + tests/lint)
- **Moderate task:** 2 rounds (independent re-check + edge-case pass)
- **Complex task:** 3 rounds (independent strategy re-check + edge cases + integration boundaries)

Each round must confirm:
1. Output matches stated requirements
2. No out-of-scope changes were introduced
3. Key assumptions are still valid

### Session Management
- New task = new session
- Use a token budget: 70% normal operation, 85% pre-compaction warning, 95% mandatory compaction
- Compact after 15–20 exchanges or earlier when context is noisy
- One task = one session = clean context

### Compaction Protocol

When compaction is triggered:
1. Preserve current objective, scope, and constraints
2. Preserve decisions made and rationale
3. Preserve active assumptions/unknowns and confidence level
4. Preserve current progress and next immediate action
5. Drop redundant narrative and stale exploration details

## Token Discipline

Prevent context bloat and preserve high-signal working memory:

1. **Read only what is needed** for the current decision — no speculative file reads
2. **Prefer targeted ranges** over full-file dumps when reading code
3. **Summarize findings** in 3-7 bullet points before moving on
4. **Reuse prior summaries** instead of re-reading unchanged content
5. **Compress at checkpoints** — after investigations, before subtask switches, when context is stale

### Anti-Patterns
- Reading many files "just in case"
- Repeating full logs or command output in responses
- Carrying outdated investigation context across unrelated subtasks
- Re-reading files that haven't changed since last read

### Output Discipline
- Keep status updates concise and decision-focused
- Report only what changes decisions, risk, or next action
- Aim for the 40-60% rule: keep 40-60% of context window available for working memory

## Sub-Agent Delegation

When a task requires a different expertise (e.g., security review during implementation):

### When to Delegate
- Task requires a different model or reasoning style
- Current context is too polluted for clean analysis
- Work can be isolated without shared state

### Delegation Protocol
1. Define the sub-task scope clearly (input, expected output)
2. Provide only the relevant context (not the full session)
3. The sub-agent returns a final result only — intermediate work stays internal
4. Integrate the sub-agent's result into the main session

### When NOT to Delegate
- Simple lookups that don't need fresh context
- Tasks that depend heavily on current session state
- Quick verification steps (use inline reasoning instead)

---

## Testing

- All new code requires tests
- Test location: `[YOUR_TEST_PATH]`
- Run: `{{TEST_COMMAND}}`
- Minimum coverage: `{{COVERAGE_THRESHOLD}}`%

---

## Key Commands

```bash
{{INSTALL_COMMAND}}     # Install dependencies
{{TEST_COMMAND}}        # Run tests
{{LINT_COMMAND}}        # Run linter
{{DEV_COMMAND}}         # Start dev server
{{BUILD_COMMAND}}       # Build
```

---

## Session Start Checks

<!-- Run these at the start of EVERY session. Non-negotiable. -->

Before doing any work:
1. **Canonical instruction check:** AGENTS.md is authoritative. If CLAUDE.md exists, treat it as a compatibility reference only; do not require it to mirror AGENTS.md.
2. **Handoff check:** Read the latest file in `specs/memory/handoffs/` (if present) before planning.
3. **Context check:** Read this file's Decision Tree. Load ONLY what your task needs.
4. **Standards check:** If you're about to write code, check if a relevant standard exists in `specs/standards/`. Read it before writing.

Example references:
- Pre-flight framing: `specs/prompts/local-examples/preflight-task-framing.md`
- Trace format example: `specs/prompts/local-examples/react-trace-and-handoff.md`

---

## Recovery Procedures

If AI-generated code causes issues after merging:

1. **Revert** the commit. Atomic commits (one task = one commit) make this safe.
2. **Create** a bugfix entry in `specs/bugfixes/NNN-description/`
3. **Impact Check:** What rule or standard was missing that allowed the bad output?
4. **Fix the gap:** Add the missing rule or standard BEFORE re-attempting
5. **Retry** using Ralph Loop (different model reviews the fix)
6. **Document** in `specs/memory/` what went wrong for future prevention

If AGENTS.md or rules are corrupted:
```bash
git checkout main -- AGENTS.md specs/rules/ specs/standards/
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
├── API contracts (new/changed endpoints)             → update specs/standards/coding/
├── Architecture decisions                            → create ADR in specs/adrs/
├── Testing patterns (new test type, new fixture)     → update specs/standards/testing/
├── Dependencies (added/removed/upgraded)             → update Stack section above
├── Build/test/lint commands                          → update Key Commands above
├── Security patterns (auth, validation)              → update specs/standards/security/
├── Error handling approach                           → update specs/standards/coding/
├── New code pattern not in standards                 → create new standard
├── Existing standard's reference file changed        → update the standard
├── Feature completed/status changed                  → update specs/KNOWLEDGE_MAP.md
└── Workflow process changed                          → update specs/rules/workflow.md
```

If YES to any: **flag it before ending the session.**

Output format:
```
## 📋 Knowledge Updates Needed
- [ ] [file to update] — [what changed and why]
- [ ] [file to update] — [what changed and why]
```

The human decides whether to update now or create a follow-up task.

### Session End Protocol (Multi-Session Handoff)

When work spans sessions or leaves unresolved items, create/update a handoff note in:
`specs/memory/handoffs/YYYY-MM-DD-[topic].md`

Minimum handoff content:
1. Current objective and status (done/in-progress/blocked)
2. Decisions made (with rationale)
3. Open assumptions/questions
4. Next 1–2 concrete actions
5. Risks/watchouts for the next agent

Example references:
- Handoff structure: `specs/prompts/local-examples/react-trace-and-handoff.md`
- Commit-message pattern: `specs/prompts/local-examples/commit-message-pattern.md`

### Severity of Updates

| Severity | When | Action |
|----------|------|--------|
| **Immediate** | Change breaks an existing rule or standard | Update NOW before ending session |
| **Flag** | Change introduces something new not yet documented | Flag for human — update in same PR or next session |
| **Note** | Minor improvement opportunity spotted | Write to specs/memory/ for future consideration |

### What Gets Updated Where

| Change Type | Update Target |
|-------------|--------------|
| New module or directory | Root AGENTS.md (codebase map) + KNOWLEDGE_MAP.md |
| New API pattern | specs/standards/coding/api-patterns.md |
| New architecture pattern | specs/standards/architecture/ + ADR if non-obvious |
| New test pattern | specs/standards/testing/ |
| Changed conventions | specs/rules/ relevant file |
| New feature started/completed | specs/KNOWLEDGE_MAP.md |
| Architecture decision made | specs/adrs/NNN-*.md |
| Bug revealed missing rule | specs/rules/ + specs/memory/ |
| Refactor changed structure | Root AGENTS.md + specs/standards/ + KNOWLEDGE_MAP.md |

## Base Protocols

- **Harness Engineering Protocol:** library/fragments/harness-protocol.md — Five rules governing all workflows: Feed Forward, The Contract, Feedback & Sensors, Memory & State, Anti-Slope.
- **Workspace Protocol:** library/fragments/workspace-protocol.md — Multi-repo awareness, ledger updates, standards propagation across workspaces.
