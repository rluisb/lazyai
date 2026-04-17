# ai-setup

<system-context>
  <project-name>ai-setup</project-name>
  <planning-dir>spec</planning-dir>
  <primary-language>Unknown</primary-language>
  <framework></framework>
  <workspace-type>project</workspace-type>
</system-context>


## AI Assistant Configuration

This project uses Claude Code with ai-setup integration.


<context-discipline>

### Context Discipline

Minimize context to reduce cost and improve accuracy.

**Before reading files**:
1. State what you need to know and why.
2. Read the minimum files to answer that question.
3. Stop reading when you have enough context.

**Context budget**:
- Read max 10 files before making a change.
- Prefer focused snippets over full files.
- If you have read 15+ files without a plan, STOP and ask the user.

**Session hygiene**:
- One task per session — start fresh for new tasks.
- If context grows large, summarize progress and offer to continue in a new session.
- Do not re-read files already discussed in this session.

**Priority order** (read in this order, stop when sufficient):
1. Task description and acceptance criteria
2. Files being modified + their tests
3. Type definitions and interfaces
4. Related standards/rules in specs/
5. ADRs only if an architectural decision is needed

**Anti-patterns**:
- ❌ Reading many files "just in case"
- ❌ Re-reading unchanged files from earlier in the session
- ❌ Loading full files when a function signature is enough

</context-discipline>




<rpi-workflow>

### RPI Workflow — Research, Plan, Implement

For non-trivial tasks, follow this structured flow:

**1. Research** — Understand before proposing.
- Identify affected files and components.
- Review existing patterns in the codebase.
- Check for related ADRs or documentation.
- **Output**: Research findings (what exists, what matters).
- ⛔ HUMAN GATE: Confirm understanding before planning.

**2. Plan** — Define approach before coding.
- Define clear acceptance criteria.
- Break work into incremental steps.
- Identify risks and mitigations.
- If multiple approaches exist, run the Decision Protocol.
- **Output**: plan.md with tasks.
- ⛔ HUMAN GATE: Approve plan before implementing.

**3. Implement** — Execute plan with continuous validation.
- One task at a time, in order.
- Write tests before or alongside code.
- Run quality gates after each change.
- Commit frequently with clear messages.

**Pivot handling**: If implementation reveals the plan is wrong:
1. STOP current work.
2. Document why the plan is no longer viable.
3. Create an ADR if the pivot affects architecture.
4. Return to Research phase with new information.

**Skip RPI for**: Trivial changes (<20 lines), typo fixes, dependency bumps, documentation-only changes.

</rpi-workflow>




<reasoning-protocol>

### Reasoning Protocol

Before acting on non-trivial tasks, show your reasoning:

<cot>
1. **Affected**: What files, functions, and tests are involved?
2. **Plan**: What is the minimum change? List concrete steps.
3. **Risks**: What could break? Edge cases to consider.
4. **Verdict**: Proceed / need clarification / blocked.
</cot>

Then implement.

**When to use**: Tasks that modify logic, architecture, or >20 lines.
**Skip for**: Renaming, formatting, typo fixes, single-line changes, adding comments.

</reasoning-protocol>




<decision-protocol>

### Decision Protocol

When multiple approaches exist, evaluate before choosing.

**Trigger**: Architecture decisions, complex refactors, technology choices, performance tradeoffs.

**Format**:

#### Approaches Considered

**Option A: [name]**
- Approach: [1-2 sentences]
- Pros: [list]
- Cons: [list]
- Effort: [low/medium/high]

**Option B: [name]**
- [same structure]

#### Decision: [chosen option]
- Rationale: [why this one wins now]
- Tradeoff accepted: [what we give up]
- Record as ADR: [yes/no — yes if it affects architecture]

**Example**:
- A: Keep sync workflow (low complexity, poor performance under load)
- B: Queue + worker (higher complexity, better scalability)
- Decision: **B** — latency/SLO risk outweighs implementation cost
- Tradeoff: Added operational surface (queue monitoring)

**Skip for**: Single obvious approach, bug fixes, style changes.

</decision-protocol>




<quality-gates>
  <description>Required quality checks before code completion</description>
  
  <gates>
    <gate name="lint" required="true">
      <description>Code style and static analysis</description>
      <commands>
        <default>npm run lint</default>
        <ruby>bundle exec rubocop</ruby>
        <go>go vet ./...</go>
      </commands>
    </gate>
    
    <gate name="typecheck" required="true">
      <description>Type safety validation</description>
      <commands>
        <typescript>npm run typecheck</typescript>
        <ruby>bundle exec srb tc</ruby>
        <go>go build ./...</go>
      </commands>
    </gate>
    
    <gate name="test" required="true">
      <description>Automated test suite</description>
      <commands>
        <default>npm test</default>
        <ruby>bundle exec rspec</ruby>
        <go>go test ./...</go>
      </commands>
    </gate>
    
    <gate name="build" required="false">
      <description>Production build verification</description>
      <commands>
        <default>npm run build</default>
      </commands>
    </gate>
  </gates>
  
  <coverage>
    <minimum-client>85</minimum-client>
    <minimum-server>90</minimum-server>
    <require-new-tests>For new features and bug fixes</require-new-tests>
  </coverage>
