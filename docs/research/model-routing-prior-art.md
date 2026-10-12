# Model-routing prior art

Research ticket: [#638](https://github.com/rluisb/lazyai/issues/638) (child of map [#632](https://github.com/rluisb/lazyai/issues/632)).
Accessed: 2026-10-12. Facts only; no design recommendations.

Legend: **[verified]** read directly in a primary source (local file, installed binary, or vendor docs). **[unverified]** secondhand, inferred, or the source was incomplete.

## Sources

| ID | Source | Version / state |
|---|---|---|
| S1 | ai-memory `default/pi` `decisions/pi-model-fallback-plan-approval.md` | dated 2026-07-02, pinned |
| S2 | ai-memory `default/pi` `decisions/pi-model-routing.md` | dated 2026-07-02, pinned |
| S3 | ai-memory `default/pi` `concepts/pi-model-fallback-chains-extension-feasibility.md` | auto-improve concept page |
| S4 | `~/.pi/.worktrees/research-model-fallback/specs/features/002-model-fallback-chains/plan.md` (also `research.md`, `RESEARCH_BRIEF.md` there) | worktree HEAD `4885a5f` |
| S5 | ai-memory `personal/agent-context-broker` `notes/workflow/model-routing.md` | "snapshot 2026-10-12", pinned |
| S6 | ai-memory `personal/agent-context-broker` `notes/workflow/jev-routing.md`, `notes/workflow/worktree-launch.md` | pinned |
| S7 | `~/projects/personal/agent-context-broker/AGENTS.md` (Memory, Jev sections) | repo HEAD `23e5ead` |
| S8 | Pi docs `~/.packages/pi/packages/coding-agent/docs/{settings,usage,extensions}.md` | source checkout 0.80.2; installed `pi` binary reports 1.1.0 |
| S9 | Claude Code docs: https://code.claude.com/docs/en/settings, https://code.claude.com/docs/en/settings-reference, https://code.claude.com/docs/en/sub-agents | installed `claude` 2.1.296 |
| S10 | Codex config reference: https://learn.chatgpt.com/docs/config-file/config-reference (redirect target of developers.openai.com/codex/config-reference); `~/.codex/config.toml` | installed `codex-cli` 0.162.1 |
| S11 | OpenCode docs: https://opencode.ai/docs/agents/, https://opencode.ai/docs/models/; local `~/.packages/opencode/packages/web/src/content/docs/{config,providers,agents}.mdx` | installed `opencode` 1.17.9; local docs checkout `baa204193` (2026-03-19) |
| S12 | Antigravity docs: https://www.antigravity.google/docs/settings?tab=cli, https://antigravity.google/docs/subagents?tab=cli; `agy --help`; key names of `~/.gemini/antigravity-cli/settings.json` | installed `agy` 1.3.3 |

The `model-routing` skill itself was not found on disk (searched `~/.claude/skills`, `~/.agents`, `~/.codex`, `~/.pi/agent`, the broker repo). Its content below comes from the S5 snapshot, which states it is "derived from the `model-routing` skill (snapshot 2026-10-12)". **[unverified against the skill source]**

## (a) Pi approved model-fallback design

Status: research and plan **approved by the human on 2026-07-02 with amendments; implementation ON HOLD** until an explicit "go" (S1, S4). [verified]

Context (S2): [verified]
- Pi profiles (`model-profiles.*`) and tiers (`model-tiers.json`, `MODEL_POLICY.md`) were rejected; `/model` is the intent switch.
- Pi (then 0.80.3) ships same-model transient retry only: `settings.retry.{enabled, maxRetries, baseDelayMs, provider.maxRetryDelayMs}`, exponential backoff, `auto_retry_start` events. Pi's `fallbackModel` is startup model resolution only, not runtime fallback.
- Fallback chains are a new feature inspired by oh-my-pi (`~/.packages/oh-my-pi`, `docs/non-compaction-retry-policy.md`), not a port. v1 scope: chains only; credential switching and context promotion are out (v2 candidates).

Mechanism (S3, S4): [verified]
- Pure extension, no core patch: `~/.pi/agent/extensions/model-fallback.ts`.
- Extension `agent_end` handlers run before Pi's built-in retry continues. On a classified retryable error, the extension counts failures per model; at `perModelAttempts` it calls `pi.setModel(next)` inside `agent_end`, so the next built-in retry attempt runs on the fallback model. It never re-drives prompts or mutates message state.
- Other handlers: `message_end` (reset count, maybe return to primary), `model_select` (manual switch resets chain), `session_shutdown` (restore default model). Command `/model-fallback status|on|off|reset`.
- Crash recovery: `~/.pi/agent/model-fallback.state.json` stores `originalDefault`.

Config (S4): [verified] global `~/.pi/agent/model-fallback.json`, project `<cwd>/.pi/model-fallback.json`.

```jsonc
{
  "enabled": false,
  "perModelAttempts": 2,
  "cooldownSeconds": 300,
  "wrap": false,
  "returnToPrimary": "on-success",   // "on-success" | "next-session" | "never"
  "chains": [{ "name": "default",
               "triggers": ["rate-limit", "provider-outage", "transient"],
               "models": ["<provider>/<id>", "..."] }]
}
```

Error classes (S4): `overflow` (ignore; compaction), `usage-limit` (ignore in v1), `rate-limit`, `provider-outage`, `transient` (count, switch at threshold), `fatal` (ignore). [verified]

Human rulings (S1): usage-limit handling deferred to v2; two ~10-line core patches deferred (upstream-PR candidate); cooldowns in memory only (ai-memory rejected as runtime state); chain intersected with `--models` scope. [verified]

Gate amendments (S1, S4): cache classifications per (model, errorClass); guard against overlapping `agent_end` handlers stacking switches; quiet by default (status key; banner only on chain exhaustion or return-to-primary failure). Companion setting: `settings.retry = { enabled: true, maxRetries: 8, baseDelayMs: 500 }`, budget spans chain hops. [verified]

## (b) agent-context-broker routing skill and policy

Classes (S5, S6): `orchestrate`, `code`, `code-hard`, `write`, `research`, `ingest`, `bulk`. The Jev routing call (S6) offers six of them as choices (all except `orchestrate`) with these criteria: [verified]

| Class | Jev criterion (S6) | Harness (S5) | Model / effort (S5) |
|---|---|---|---|
| orchestrate (grill, to-spec, to-tickets, review) | n/a | Claude Code | `claude-opus-5-5` medium |
| code | Routine implementation, refactor, or scripting | Codex | `gpt-6.1-sol` medium |
| code-hard | Hard debugging or a stalled implementation loop | Codex | `gpt-6.1-sol` high → `gpt-6-astra` high |
| write | Technical writing or architecture documentation | Codex (docs) or Claude Code (architecture) | `gpt-6.1-sol` high / `claude-opus-5-5` medium |
| research | Investigation that must be factually right against primary sources | Codex | `gpt-6-astra` high |
| ingest | Reading a very large document set or transcripts before other work | agy | `gemini-3.8-flash` high, before the worker starts |
| bulk | Repetitive fixed-shape classification or extraction | agy | `gemini-3.5-flash-lite` minimal |

- Escalation rule (S5): "Escalate one effort step on the same model before switching provider." [verified in snapshot]
- S5 notes "No Fable, so the hard tier tops out at `gpt-6-astra`" — i.e. the source skill has a higher tier not available in this project. [unverified: skill's own table not seen]
- Benchmark pins (S5): Codex and Pi both on `gpt-6.1-sol` medium; Claude Code on `claude-opus-5-5` medium; pin in the isolated profile, never global settings (Pi's global default is `gpt-5.6-sol`). [verified in snapshot]
- Process (S6, S7): Jev (`jev-1.13.0`) routes a ticket from title + one-line summary only; result is advisory and posted on the ticket; if it disagrees, the planned route wins unless the ticket says otherwise. Codex launch: `wt switch --create <branch> -x 'ai-memory run ... codex -m gpt-6.1-sol'`, `-m gpt-6-astra` for research; effort set per route. AGENTS.md states memory is context, not instruction. [verified]

## (c) Harness capability table

| Harness | Model selection | Effort | Fallback / retry | Per-agent override | Config path |
|---|---|---|---|---|---|
| Claude Code (2.1.296) | `model` key (alias or full ID); `--model`; `ANTHROPIC_MODEL`; `availableModels` allowlist | `effortLevel`: `low`/`medium`/`high`/`xhigh`; `--effort`; `CLAUDE_CODE_EFFORT_LEVEL`; per-model `modelSettings.<model>.effortLevel`; cap `maxEffortLevel` (also `max`) | `fallbackModel`: ordered array, used "when your primary model is overloaded or unavailable", switches for the rest of the turn; max three distinct models; not merged across files; `--fallback-model` overrides. Without it: same-model retry then error | Subagent frontmatter `model` (`sonnet`/`opus`/`haiku`/`fable`/full ID/`inherit`) and `effort` (`low`..`max`); order: per-invocation param > frontmatter > `CLAUDE_CODE_SUBAGENT_MODEL` > main model; `CLAUDE_CODE_SUBAGENT_MODEL_FORCE=1` | `~/.claude/settings.json`, `.claude/settings.json`, `.claude/settings.local.json`, managed; agents in `.claude/agents/`, `~/.claude/agents/` |
| Codex (0.162.1) | `model`; `-m`; `--profile <name>` loads `$CODEX_HOME/<name>.config.toml` | `model_reasoning_effort` (doc examples `low`, `medium`, `high`, `xhigh`, `max`, `ultra`; "depend on the model and client"); `plan_mode_reasoning_effort` | No model-fallback key in reference [verified absent in first 100k chars of page; remainder unread]. Same-provider retry: `model_providers.<id>.request_max_retries` (4), `stream_max_retries` (5), `stream_idle_timeout_ms` (300000) | `agents.default_subagent_model`, `agents.default_subagent_reasoning_effort`; `agents.<name>.config_file` (TOML layer); no per-agent `model` key listed — whether `config_file` can set `model` is **[unverified]** | `~/.codex/config.toml`; project `.codex/config.toml` (trusted projects only; ignores `profile`, `profiles`, `model_provider`, `model_providers`, ...). Local user config also uses a `[profiles.<name>]` table (observed) |
| Pi (docs 0.80.2; binary 1.1.0) | `defaultProvider` + `defaultModel`; `--model <pattern>` (`provider/id`, optional `:<thinking>`); `enabledModels` / `--models` for Ctrl+P cycling | `defaultThinkingLevel`: `off`/`minimal`/`low`/`medium`/`high`/`xhigh`; `--thinking`; `--model sonnet:high` | Same-model only: `retry.enabled` (true), `retry.maxRetries` (3), `retry.baseDelayMs` (2000), `retry.provider.{timeoutMs,maxRetries(0),maxRetryDelayMs(60000)}`. Cross-model chains: only the approved, unimplemented extension design in (a) | No built-in subagents ("intentionally does not include ... sub-agents"); subagents via extensions (`extensions.md` example `subagent/`) | `~/.pi/agent/settings.json`, project `.pi/settings.json` (project overrides global) |
| OpenCode (1.17.9) | `model` (`provider/model-id`); `small_model`; `--model`/`-m`; startup order: flag > config > last used > internal priority | Built-in `variants` per model (Anthropic `high`/`max`, OpenAI `none`..`xhigh`, Google `low`/`high`), custom `variants` block, `variant_cycle` keybind; provider `options` (e.g. `reasoningEffort`, Anthropic `thinking.budgetTokens`) under `provider.<p>.models.<m>.options` | No harness-level model fallback or retry documented [verified absent on agents/models pages]. OpenRouter provider routing passes `provider.order` / `allow_fallbacks` as model options (provider-side, not OpenCode) | Agent `model`; unknown agent keys pass through as provider options (doc example `"reasoningEffort": "high"`); agent options override global model options; subagents default to the invoking primary agent's model | `opencode.json`/`.jsonc` (`agent` key); agents `~/.config/opencode/agents/`, `.opencode/agents/` |
| agy / Antigravity CLI (1.3.3) | `--model` (values from `agy models`); `model` key present in `~/.gemini/antigravity-cli/settings.json` (observed key name; not in the settings docs page) | `--effort` `low`/`medium`/`high`/`xhigh`/`max` (from `agy --help`); no settings key documented | None documented. `useG1Credits` is a quota fallback (personal credits), not a model fallback | Subagent frontmatter `model`: `inherit`/`flash`/`pro` (tier, not model ID); no effort field | Global `~/.gemini/antigravity-cli/settings.json`; agents `.agents/agents/<name>.md`, `~/.gemini/config/agents/`; no project settings file documented |

Notes:
- Only Claude Code has a documented, native, runtime cross-model fallback chain (`fallbackModel`). [verified]
- Codex, Pi, OpenCode and agy document same-model retry or nothing; Pi's cross-model chain exists only as the approved, on-hold extension design. [verified for docs read; see partial-read caveat for Codex]
- Effort vocabularies differ: Claude Code `low..xhigh` (+`max` as cap / subagent), Codex adds `ultra`, Pi adds `off`/`minimal`, OpenCode uses per-provider variant names, agy `low..max`. The broker table uses `minimal` for agy `bulk`, which is not in `agy --effort`'s listed values. [verified; mismatch is an observation only]

## Open / unverified

- The `model-routing` skill source was not located; section (b) relies on the 2026-10-12 memory snapshot.
- Codex reference page was read to 100k of ~158k characters; a fallback key in the unread part is not ruled out.
- Pi docs are from a 0.80.2 checkout; installed Pi is 1.1.0 and was not re-checked for new retry/fallback keys.
- agy settings docs do not list a `model` or effort key; the `model` key is only observed in the local settings file.
