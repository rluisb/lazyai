# Codex compile target — research & design dossier

Status: **Research complete; awaiting plan approval (RPI gate).**
Branch: `feat/codex-surfaces` · Worktree: `.worktrees/feat-codex-surfaces`
Goal: add `codex` as a first-class LazyAI compile target, mirroring existing adapters (kiro/pi) while respecting Codex's official native formats.

---

## 1. Authoritative Codex facts (all from https://developers.openai.com/codex)

### 1.1 Config file
- User-level: `~/.codex/config.toml` (home is `$CODEX_HOME`, default `~/.codex`). [config-basic, config-reference]
- Project-scoped overrides: `.codex/config.toml`, **loaded only for trusted projects**.
- Precedence (high→low): CLI flags/`-c` → project `.codex/config.toml` (root→cwd, closest wins) → profile `$CODEX_HOME/<name>.config.toml` → user `~/.codex/config.toml` → system `/etc/codex/config.toml` → built-in defaults.
- **Project config CANNOT override**: `model_provider`, `model_providers`, auth/base-url keys, `notify`, `profile(s)`, telemetry/`otel`. **`mcp_servers` IS allowed in project config.** [config-reference]

### 1.2 MCP servers (THE core mapping — `.ai/mcp.json` → Codex) [codex/mcp, config-reference]
TOML table `[mcp_servers.<id>]`.
- **stdio**: `command` (req), `args` (array), `env` (map), `env_vars` (array of names or `{name, source="local"|"remote"}`), `cwd`, `experimental_environment`.
- **streamable HTTP**: `url` (req), `bearer_token_env_var`, `http_headers` (map), `env_http_headers` (map name→envvar).
- **common**: `enabled`, `required`, `startup_timeout_sec` (def 10) / `startup_timeout_ms`, `tool_timeout_sec` (def 60), `enabled_tools` (allow), `disabled_tools` (deny, after allow), `default_tools_approval_mode` (auto|prompt|approve), `tools.<tool>.approval_mode`, `scopes`, `oauth_resource`.

Verbatim examples:
```toml
[mcp_servers.context7]
command = "npx"
args = ["-y", "@upstash/context7-mcp"]
env_vars = ["LOCAL_TOKEN"]

[mcp_servers.context7.env]
MY_ENV_VAR = "MY_ENV_VALUE"
```
```toml
[mcp_servers.figma]
url = "https://mcp.figma.com/mcp"
bearer_token_env_var = "FIGMA_OAUTH_TOKEN"
http_headers = { "X-Figma-Region" = "us-east-1" }
```

### 1.3 Instructions — `AGENTS.md` [guides/agents-md]
- Global: `~/.codex/AGENTS.md` (or `AGENTS.override.md`, first non-empty wins).
- Project: from git root down to cwd; per dir `AGENTS.override.md` → `AGENTS.md` → `project_doc_fallback_filenames`; one file per dir; concatenated root→down (closer overrides).
- Limit `project_doc_max_bytes` (32 KiB default). Plain markdown; no frontmatter required.
- `/init` generates an `AGENTS.md` scaffold. `model_instructions_file` can replace built-in instructions.

### 1.4 Skills — `.agents/skills/<name>/SKILL.md` [codex/skills]  ← KEY
- **Codex reads the SAME `.agents/skills` standard LazyAI already emits.** Scan order: `$CWD/.agents/skills`, parent dirs up to repo root, `$HOME/.agents/skills`, `/etc/codex/skills`, plus bundled system skills. Follows symlinks.
- `SKILL.md` requires `name` + `description` frontmatter. Optional `scripts/`, `references/`, `assets/`, `agents/openai.yaml` (UI metadata + `policy.allow_implicit_invocation`).
- Enable/disable via `[[skills.config]]` (`path`, `enabled`) in `~/.codex/config.toml`.
- **Implication:** the canonical `.agents/skills` tree is natively consumed by Codex — likely no per-target transformation needed for skills.

### 1.5 Custom prompts — **DEPRECATED** [codex/custom-prompts]
- `~/.codex/prompts/*.md` (global only, explicit invocation `/prompts:<name>`, frontmatter `description`/`argument-hint`, `$1..$9`/`$ARGUMENTS`/`$NAMED` placeholders).
- Officially deprecated; **Codex steers reusable instructions to skills.** → Map our prompts/commands to skills, not to `~/.codex/prompts`.

### 1.6 Subagents — `[agents.<name>]` [config-reference, codex/subagents]
- `agents.<name>.description`, `agents.<name>.config_file` (TOML layer path), `nickname_candidates`; globals `agents.max_threads` (6), `agents.max_depth` (1). Gated by `features.multi_agent` (on by default).

