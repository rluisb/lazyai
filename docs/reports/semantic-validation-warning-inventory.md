# Semantic Validation Warning Inventory

Date: 2026-06-22  
Scope: Shipped library assets under `packages/cli/library/`  
Method: Map library assets into a temp `.ai/` tree, run `validate.All()` (personal profile), capture every issue.  
Command: `go test -v -count=1 -run TestLibraryAssetWarningInventory ./packages/cli/internal/validate/...` (ad-hoc test, removed after capture)  
Profile: `personal` (default — inline secrets are warnings, not errors)

## Summary

| Rule | Total | Errors | Warnings | True Positives | False Positives |
|------|-------|--------|----------|----------------|-----------------|
| hook | 3 | 3 | 0 | 0 | 3 |
| **Total** | **3** | **3** | **0** | **0** | **3** |

All 3 findings are **false positives** — the validate engine's `dangerousHookPatterns` matcher scans every file under `.ai/hooks/` for literal shell-dangerous strings, but `block-destructive-shell.md` is a **documentation policy** that lists denied commands in a markdown bullet list. The engine has no way to distinguish documentation from executable shell scripts.

No other rule category (`skill`, `agent`, `manifest`, `mcp`, `secret`, `path`) produced any issues. All 52 shipped library assets (8 canonical agents, 4 canonical skills, 2 canonical hooks, 6 library hooks, 32 library skills, MCP catalog) pass validation cleanly.

---

## Detailed Findings

### Rule: `hook` — 3 errors

All three errors come from a single file: `.ai/hooks/block-destructive-shell.md`.

| # | File | Severity | Message | Assessment |
|---|------|----------|---------|------------|
| 1 | `.ai/hooks/block-destructive-shell.md` | error | contains dangerous command "rm -rf /" | **False positive** — appears in a markdown bullet list of denied commands |
| 2 | `.ai/hooks/block-destructive-shell.md` | error | contains dangerous command "mkfs" | **False positive** — same documentation list |
| 3 | `.ai/hooks/block-destructive-shell.md` | error | contains dangerous command "dd if=" | **False positive** — same documentation list |

**Root cause:** The `validateHooks` function in `packages/cli/internal/validate/validate.go` normalizes whitespace and scans every file under `.ai/hooks/` for `dangerousHookPatterns` (lines 242–254). It does not distinguish between:
- Executable shell scripts (`.sh`, `.bash`, no extension) — where these patterns are genuinely dangerous
- Documentation files (`.md`, `.yml`) — where these patterns appear in prose examples

The file `block-destructive-shell.md` is a **policy document** that explicitly lists denied commands. The engine flags the documentation itself as dangerous.

**Recommended cleanup:** Add a file-extension filter to `validateHooks` so that dangerous-pattern scanning only applies to shell-script extensions (`.sh`, `.bash`, no-extension executables). Markdown and YAML files under `.ai/hooks/` should be exempt from dangerous-pattern matching, or the scan should only apply to files that have a shebang line.

---

## Assets That Passed Cleanly

### Canonical Agents (8 files)
All have valid frontmatter with `name`, `description`, `role`, `mode`, `temperature`, `steps`:
- `guide.md`, `implementer.md`, `researcher.md`, `planner.md`, `reviewer.md`, `deployer.md`, `responder.md`, `evidence-verifier.md`

### Canonical Skills (4 files)
All have valid frontmatter with `name`, `description`, `trigger`, `tier`, `thinking`, `risk`:
- `codebase-exploration.md`, `diagnose.md`, `pr-review.md`, `test-first-change.md`

### Canonical Hooks (2 files)
Both are markdown documentation (no shebang, no dangerous patterns):
- `pre-commit.md`, `session-start.md`

### Library Hooks (6 files)
- `block-destructive-shell.md` — **3 false positives** (see above)
- `caveman-memory-promotion.md`, `objective-workflow-gate.md`, `startup-self-heal.md` — clean
- `pre-commit` — shell script with shebang, no dangerous patterns
- `rpi-gate-check.yml` — YAML CI config, no dangerous patterns

### Library Skills (32 files)
All have valid frontmatter with `name` and `description`:
- `adhd-engineer.md`, `anti-speculation.md`, `architecture-review.md`, `bugfix.md`, `caveman.md`, `chain-verify.md`, `codebase-exploration.md`, `create-agent.md`, `create-hook.md`, `create-skill.md`, `create-workflow.md`, `diagnose.md`, `doc-backed-clarify.md`, `extract-standards.md`, `fast-feedback.md`, `four-point-vibe-coding.md`, `handoff.md`, `housekeeping.md`, `impact-check.md`, `implement.md`, `improve-codebase-architecture.md`, `issue-triage.md`, `iterate.md`, `memory-promotion.md`, `memory-write.md`, `no-workarounds.md`, `parallel-execution.md`, `plan.md`, `process-audit.md`, `project-guardrails-init.md`, `proof-of-concept.md`, `red-team-plan.md`, `research.md`, `review.md`, `rpi.md`, `self-improve.md`, `skill-authoring.md`, `slack-message-formatter.md`, `slackfmt.md`, `speckit-analyze.md`, `speckit-checklist.md`, `speckit-clarify.md`, `speckit-constitution.md`, `speckit-implement.md`, `speckit-plan.md`, `speckit-specify.md`, `speckit-tasks.md`, `spike.md`, `task-to-issues.md`, `tdd-loop.md`, `tdd-planning.md`, `test-first-change.md`, `update-memory.md`, `zoom-out.md`

### Library Skills (directory-style: `populate/SKILL.md`)
- `populate/SKILL.md` — valid frontmatter

### MCP Catalog
- `.ai/mcp.json` — valid JSON, all 5 servers have `command` or `url`

### Secret Scan
No inline secrets detected in any library asset.

### Path Scan
No symlinks in the library assets.

### Manifest
No `.ai/lazyai.json` present (expected — library assets are not a project init tree).

---

## Future Cleanup Path

1. **Fix the false positive** (low effort, high signal): Add an extension filter to `validateHooks` so dangerous-pattern scanning only applies to shell-script extensions. This is a one-line change in `packages/cli/internal/validate/validate.go` around line 271:

   ```go
   isShell := ext == ".sh" || ext == ".bash" || ext == ""
   ```

   The dangerous-pattern scan (lines 279–283) should be gated on `isShell`, just like the shebang check on line 285. Currently the dangerous-pattern scan runs on **all** files regardless of extension.

2. **No bulk asset rewrites needed.** The library assets are well-formed. The only issue is a validator false positive.

3. **If the false positive is fixed**, the library asset tree will produce **zero validation issues** under the personal profile.
