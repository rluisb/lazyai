---
name: slackfmt
description: Use when turning markdown-heavy or code-heavy text into Slack-optimized formatting. Keep text short, preserve intent, and use npx @slackfmt/cli without making the project depend on npm.
---

# Slackfmt

## When to Use

Use this skill when you need to:
- Format a message for Slack mrkdwn.
- Preserve lists, code, links, and emphasis in a compact reply.
- Convert a long draft into something readable in Slack.

## Rule

Write for Slack readability first: short lines, minimal nesting, and no formatting that Slack will mangle.

## Workflow

1. Trim the message to the essential points.
2. Rewrite with Slack mrkdwn conventions.
3. Keep code blocks short and legible.
4. If a CLI pass is needed, run `npx @slackfmt/cli` from an existing Node/npm environment rather than adding an npm dependency to LazyAI.
5. Verify the output still preserves meaning and tone.

## Constraints

- Do not over-format or add decorative markup.
- Do not assume npm is installed as part of LazyAI.
- Do not change the message’s meaning.
