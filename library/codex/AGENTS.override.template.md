<!--
This file is loaded hierarchically by Codex (openai/codex CLI). See
https://developers.openai.com/codex/guides/agents-md for how AGENTS.md and
AGENTS.override.md interact. Override wins where present.

Delete this comment block before committing; fill in the [YOUR_*] markers
below with your team's conventions. Keep this file focused on rules Codex
should never relax — prefer AGENTS.md for narrative context.
-->

# AGENTS.override.md — Codex Conventions

**Project:** [YOUR_PROJECT_NAME]
**Organization:** [YOUR_ORG]
**Team:** [YOUR_TEAM]

---

## Execution Style

- Prefer `/plan` mode for any change that touches more than a single file.
- When a change is reversible and small, proceed; when it is destructive or
  shared (migrations, env config, CI), surface a plan first.
- Default to `--skip-git-repo-check` only for one-off throwaway dirs; in the
  primary repo, stay inside git so diffs are reviewable.

## Trust and Permissions

- Read-only tools are always available.
- Write / shell tools require explicit approval unless the operation matches
  a documented workflow below.
- Never bypass approval for destructive shell operations (`rm -rf`, force
  push, migration rollbacks) — always ask, even in fast mode.

## Documented Workflows

<!-- Describe each repeatable task as a named workflow. Keep names stable
     so they can be invoked by reference: "run the nightly-sync workflow". -->

### [YOUR_WORKFLOW_1]

- Purpose: [YOUR_PURPOSE]
- Tools: [YOUR_TOOLS]
- Approval: [auto-approved | requires review]
- Expected duration: [rough estimate]

### [YOUR_WORKFLOW_2]

- Purpose: [YOUR_PURPOSE]
- Tools: [YOUR_TOOLS]
- Approval: [auto-approved | requires review]
- Expected duration: [rough estimate]

## Boundaries

- Files that MUST NOT be edited without human approval:
  - [YOUR_PATH_1] — [reason]
  - [YOUR_PATH_2] — [reason]
- Directories that are read-only for Codex:
  - [YOUR_READ_ONLY_DIR] — [reason]

## Testing

- Run tests with [YOUR_TEST_COMMAND] before declaring work complete.
- Integration tests that require external services: [YOUR_POLICY].
- If a test is newly failing, investigate before proposing a fix — do not
  mark flaky without evidence.

## Handoff and Memory

- Use `/plan` output as the primary handoff artifact. Save key decisions
  into `specs/memory/handoffs/YYYY-MM-DD-topic.md` on multi-session work.
- When context is getting long, call `/compact` proactively rather than
  letting the session auto-compress.

## Rules

- [YOUR_RULE_1]
- [YOUR_RULE_2]

## Do NOT

- Do not modify generated files without updating their source template.
- [YOUR_DO_NOT_1]
- [YOUR_DO_NOT_2]
