# Cookbook Recipe

> Template for recurring technical problems. One recipe = one solved pattern.

## The Four Points

- **WHAT** — A specific problem (e.g., "add OAuth to a Go web app", "debug a memory leak", "set up Postgres in CI").
- **HOW** — Prerequisites, steps, verification, and common mistakes.
- **What I DON'T want** — Vague advice; links without commands; steps that assume unstated context.
- **How we VALIDATE** — Copy-paste the recipe into a fresh environment; run the verification step; expect green.

## Recipe Shape

```markdown
# Recipe: <Short Imperative Title>

> One-line description of the solved problem.

## Prerequisites

- Tool versions (e.g., Go 1.22+, Docker, psql).
- Files that must exist beforehand.
- State that must be true (e.g., "app already serves HTTP on :8080").

## Steps

1. **<Action verb>** — Exact command or code block. No hand-waving.
2. **<Action verb>** — Next exact step.
3. **<Action verb>** — ...

## Verification

Run this command and expect this output:

```bash
$ <command>
<expected output snippet>
```

## Common Mistakes

| Symptom | Cause | Fix |
|---------|-------|-----|
| <error or behavior> | <root cause> | <exact command or change> |

## References

- Link to canonical doc that governs this area.
- Link to related recipes.
```

## Rules

1. **Respect the project's existing architecture first.** A recipe adapts to the repo; the repo does not contort around the recipe.
2. **Prerequisites before steps.** Never start a recipe with "First, install X" without stating what X is and why.
3. **Commands are copy-pasteable.** Use `$` prefix for shell, no `$` for code blocks meant to be pasted into files.
4. **Verification is a command, not a feeling.** "It should work" is not a verification step.
5. **One recipe per concern.** "Set up a full stack" is a workflow, not a recipe. Break into: Postgres recipe, migrations recipe, OAuth recipe.
6. **Version-pin everything.** Use `@version` in npx, `go get package@version`, `docker run image:tag`.

## Categories

| Tag | Examples |
|-----|----------|
| `backend` | DB migrations, API auth, rate limiting |
| `frontend` | Component pattern, build optimization, a11y fix |
| `devops` | CI setup, container build, secret rotation |
| `debug` | Profile, trace, reproduce, bisect |
| `agent` | MCP install, skill add, context strip tuning |

## When to Create a Recipe

Create one when:

- You solve the same problem a second time.
- A teammate asks you for the same steps twice.
- The solution spans more than 3 commands and you always forget one.

Do not create one when:

- The tool's own README is sufficient and already version-pinned.
- The problem is a one-off exploration with no repeatable path.

## Storage

Store in `docs/recipes/` or `canonical/recipes/` depending on scope:

- **Project-local** (`docs/recipes/`): Framework-specific, stack-specific.
- **Canonical** (`canonical/recipes/`): Tool-agnostic patterns (debugging, auth, CI) that cross projects.
