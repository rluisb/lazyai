---
name: meeting-prep
description: Meeting preparation assistant — generates structured prep materials for recurring meetings. Pulls Jira/Confluence context, PR status, recent commits, and produces a meeting-ready brief.
trigger: /meeting-prep
skill_path: skills/meeting-prep
scripts:
  - name: prep-meeting.sh
    description: Generate prep materials for a specific meeting
    path: scripts/prep-meeting.sh
  - name: weekly-brief.sh
    description: Generate weekly status brief across all active work
    path: scripts/weekly-brief.sh
---

# Meeting Prep — Structured Meeting Preparation

## Purpose
Generate structured, context-rich preparation materials for recurring meetings. Pulls data from Jira, Confluence, GitHub PRs, recent commits, and session history to produce a meeting-ready brief.

**Use when:**
- Before any recurring meeting — "prep for Growth Standup"
- Weekly sync preparation — "generate my weekly brief"
- Replenishment meeting — "prep for prioritization"
- Oncall handoff — "prep for oncall review"

## Your Meeting Schedule

| Meeting | Day | Time (ET) | Prep Focus |
|---------|-----|-----------|------------|
| Growth Team Standup | Mon, Thu | 10:30-10:50 | Yesterday/today/blockers, PR status |
| Growth Team Prioritization | Mon | 10:30-11:30 | Backlog items, priority decisions needed |
| Services Guild Meeting | Mon | 13:00-14:00 | Services architecture updates |
| FrontEnd Guild Meeting | Mon | 16:00-17:00 | Frontend patterns, shared components |
| Teachable + Stripe Sync | Tue | 09:30-10:30 MT | Integration status, blockers |
| Oncall Handoff & Operability | Tue | 11:00-11:30 | Incidents, SLOs, oncall notes |
| Growth Eng Weekly | Tue | 16:00-17:00 | Engineering deep-dive, demos |
| Project Bee-Gone Sync | Thu | 14:30-15:00 | Project milestones, deliverables |
| All Hands | 4th Thu | 12:00-13:00 | Company updates, team wins |
| Learnable | Fri | 14:00-14:30 | Learning topics, share-outs |

## Scripts

| Script | Purpose |
|--------|---------|
| `prep-meeting.sh <meeting-name>` | Generate prep for a specific meeting |
| `weekly-brief.sh` | Generate consolidated weekly status brief |

## Workflow

### Step 1: Identify the Meeting
Run `prep-meeting.sh <meeting-name>` or ask the user which meeting they're preparing for.

### Step 2: Gather Context
Pull relevant data based on meeting type:

**Standup meetings:**
- Jira issues updated in last 24-48h assigned to user
- PRs opened/updated/reviewed recently
- Recent commits (last 2 days)
- Any blockers from session-db errors

**Sync meetings:**
- Jira issues in active sprint
- Confluence pages updated recently
- PR status (open, review requested, merged)
- Milestone progress

**Prioritization/Replenishment:**
- Backlog items by priority
- Untriaged issues
- Capacity vs demand analysis
- Dependencies blocking work

**Oncall/Operability:**
- Recent incidents (P1-P4)
- SLO status
- Error budget remaining
- On-call notes from session-db

**Guild meetings:**
- Architecture decisions made
- Patterns discovered
- Cross-team dependencies
- Shared component updates

### Step 3: Generate Brief
Produce a structured brief:

```
## Meeting: <name>
**Date:** <date>
**Duration:** <duration>

### What I Worked On
- <item 1>
- <item 2>

### What I'm Working On
- <item 1>
- <item 2>

### Blockers / Risks
- <blocker 1>
- <risk 1>

### Discussion Points
- <point 1>
- <point 2>

### Decisions Needed
- <decision 1>
```

### Step 4: Review and Refine
Present the brief to the user for review. Allow edits before the meeting.

## Integration with Other Skills

- **slurp-juice**: Pull session history for recent work context
- **qmd**: Search vault for meeting notes, decisions, patterns
- **atlassian**: Fetch Jira issues, Confluence pages
- **pr-watcher**: Check PR status for review requests

## Tips

- Run prep 15-30 minutes before the meeting
- Save briefs to `docs/meeting-prep/<date>-<meeting>.md` for reference
- Use `weekly-brief.sh` on Monday morning for the full week overview
- Cross-reference with PR watcher for pending review requests