</quality-gates>




<git-conventions>
  <description>Repository conventions for branches, commits, and pull requests</description>

  <branches>
    <naming>{type}/{ticket}-{description}</naming>
    <examples>
      <example>feat/PROJ-123-add-login-flow</example>
      <example>fix/PROJ-456-handle-null-response</example>
      <example>chore/PROJ-789-update-dependencies</example>
    </examples>
    <rules>
      <rule>Prefer short, lowercase, hyphenated descriptions</rule>
      <rule>Keep branch scope to one primary ticket or objective</rule>
    </rules>
  </branches>

  <commits>
    <format>{type}({scope}): {description}</format>
    <types>feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert</types>
    <examples>
      <example>feat(auth): add passwordless login flow</example>
      <example>fix(api): guard against undefined payload</example>
      <example>test(checkout): cover discount edge case</example>
    </examples>
    <rules>
      <rule>Use imperative mood ("add", "fix", "update")</rule>
      <rule>Keep subject line concise and descriptive</rule>
      <rule>Reference ticket ID in body when required by team policy</rule>
    </rules>
  </commits>

  <pull-requests>
    <title-format>{type}({scope}): {summary}</title-format>
    <checklist>
      <item>Describe why the change is needed</item>
      <item>Summarize behavior changes and impacted areas</item>
      <item>List validation performed (tests, lint, typecheck, build)</item>
      <item>Include screenshots or logs for UI/behavior changes when relevant</item>
    </checklist>
  </pull-requests>
</git-conventions>




<agent-harness>

### Agent Coordination

When using multiple specialized agents, follow these coordination rules:

**Agent roles**:
| Agent | When to use | What it does | What it does NOT do |
|-------|-------------|-------------|---------------------|
| Scout | Research phase | Maps codebase, identifies patterns | Does not suggest, plan, or write code |
| Planner | Plan phase | Creates plans, asks clarifying questions | Does not implement code |
| Builder | Implement phase | Executes plan, writes code and tests | Does not add unrequested features |
| Reviewer | After implementation | Finds issues, rates severity | Does not fix code |
| Documenter | After completion | Writes docs, updates standards | Does not modify source code |

**Handoff protocol**:
- Each agent reads the previous agent's output before starting.
- Each agent writes its output to the designated location (specs/ or code).
- If an agent is blocked, it STOPS and describes the blocker — does not guess.

**Escalation**: When confidence is low, return to the human with:
1. What you know.
2. What you are uncertain about.
3. What you need to proceed.

</agent-harness>




<bug-resolution>
  <description>Structured approach to debugging and bug fixes</description>
  
  <workflow>
    <step name="reproduce">
      <action>Create minimal reproduction case</action>
      <action>Document exact steps to trigger bug</action>
      <action>Identify affected environments</action>
    </step>
    
    <step name="diagnose">
      <action>Trace execution path</action>
      <action>Identify root cause vs symptoms</action>
      <action>Check for related issues in history</action>
    </step>
    
    <step name="fix">
      <action>Write failing test that captures the bug</action>
      <action>Implement minimal fix</action>
      <action>Verify test passes</action>
    </step>
    
    <step name="verify">
      <action>Run full test suite</action>
      <action>Test edge cases</action>
      <action>Verify fix in affected environments</action>
    </step>
  </workflow>
  
  <documentation>
    <require-rca>For P1/P2 bugs or recurring issues</require-rca>
    <template>spec/specs/research/bugfix-rca-{ticket}.md</template>
  </documentation>
</bug-resolution>



## Git Safety

- **Never run `git push`** without explicit user approval — this includes `gh pr push`, `git push --force`, and any remote-writing git operation.
- Use `gh` CLI for GitHub operations: issues, PRs, reviews, checks — all read operations are allowed freely.
- Creating branches and commits locally is fine. Pushing to remote requires the user to say "push" or "push it".
- Never force-push to any branch. Never push directly to main/master.
- When creating PRs, use `gh pr create` but do NOT auto-push — ask the user first.


## Claude Code-Specific Notes

- Project settings: `.claude/settings.json`
- Modular rules: `.claude/rules/<name>.md` (supports `paths` frontmatter for scoping)
- Skills: `.claude/skills/<name>/SKILL.md`
- Agents: `.claude/agents/<name>.md`
- Personal overrides: `CLAUDE.local.md` (gitignore this)

## Project-Specific Instructions