### 1.7 Hooks — `hooks.json` / inline `[hooks]` [config-reference; features.hooks default true]
- Events: `PreToolUse`, `PermissionRequest`, `PostToolUse`, `PreCompact`, `PostCompact`, `SessionStart`, `SubagentStart`, `SubagentStop`, `UserPromptSubmit`, `Stop`. Command hooks supported; prompt/agent handlers parsed but skipped. Project `.codex/hooks.json` honored only for trusted projects.

### 1.8 CLI surface [codex/cli, cli/reference]
- Open source, Rust. Install via npm (`@openai/codex` → `codex` binary) or Homebrew. Auth: `codex login` (ChatGPT OAuth / API key); creds in `$CODEX_HOME`.
- Relevant subcommands: `codex` (TUI), `codex exec` (non-interactive), `codex mcp` (add/list/remove/login), `codex doctor`, `codex features`, `codex login/logout`. Sandbox: `read-only|workspace-write|danger-full-access`. Approval: `untrusted|on-request|never`.

---

## 2. Repo wiring footprint (verified file:line in this worktree)

`ConfigDir` for Codex = **`.codex`**. MCP output = **`.codex/config.toml`** (TOML — unlike every other adapter, which emit JSON).

### 2.1 Core (must change)
1. `packages/cli/internal/types/types.go:47-67` — add `ToolIdCodex ToolId = "codex"`; append to `SupportedToolIDs`.
2. `packages/cli/internal/adapter/registry.go:40-59` — add `case types.ToolIdCodex: return &CodexAdapter{}, nil`.
3. `packages/cli/internal/adapter/codex.go` — **NEW** `CodexAdapter` (`ID/Name/ConfigDir=".codex"/Install/CompileMCP/CanRunHeadless=false/RunHeadless* no-ops`). Install ensures `.codex/`, writes root `AGENTS.md` (canonical context), relies on canonical `.agents/skills`. Mirror `kiro.go`/`pi.go` + `CopyLibraryDirectory`.
4. `packages/cli/internal/adapter/mcp_compiler.go:94-110` — add `case types.ToolIdCodex: return compileCodexMCP(...)`; implement `compileCodexMCP` emitting `[mcp_servers.*]` TOML into `.codex/config.toml` (managed-region/idempotent write); extend the supported-tools error string (line 110).
5. `packages/cli/internal/aimanifest/aimanifest.go:74-89,217-219` — add `"codex": types.ToolIdCodex` to both maps; add `"codex"` to `Default()` targets.
6. `packages/cli/internal/compiler/compiler.go:51-59` — add Codex `Description` + `Notes` (context-doc generation).
7. `packages/cli/internal/globalpaths/globalpaths.go:40-62` — add `case types.ToolIdCodex: return filepath.Join(homeDir, ".codex")`; add to the supported switch.
8. `packages/cli/internal/adapter/scope.go` — add Codex to scope-support map (verify exact shape).
9. `packages/cli/internal/adapter/capabilities.go` — add Codex capability entry (verify).
10. `packages/cli/internal/adapter/output_mapping.go` / `output_contract.go` — add Codex mapping if these enumerate targets (verify).
11. `packages/cli/internal/migration/detector.go:39-57` — add Codex detection (`.codex`) + display name.
12. `packages/cli/library/codex/...` — **NEW** library assets dir (parallels `antigravity/`, `pi/`); + `library/asset_manifest.go:28` `CurationCoverageRoots` if Codex ships curated assets.

### 2.2 Guardrails to FLIP (currently assert codex = rejected)
- `packages/cli/internal/adapter/scope_test.go:93` — `{ToolId("codex"), Global, {"", true}}` → now supported.
- `packages/cli/internal/globalpaths/globalpaths_test.go:79` — `{ToolId("codex"), false}` → true.
- `packages/cli/internal/aimanifest/aimanifest_test.go:52,218` — `"codex rejected"` cases → accepted; update `"all seven"`→ eight.

### 2.3 Tests to ADD
- `packages/cli/internal/adapter/codex_adapter_test.go` — Install layout + idempotency.
- `mcp_compiler` Codex TOML test (mirror `mcp_compiler_*_test.go`): stdio + HTTP servers → exact TOML.
- `registry_test.go`, `workflow_boundary_test.go:24-27` adapter map, `golden_test.go` fixtures (regenerate).

### 2.4 Docs/specs enumerating "7 targets" (update to 8)
- `.github/copilot-instructions.md` ("Codex is not a target" — reverse), root `README.md`, MkDocs target lists, `specs/KNOWLEDGE_MAP.md`, any spec mentioning the target set.

---

## 3. Proposed adapter design (mirror kiro/pi)

