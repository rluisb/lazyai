---
title: Pi (pi.dev) — Tool System
summary: Built-in tools, TypeScript custom tools, skills, and why Pi has NO native MCP system.
status: verified
verified_on: 2026-06-29
provenance: official-docs (pi.dev) via research subagent
---

# Pi (pi.dev) — Tool System

## Built-in tools

Seven, enabled by default:

| Tool | Purpose |
|---|---|
| `read` | Read files/directories |
| `bash` | Execute shell commands |
| `edit` | Targeted file edits |
| `write` | Create/overwrite files |
| `grep` | Regex search |
| `find` | Find files by name/pattern |
| `ls` | List directory contents |

### Enable/disable — CLI flags only

No per-tool config file. Controlled at invocation:

```
--tools <list>, -t <list>            # allowlist built-in + extension tools
--exclude-tools <list>, -xt <list>   # disable specific tools
--no-builtin-tools, -nbt             # disable all built-ins, keep extension tools
--no-tools, -nt                      # disable all tools
```

Extensions can also block built-in calls at runtime via the `tool_call` event returning `{ block: true, reason: "..." }`. Skills may use the experimental `allowed-tools` frontmatter field. **No declarative permissions file. No sandbox** — tools run with the Pi process user's full permissions.

## Custom tools

Registered via **TypeScript extensions** (code, not schema). `.ts` files loaded by [jiti](https://github.com/unjs/jiti) — no build step.

```typescript
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { Type } from "typebox";

export default function (pi: ExtensionAPI) {
  pi.registerTool({
    name: "greet",
    label: "Greet",
    description: "Greet someone by name",
    parameters: Type.Object({ name: Type.String({ description: "Name to greet" }) }),
    async execute(toolCallId, params, signal, onUpdate, ctx) {
      return { content: [{ type: "text", text: `Hello, ${params.name}!` }], details: {} };
    },
  });
}
```

`registerTool()` fields: `name`, `label`, `description`, `parameters` (typebox `TSchema`), `execute(toolCallId, params, signal, onUpdate, ctx)`. Return `{ content: [...], details: {} }`.

### Discovery paths

| Path | Scope | Trigger |
|---|---|---|
| `~/.pi/agent/extensions/*.ts` (+ `*/index.ts`) | Global | Always |
| `.pi/extensions/*.ts` (+ `*/index.ts`) | Project | After project trust |

Also: `settings.json` `extensions[]`; CLI `pi -e ./path.ts`; hot-reload `/reload`. Package installs: `pi install npm:@foo/bar@1.0.0 | git:... | /abs/path` → `~/.pi/agent/npm/` or `git/` (global), `.pi/npm/` or `git/` (project).

### Skills (NOT function-call tools)

Markdown instruction packages the agent reads and acts on (via `bash`), implementing the [Agent Skills standard](https://agentskills.io/specification). Directory with `SKILL.md`.

Frontmatter: `name` (req, ≤64, kebab), `description` (req, ≤1024 — drives auto-selection), `license`, `compatibility`, `metadata`, `allowed-tools` (experimental), `disable-model-invocation`.

| Path | Scope | Rules |
|---|---|---|
| `~/.pi/agent/skills/` | Global | root `.md` + `SKILL.md` subdirs |
| `~/.agents/skills/` | Global | `SKILL.md` subdirs only (root `.md` ignored) |
| `.pi/skills/` | Project (trusted) | root `.md` + `SKILL.md` subdirs |
| `.agents/skills/` | Project (trusted) | `SKILL.md` subdirs, walks to git root |

Invoke `/skill:name [args]`. Toggle with `settings.json` `enableSkillCommands` (default true); `--skill <path>` adds; `--no-skills` disables.

## MCP tools

!!! danger "Pi has NO native MCP system"
    There is **no `.mcp.json`/`mcp.json`** and no built-in MCP client documented anywhere in Pi's official docs. MCP-like functionality requires a hand-written TypeScript extension that spawns a stdio subprocess / connects to SSE / makes HTTP calls and exposes results via `pi.registerTool()`.

    **For LazyAI:** do NOT emit any MCP config file for the Pi target. [INFERENCE] Pi's design treats the TypeScript extension API as its MCP-equivalent.

## Config files & locations

| Scope | Path | Format |
|---|---|---|
| Global settings | `~/.pi/agent/settings.json` | JSON (note the `agent/` subdir) |
| Project settings | `.pi/settings.json` | JSON (trust-gated) |
| Trust decisions | `~/.pi/agent/trust.json` | JSON |
| Global skills | `~/.pi/agent/skills/`, `~/.agents/skills/` | `SKILL.md` dirs |
| Global extensions | `~/.pi/agent/extensions/` | `.ts` |
| Global prompts | `~/.pi/agent/prompts/` | `.md` |
| Global context | `~/.pi/agent/AGENTS.md` | Markdown |
| System prompt | `~/.pi/agent/SYSTEM.md` (replace) / `APPEND_SYSTEM.md` (append) | Markdown |
| Project context | `AGENTS.md` or `CLAUDE.md` (loaded regardless of trust) | Markdown |
| Sessions | `~/.pi/agent/sessions/` | JSONL |

`settings.json` tool-relevant keys: `extensions[]`, `packages[]`, `skills[]`, `prompts[]`, `enableSkillCommands`, `npmCommand`, `defaultProjectTrust`.

## Gotchas

1. **No MCP config file** — emitting one for Pi is wrong.
2. Extensions (TS, real tools) vs skills (Markdown, read-and-act) are fundamentally different.
3. Project `.pi/` resources are **trust-gated**; non-interactive runs (`-p`, `--mode json/rpc`) skip them unless `defaultProjectTrust: "always"` or `--approve`.
4. Global settings live at `~/.pi/agent/settings.json` — the `agent/` subdir is required.
5. Skill `name` may differ from dir name in Pi (lenient), but the canonical standard requires they match — shared skills may collide.
6. Skill name collisions: first found wins (global before project).
7. Skills missing `description` are silently dropped.
8. Extension factories must defer background resources to `session_start` (paired with `session_shutdown`).
9. Published packages install with `--omit=dev`; runtime deps must be in `dependencies`.
10. `--no-builtin-tools` disables `bash`/`read`/`write`/etc.; skills relying on `bash` then fail. Extension tools are unaffected.
11. `~/.agents/skills/` and `.agents/skills/` ignore root-level `.md` files (subdirs with `SKILL.md` only).

## Sources

- https://pi.dev/docs/latest
- https://pi.dev/docs/latest/extensions
- https://pi.dev/docs/latest/skills
- https://pi.dev/docs/latest/settings
- https://pi.dev/docs/latest/usage
- https://pi.dev/docs/latest/security
- https://pi.dev/docs/latest/prompt-templates
- https://pi.dev/docs/latest/packages
