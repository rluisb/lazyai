# Task 005 — Migrate skills → `.agent.yaml` (drop `.prompt.md` transform)

**Phase:** 2 (defect G7 closeout)
**Estimated LOC:** ~80

## Goal

Replace the current "skills transformed into `.prompt.md` with `mode: agent` frontmatter" flow with a "skills emitted as `.agent.yaml`" flow. The standalone Copilot CLI reads `.github/agents/` but ignores `.github/prompts/`, so keeping the old path leaves skills invisible to the CLI.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/copilot.go` | Remove `copyLibrarySubdirAsSkillPrompts` + `copySubdirAsSkillPromptsFromFS` + `copySkillAsPromptWithRecord` + `copySkillAsPromptFromFS`. Replace with `copySkillsAsAgents(ctx, agentsDir)` that reads each library skill and emits `.github/agents/<skill>.agent.yaml` with a Copilot-flavored frontmatter. |
| `internal/adapter/shared.go` | If `EnsureModeAgentFrontmatter` is Copilot-only, remove it. If shared, leave but stop calling from Copilot path. |
| `internal/adapter/copilot_test.go` | Add `TestCopilot_SkillsEmittedAsAgents`; remove or update assertions that expected `.prompt.md` skill output. |

## Skill → agent transform

For each library skill:
1. Read skill body.
2. Construct a `.agent.yaml` with:
   - `name: <skill-id>`
   - `displayName: <Title Case>`
   - `description: <first paragraph of skill, ≤200 chars>`
   - `tools: ["*"]`
   - `prompt: |<full skill body>`
3. Write to `.github/agents/<skill-id>.agent.yaml`.

## Migration hygiene

- On install, if a previously-emitted `.github/prompts/<skill-id>.prompt.md` exists AND ai-setup owns it (file-records lookup), remove it during this install so the user doesn't end up with duplicates.
- Non-ai-setup-owned files left alone.

## Acceptance criteria

- [ ] Skills end up under `.github/agents/` not `.github/prompts/`
- [ ] `library/prompts/*` content continues to emit as `.prompt.md` files (unchanged behavior)
- [ ] Previously-emitted `skill.prompt.md` files from ai-setup are cleaned up on re-run
- [ ] Frontmatter matches the zod schema from research §2.2

## Test plan

- Seed a fake library skill, run install, assert `.github/agents/<id>.agent.yaml` content
- Seed `.github/prompts/foo.prompt.md` with matching file-record, run install, assert file removed and `.github/agents/foo.agent.yaml` present
- Ensure `library/prompts/compact.md` still emits as `.github/prompts/compact.prompt.md`

## Notes

- Chatmodes behavior is unchanged (decision Q9 — VS-Code-only).
- This task and 004 can be collapsed into one commit if the reviewer prefers; kept separate for clearer diff review.
