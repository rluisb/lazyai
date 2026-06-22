# Self-Assessment Scorecard

> Evaluate project or personal readiness for AI-assisted development. Run monthly or before new project starts.

## The Four Points

- **WHAT** — A short checklist that scores readiness across six dimensions.
- **HOW** — Answer yes/no/partial per question; tally; prioritize gaps.
- **What I DON'T want** — Bureaucratic scoring with no action; comparing against teams you don't share context with.
- **How we VALIDATE** — The lowest-scoring dimension produces a concrete next task in the backlog.

## Dimensions

| # | Dimension | What it measures |
|---|-----------|------------------|
| 1 | **Tooling** | Terminal-first tool, rule files, model access |
| 2 | **Context Engineering** | Project memory, handoffs, decisions log |
| 3 | **Extensibility** | MCP servers, hooks, plugins |
| 4 | **Parallelism** | Git worktrees, multi-agent workflows, review automation |
| 5 | **Vibe Coding** | Tests runnable by agents, small functions, SRP, greppable names |
| 6 | **Operations** | CI, lint, deployment, human gates for risky ops |

## Scoring

Each question scored:

- **0** — Not present, no plan.
- **1** — Partial, ad-hoc, or only some projects.
- **2** — Consistent, documented, and the default.

Max per dimension: 10. Max total: 60.

## Questions (Personal / Team)

### 1. Tooling

- [ ] I use a supported terminal-first AI tool (Claude Code, OpenCode, or Pi) as my daily driver.
- [ ] Every active repo has the context/adapters its tool can consume (`CLAUDE.md`, `AGENTS.md`, `.claude/`, `.opencode/`, or `.pi/skills` as applicable).
- [ ] I know which model to use for which task (fast vs. smart vs. cheap).
- [ ] I have API keys or local model access ready.
- [ ] I can spin up a new project with agent context in under 5 minutes.

### 2. Context Engineering

- [ ] Every repo has a `docs/` directory with decisions, gotchas, and research.
- [ ] I write handoff documents between sessions or agents.
- [ ] I use wikilinks or structured references between related notes.
- [ ] I keep a canonical source of truth (not scattered Slack threads).
- [ ] Agents can find relevant past decisions without asking me.

### 3. Extensibility

- [ ] I have MCP servers configured for at least one supported tool where MCP is verified.
- [ ] Hooks or guardrails prevent accidental dangerous commands where the tool supports runtime hooks.
- [ ] I can add a new skill, agent, or hook without modifying core code, and I know which surfaces support each artifact.
- [ ] My setup works across machines (synced config).
- [ ] I review new MCP servers before installing (pin versions, no blind npx).

### 4. Parallelism

- [ ] I use git worktrees for parallel feature branches.
- [ ] I have automated PR creation from agent branches.
- [ ] I can run multiple agents on different worktrees simultaneously.
- [ ] CI runs on every push, not just PRs.
- [ ] I have a review workflow that doesn't bottleneck on one person.

### 5. Vibe Coding

- [ ] Tests are runnable by the agent (not just human-readable).
- [ ] Functions are small (<20 lines), files are focused (<200 lines).
- [ ] Naming is greppable and explicit; no abbreviations without context.
- [ ] Comments explain why, not what.
- [ ] I refactor before adding features, not after.

### 6. Operations

- [ ] CI passes before merge; lint, typecheck, and tests are automated.
- [ ] Deployment is one command or one button.
- [ ] Risky operations (prod DB, secrets) require human approval.
- [ ] Rollback plan exists for every deploy.
- [ ] I measure time-to-recovery, not just uptime.

## Maturity Levels

| Score | Level | Meaning |
|-------|-------|---------|
| 0–15 | Reactive | Fixing problems as they surface. |
| 16–30 | Coordinated | Some systems in place; inconsistent across projects. |
| 31–45 | Optimized | Defaults are good; most projects follow the playbook. |
| 46–60 | Strategic | Continuous improvement; new team members onboard quickly. |

## Evidence Rule

Every `2` must cite a file, command, CI check, or observed workflow. Every `1` must name what is missing. A score without evidence is a vanity score.

## Next Step Rule

After scoring, pick **the lowest dimension** with the most 0s. Create exactly one task that moves the highest-impact 0 to a 1. Do not try to fix everything at once.