| Surface | Source (canonical) | Codex output | Notes |
|---|---|---|---|
| Instructions | canonical context | `AGENTS.md` (project root) | Codex-native; plain markdown |
| MCP catalog | `.ai/mcp.json` | `.codex/config.toml` `[mcp_servers.*]` | **TOML emitter (new)**; stdio + HTTP per §1.2 |
| Skills | `.agents/skills/<n>/SKILL.md` | `.agents/skills/...` (shared) | Native standard — no transform |
| Agents | canonical agents | `.codex/` + `[agents.<name>]`? | **Open question** — see §4 |
| Hooks | canonical hooks | `.codex/hooks.json`? | **Open question** — see §4 |
| Prompts | canonical prompts | (skills) | Codex custom-prompts deprecated |

---

## 4. Open questions (resolve in plan before implementing)
1. **Scope of v1**: minimal (AGENTS.md + MCP `config.toml` + rely on shared `.agents/skills`) vs full (also agents `[agents.*]` + `hooks.json`)? Recommend **minimal first** — matches the cleanest native surfaces and how pi/kiro grew incrementally.
2. **MCP into `config.toml`**: emit a dedicated managed block within `.codex/config.toml` (must preserve user keys + be idempotent). Confirm the managed-region writer supports TOML (others write JSON files wholesale). Likely need a TOML-aware managed write or a separate `.codex/config.toml` owned-region strategy.
3. **Agents mapping**: Codex `[agents.<name>]` needs per-role `config_file` TOML layers — heavier than kiro's `.kiro/agents/<name>.md`. Defer to v2?
4. **Global scope**: `~/.codex/config.toml` + `~/.codex/AGENTS.md` for global installs (globalpaths already patterned).
5. **Guardrail reversal sign-off**: confirm the deliberate "codex rejected" tests/manifest checks should be flipped (user already directed this).

---

## 5. Sources
- https://developers.openai.com/codex/config-reference
- https://developers.openai.com/codex/config-basic
- https://developers.openai.com/codex/mcp
- https://developers.openai.com/codex/skills
- https://developers.openai.com/codex/guides/agents-md
- https://developers.openai.com/codex/custom-prompts (deprecated)
- https://developers.openai.com/codex/cli + /cli/reference + /cli/slash-commands

---

## 6. Implementation status (full parity — built)

Status: **Implemented + verified.** `codex` is a first-class target. Build, `go vet`, `gofmt`, full `go test ./...`, `mkdocs build --strict`, and an end-to-end `init --tools codex` + `compile --tool codex` smoke all pass.

### Delivered surfaces
- **Instructions** → root `AGENTS.md` (scaffold `RootFileByTool` + compiler context-doc entry).
- **MCP** → `.codex/config.toml` `[mcp_servers.*]` via `compileCodexMCP` + `configmerge.MergeTOMLFile` (user-preserving, idempotent; stdio + HTTP).
- **Subagents** → `.codex/agents/<name>.toml` (name/description/developer_instructions) via `RewriteAgentForCodex` (markdown→TOML); default `guide` excluded.
- **Skills** → `.agents/skills/<name>/SKILL.md` (Codex-native standard).
- **Hooks** → `.codex/hooks.json` (Codex event schema) + `.codex/hooks/lazyai/block-destructive-shell.sh`.

### Key files
- New: `internal/adapter/codex.go`, `internal/adapter/codex_test.go`, `library/codex/hooks.json`, `library/codex/hooks/lazyai/block-destructive-shell.sh`.
- `RewriteAgentForCodex` + `codexAgentTOML` in `agent_transform.go`; `compileCodexMCP` + `toCodexMcp` in `mcp_compiler.go`.
- Enumeration wired: `types`, `registry`, `aimanifest`, `globalpaths`, `scope`, `compiler`, `migration`, `output_mapping`, `capabilities`, `setupscan`, `models/catalog`, `scaffold/filemap`, `library/embed.go`, wizard (`phase1`, `hover_descriptions`).
- Guardrails flipped: removed the hardcoded codex rejections (`aimanifest` resolver/`ErrCodexUnsupported`) and updated the codex-rejected unit tests to codex-accepted.

### Decisions made during build
- **Support level = Beta.** Surfaces are docs-verified + golden/unit-tested, but no runtime smoke against the real Codex binary. `TestNoBetaAdaptersRemain` narrowed to permit only codex below stable. Promote to Stable after binary smoke.
- **Hooks scope: PreToolUse only.** The PreToolUse `block-destructive-shell` hook maps cleanly (stable `tool_name`/`tool_input.command` fields + `permissionDecision:"deny"`). The Stop `objective-workflow-gate` was **intentionally omitted**: Codex's Stop payload exposes no stable last-assistant-message field (only an explicitly-unstable `transcript_path`), so a Stop gate would be a fragile near-stub. Revisit if Codex stabilizes a Stop transcript/message field.
- **`KNOWLEDGE_MAP` spec-029 line left intact** (it accurately records that 029 dropped Codex); this re-addition supersedes it going forward.
