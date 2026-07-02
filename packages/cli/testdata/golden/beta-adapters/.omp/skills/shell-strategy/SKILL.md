---
name: shell-strategy
description: Use when running shell commands in a headless/non-interactive agent session (including background subagent dispatches) so no command hangs waiting for a TTY, editor, pager, or prompt.
---

# Shell Strategy (Headless / Non-Interactive)

## When to Use

Use this skill whenever the agent executes shell commands without a TTY:
- Any bash tool call in an agent session.
- Background subagent dispatches (e.g. `pi -p` headless runs).
- Scripts intended to run unattended (CI, hooks, cron).

Do not use for guidance on *which* command proves a change worked — that is `fast-feedback`.

## Rule

Assume `CI=true`: no command may wait for human input. Anything that opens an editor, pager, or interactive prompt will hang the session. Supply explicit answers (`-y`, `-m`, `-f`, `--yes`) up front and prefer agent file tools (read/write/edit) over shell text manipulation.

## Environment Hardening

Export these when spawning long-lived or subagent shells:

| Variable | Value | Purpose |
|----------|-------|---------|
| `CI` | `true` | General CI/headless detection |
| `GIT_TERMINAL_PROMPT` | `0` | Block git credential prompts |
| `GIT_EDITOR` | `true` | Block git editor invocations |
| `GIT_PAGER` / `PAGER` | `cat` | Disable pagers |
| `EDITOR` / `VISUAL` | `true` | Block editor invocations |
| `DEBIAN_FRONTEND` | `noninteractive` | Debian/Ubuntu package installs |

## Non-Interactive Substitutions

| Instead of | Use |
|------------|-----|
| `git commit` | `git commit -m "msg"` (or `git commit -F /tmp/msg.txt` for long bodies) |
| `git rebase -i` | `git rebase --autosquash <base>` or non-interactive rebase |
| `vim`/`nano`/`less`/`more`/`man` | agent read/write/edit tools; `command --help \| cat` |
| `npm init` | `npm init -y`; `npx --yes <cmd>` |
| `apt-get install pkg` | `apt-get install -y pkg` |
| `ssh host` | `ssh -o BatchMode=yes host` |
| `rm file` (prompting alias) | `rm -f file` |

Multi-line commit messages: `git commit -m "title" -m "body"`, or write the message to a temp file and `git commit -F` it.

## Timeout Protection

Wrap network or potentially-hanging operations in `timeout`:

```bash
timeout 30 curl -s https://api.example.com
timeout 60 git fetch origin
timeout 300 <long build command>
```

## Constraints

- Never invoke editors or pagers; they are banned, not merely discouraged.
- Never pass `-i` (interactive) flags to any command.
- Prefer heredocs (`cat > file << 'EOF'`) or the write tool for file creation; never open an editor.
- Force flags (`-f`, `--force`) satisfy *prompt suppression* only — they do not waive destructive-operation approval rules from damage-control or worktree policies.

## Verification Checklist

- [ ] No command in the plan opens an editor, pager, or prompt.
- [ ] Git commands carry explicit `-m`/`-F` messages and non-interactive flags.
- [ ] Network/long operations have timeouts.
- [ ] Headless environment variables set for spawned subshells.

## Related Skills

- `fast-feedback` — choosing the smallest verification command (which must also be headless).
