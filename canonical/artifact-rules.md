# Artifact Rules

## Canonical Source

- Canonical artifacts live under `.agents/` or `canonical/`; CLI-specific outputs are generated.
- Do not duplicate canonical content under `.claude/`, `.opencode/`, or `.pi/`.
- Run `bin/inject` after canonical artifact changes.
- Run `bin/doctor` and `tests/test-provenance-drift.sh` before treating adapters as current.

## Artifact Shapes

- Rule or policy: one markdown file unless runtime enforcement is needed.
- Agent: one `.agents/agents/<name>.md` file.
- Skill: `.agents/skills/<name>/SKILL.md`; optional `scripts/`, `references/`, and `assets/` are allowed when justified.
- Hook: `.agents/hooks/<name>/POLICY.md`; optional scripts support Claude Code, and generated OpenCode plugins provide parity where possible.
- Workflow: `.agents/workflows/<name>.md` or `.agents/workflows/<name>.yml` once workflow support is enabled.

## Compatibility

- Claude Code supports skills, agents, rules, and lifecycle hooks.
- OpenCode supports skills, agents, and plugins; hook behavior maps to TypeScript/JavaScript plugins.
- OMP/Pi receives shared markdown context and skills; project-local hook support is not assumed.
- If a capability cannot be represented for one CLI, document the limitation in the artifact and make `bin/doctor` warn.
