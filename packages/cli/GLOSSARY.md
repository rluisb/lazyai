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

**Agent Plugin**:
A directory conforming to the Agent Plugins standard, with portable skills and MCP config. Both the project source and the LazyAI library are Agent Plugins.
_Avoid_: bundle, package, `.ai/` tree

**Project source**:
The `.lazyai/` directory in a project: a valid Agent Plugin that the project author edits and LazyAI compiles.
_Avoid_: `.ai/` tree, canonical tree, source plugin

**LazyAI library**:
The Agent Plugin shipped inside `lazyai-cli` that a project source references and LazyAI merges in at compile.
_Avoid_: packs, starter, library content (when meaning a copy in the project)

**Namespace folder**:
The `io.github.rluisb.lazyai/` folder inside an Agent Plugin, holding the parts the Agent Plugins standard does not cover.
_Avoid_: extension dir, LazyAI dir
