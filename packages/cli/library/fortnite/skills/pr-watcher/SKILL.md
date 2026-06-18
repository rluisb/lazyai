---
name: pr-watcher
description: GitHub PR review monitor via gh-dash integration. Watches for PRs requesting review from specific team members. Uses gh-dash config sections for structured PR queues.
trigger: /pr-watcher
skill_path: skills/pr-watcher
scripts:
  - name: gh-dash-team-reviews.sh
    description: Query gh-dash config sections for team member review requests
    path: scripts/gh-dash-team-reviews.sh
  - name: pr-status-summary.sh
    description: Generate summary of all PR status across watched repos
    path: scripts/pr-status-summary.sh
  - name: setup-gh-dash-team-sections.sh
    description: Add team review sections to gh-dash config
    path: scripts/setup-gh-dash-team-sections.sh
---

# PR Watcher — gh-dash Integration

## Purpose
Monitor GitHub PRs using **gh-dash** configuration. Leverages gh-dash's section-based filtering to track review requests from specific team members. Integrates with daily standup and meeting prep workflows.

**Use when:**
- Checking for pending review requests — "any PRs waiting for team review?"
- Preparing for standup — "what PRs are blocked on review?"
- Before meetings — "any PR updates to mention?"
- Team health check — "how many PRs are stuck in review?"

## Watched Reviewers

| Reviewer | GitHub Username |
|----------|-----------------|
| Kim | KimGonzales |
| Nick | nicholashoyte-teachable |
| Alison | alisonbuki |
| David De Los Santos | daviddelossantos-teachable |
| Ronaldo | ronaldopassos-teachable |

## gh-dash Configuration

The skill extends your existing `~/.config/gh-dash/config.yml` with team review sections:

```yaml
prSections:
  # ... existing sections ...
  - title: "Team Reviews · Kim"
    filters: is:open is:pr review-requested:KimGonzales -author:KimGonzales draft:false
  - title: "Team Reviews · Nick"
    filters: is:open is:pr review-requested:nicholashoyte-teachable -author:nicholashoyte-teachable draft:false
  - title: "Team Reviews · Alison"
    filters: is:open is:pr review-requested:alisonbuki -author:alisonbuki draft:false
  - title: "Team Reviews · David"
    filters: is:open is:pr review-requested:daviddelossantos-teachable -author:daviddelossantos-teachable draft:false
  - title: "Team Reviews · Ronaldo"
    filters: is:open is:pr review-requested:ronaldopassos-teachable -author:ronaldopassos-teachable draft:false
```

## Scripts

| Script | Purpose |
|--------|---------|
| `gh-dash-team-reviews.sh` | Query gh-dash config sections for team member review requests |
| `pr-status-summary.sh` | Generate summary of all PR status across watched repos |

## Workflow

### Step 1: Check Team Review Requests
Run `gh-dash-team-reviews.sh` to find PRs where watched reviewers are requested:
- Uses `gh pr list` with the same filters as gh-dash sections
- Outputs structured list with PR number, title, repo, and reviewer
- Supports `--notify` flag for Slack-ready output

### Step 2: Generate Status Summary
Run `pr-status-summary.sh` for a full overview:
- Open PRs by author
- PRs awaiting team review
- Recently merged PRs
- Stale PRs (>7 days no activity)

### Step 3: Integrate with gh-dash
Open `gh dash` to see the full PR queue with team review sections:
- Navigate to "Team Reviews · Kim" section
- See all PRs waiting for Kim's review
- Use `v` to view in browser, `c` to checkout

## Integration with Other Skills

- **daily-standup**: Include PR review status in standup
- **meeting-prep**: Add PR updates to meeting briefs
- **slurp-juice**: Record PR status in session checkpoints

## Tips

- Run `gh dash` for interactive PR management
- Use `gh-dash-team-reviews.sh --notify` for Slack alerts
- Check before standup to report on review activity
- Configure repo paths in gh-dash config for quick checkout
