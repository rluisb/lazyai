 # Artifact Rules

## Canonical Source

- Canonical artifacts live under `packages/cli/library/` for emitted library content; `.agents/` is the repo-level maintainer source that `bin/inject` consumes.
- Generated CLI-specific outputs (`.claude/`, `.opencode/`, `.github/`, `.pi/`, `.gemini/`) must not be hand-edited; regenerate with `lazyai-cli compile`.
- Run `lazyai-cli compile` after canonical artifact changes to propagate to tool-native surfaces.
- Run `lazyai-cli doctor` before treating adapter output as current.

## Artifact Shapes

- Rule or policy: one markdown file unless runtime enforcement is needed.
- Agent: one canonical agent file under `packages/cli/library/canonical/agents/<name>.md`.
- Skill: `packages/cli/library/skills/<name>.md` with skill frontmatter; emitted per-tool as `<tool-dir>/skills/<name>/SKILL.md`.
- Hook: `packages/cli/library/hooks/<name>.md` with hook frontmatter; emitted per-tool as native hook config plus scripts.
- Command: `packages/cli/library/opencode/commands/<name>.md` or `claudecode/commands/<name>.md` depending on the originating tool surface.

## Compatibility

- Claude Code supports skills, agents, rules, and lifecycle hooks.
- OpenCode supports skills, agents, commands, chat modes, and plugins; hook behavior maps to TypeScript/JavaScript plugins.
- Copilot uses `.github/agents/<name>.agent.md` for canonical agents; selected LazyAI skills use Agent Skills directories instead of legacy `.github/agents/<name>.agent.yaml` or `.github/agents/<name>.agent.md` outputs.
- Pi emits `.pi/extensions/*.ts` hook extensions and receives shared markdown context and skills; OMP receives shared markdown context and skills.
- Antigravity is minimal `.gemini/settings.json` plus `.gemini/hooks/` scripts; no separate skills or agents are emitted.
- If a capability cannot be represented for one CLI, document the limitation in the artifact and make `lazyai-cli doctor` warn.
