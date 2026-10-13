# CLI

`lazyai-cli`: turns one canonical source of agent configuration into each harness's native files.

## Language

**LazyAI 2.0**:
The `lazyai-cli` 2.0.0 major release whose canonical source is an Agent Plugin.
_Avoid_: LazyAI v2, v2 (spec 029 already used that name)

**Manifest compile**:
The 1.x model (spec 029): a `.ai/` tree with `lazyai.json` and `lock.json` compiled to each target.
_Avoid_: v2, V2 compile

**Harness**:
An agent CLI that LazyAI emits configuration for, such as Claude Code, Codex, Pi, OpenCode or agy.
_Avoid_: target, tool, adapter, CLI (when meaning the agent CLI)

**Harness tier**:
LazyAI's testing and maintenance commitment for a harness, distinct from its native format support. Tier 1 covers supported capabilities with compiler tests and harness smoke checks; tier 2 is provisional and experimental pending runtime verification.
_Avoid_: compatibility level, feature parity

**Context file**:
The standing-instructions file a harness loads at session start. In LazyAI 2.0 the root AGENTS.md is the one compiled context file; CLAUDE.md only imports it.
_Avoid_: memory file, rules file, instructions file

**Agent Plugin**:
A directory conforming to the Agent Plugins standard, with portable skills and MCP config. Both the project source and the LazyAI library are Agent Plugins.
_Avoid_: bundle, package, `.ai/` tree

**Project source**:
The `.lazyai/` directory in a project: a valid Agent Plugin that the project author edits and LazyAI compiles.
_Avoid_: `.ai/` tree, canonical tree, source plugin

**LazyAI library**:
The Agent Plugin shipped inside `lazyai-cli` that a project source references and LazyAI merges in at compile.
_Avoid_: packs, starter, library content (when meaning a copy in the project)

**Skill core**:
The fixed set of skills the LazyAI library ships in 2.0, each covering one step of the engineering loop or LazyAI authoring.
_Avoid_: skill pack, starter skills, speckit

**Router skill**:
The `lazyai` skill that tells the agent which skill fits the user's current step; it suggests, it never dispatches.
_Avoid_: orchestrator, dispatcher, rpi

**Owned content**:
A whole file, a marked region, or a config key that LazyAI wrote and records in the lock; everything else in the project belongs to the user.
_Avoid_: managed file, generated file (when meaning ownership)

**Lock**:
The committed record of every piece of owned content and the hash it had when LazyAI last wrote it.
_Avoid_: lockfile store, manifest, journal

**Drift**:
Owned content whose current hash no longer matches the lock, meaning someone edited it after LazyAI wrote it.
_Avoid_: conflict, stale (stale means the source changed, not the output)

**Namespace folder**:
The `io.github.rluisb.lazyai/` folder inside an Agent Plugin, holding the parts the Agent Plugins standard does not cover.
_Avoid_: extension dir, LazyAI dir
