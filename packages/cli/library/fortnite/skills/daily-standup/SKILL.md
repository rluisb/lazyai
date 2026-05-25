---
name: daily-standup
description: Daily Slack standup generator — formats yesterday/today/blockers for async standup. Pulls Jira, Git, and PR context for accurate status.
trigger: /daily-standup
skill_path: skills/daily-standup
scripts:
  - name: generate-standup.sh
    description: Generate formatted Slack standup message
    path: scripts/generate-standup.sh
---

# Daily Standup — Slack Async Standup Generator

## Purpose
Generate formatted daily standup messages for Slack async standups. Pulls context from Jira, Git commits, and PR activity to produce accurate "yesterday/today/blockers" updates.

**Use when:**
- Daily async standup — "generate my standup"
- Need to report status — "what did I work on yesterday?"
- Preparing for Growth Team Standup meetings (Mon/Thu 10:30 ET)

## Standup Format

```
*Yesterday:*
• Completed <task> — <link to PR/ticket>
• Worked on <task> — <status>

*Today:*
• Working on <task> — <goal>
• Planning to <task>

*Blockers:*
• <blocker or "None">
```

## Scripts

| Script | Purpose |
|--------|---------|
| `generate-standup.sh` | Generate formatted Slack standup message |

## Workflow

### Step 1: Gather Yesterday's Work
Pull data from multiple sources:
- **Jira**: Issues transitioned to Done/In Progress yesterday
- **Git**: Commits from yesterday
- **PRs**: PRs opened, updated, or merged yesterday
- **Session-db**: Tasks completed in recent sessions

### Step 2: Identify Today's Work
- **Jira**: Issues currently In Progress
- **Session-db**: Active tasks from last checkpoint
- **PRs**: PRs awaiting follow-up

### Step 3: Check for Blockers
- **Jira**: Issues with "Blocked" status or blocked by links
- **PRs**: PRs stuck in review > 24h
- **Session-db**: Recorded errors or blockers

### Step 4: Format for Slack
Produce Slack-formatted output with:
- Bold section headers (`*Yesterday:*`, `*Today:*`, `*Blockers:*`)
- Bullet points with `•`
- Links to Jira tickets and PRs where available
- Concise, action-oriented language

## Integration with Other Skills

- **meeting-prep**: Standup is a subset of meeting prep
- **pr-watcher**: Check PR status for blockers
- **slurp-juice**: Pull session history for recent work
- **qmd**: Search vault for context on ongoing work

## Tips

- Run early morning before standup deadline
- Copy output directly to Slack
- Edit manually if needed before posting
- Save to `docs/standups/<date>.md` for historical tracking
