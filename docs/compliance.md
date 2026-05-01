<rule>
  <scope>manual</scope>
  <description>AI compliance and audit trail documentation — load when auditors ask about AI oversight</description>
</rule>

# AI-Assisted Development — Compliance & Audit Trail

## How AI Is Used in This Project

AI coding agents assist developers with research, planning, code generation, review, and documentation. **Every AI action is human-supervised and traceable.**

## Human Oversight Model

No AI-generated code reaches production without human review at multiple gates:

```
Research ──GATE──▶ PRD ──GATE──▶ TechSpec ──GATE──▶ Tasks ──GATE──▶ Code ──GATE──▶ Review ──GATE──▶ Merge
           │              │              │               │            │              │
       Human reviews  Human approves  Human approves  Human approves  Human reviews  Human merges
       accuracy       scope           architecture    plan            code quality   to main
```

## Audit Trail

Every feature, bugfix, refactor, and tech debt item produces a `progress.md` file that records:

- **Which AI agent** performed each step (Scout, Planner, Builder, Reviewer, Red-Team)
- **What context** the agent was given (files loaded)
- **What files** were read and changed
- **What decisions** were made
- **What the human approved** at each gate
- **Timestamps** for every entry

**Location:** `docs/features/NNN-*/progress.md` (or bugfixes/, refactors/, tech-debt/)

## How to Read the Audit Trail

Each entry in progress.md follows this format:

```
### [YYYY-MM-DD HH:MM] — [Step Name] ([Agent Name])
- Agent: [which AI agent role]
- Session: new (clean context, no carryover from previous work)
- Context loaded: [exactly which files the AI read]
- Files changed: [exactly which files were modified]
- Output: [what artifact was produced]
- Decisions: [what choices were made]
- Status: ✅ Complete (human approved)
```

## Safeguards

| Safeguard | How It Works |
|-----------|-------------|
| **Human gates** | AI stops and waits for human approval between every phase |
| **Path access control** | AI cannot write to protected paths (migrations, infra, CI) |
| **Blocked commands** | Destructive commands are listed in rules and AI is instructed to never run them |
| **Cross-model review** | Reviewer agent uses a different model than the Builder to catch blind spots |
| **Atomic commits** | One task = one commit. Easy to revert any AI-generated change |
| **Standards enforcement** | Reviewer runs detection commands from project standards on every review |

## Configuration Files

All AI behavior is defined in version-controlled files:

| File | Purpose |
|------|---------|
| `AGENTS.md` | Canonical project overview, conventions, decision tree |
| `docs/rules/*.md` | Prescriptive rules (what to do / not do) |
| `docs/standards/**/*.md` | Descriptive patterns (how we do it, with real code references) |
| `.opencode/`, `.claude/`, `.github/`, `.vscode/` | Supported tool project config, skills, prompts, and MCP files |

All changes to these files require tech lead approval (see `CODEOWNERS`).
