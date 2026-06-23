---
name: slack-message-formatter
description: Use when you need a Slack-ready rewrite of a longer message. Produce sharp, legible, correctly spelled Slack mrkdwn and keep it concise enough to skim.
---

# Slack Message Formatter

## When to Use

Use this skill when:
- A message needs to be pasted into Slack.
- You want a polished, concise version of a longer draft.
- Formatting must survive Slack’s markdown rules.

## Rule

Make the message easy to scan in Slack: clear headings, short bullets, and only the formatting that helps comprehension.

## Workflow

1. Identify the core ask, update, or decision.
2. Rewrite into Slack-friendly mrkdwn.
3. Keep bullet depth shallow.
4. Use code blocks only when the exact text matters.
5. If you want tool-assisted formatting, invoke `npx @slackfmt/cli` from an existing Node/npm environment; LazyAI itself should not depend on npm.

## Constraints

- Do not add filler or corporate polish.
- Do not introduce unsupported markdown tricks.
- Do not alter technical meaning.
