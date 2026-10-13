---
status: accepted
---

# ADR-009: The project source is an Agent Plugin; the library is referenced, not copied

**Date:** 2026-10-12. Decided in [LazyAI 2.0 source layout and namespace](https://github.com/rluisb/lazyai/issues/634) on the [LazyAI 2.0 map](https://github.com/rluisb/lazyai/issues/632).

LazyAI 2.0 replaces the 1.x `.ai/` tree (ADR-006) with a project source at `.lazyai/` that is a strictly valid Agent Plugins 1.0.0 plugin. Its settings live in `plugin.json` under `extensions.io.github.rluisb.lazyai`, and the parts the standard does not cover live in the `io.github.rluisb.lazyai/` namespace folder. We chose this because Codex, Cursor and Kiro already load Agent Plugins natively, so a valid project source is usable without compiling. Putting it under `.agents/` would have made Codex, Pi, agy and Cursor load the raw source as well as the compiled output.

The LazyAI library ships inside `lazyai-cli` as its own Agent Plugin, versioned with the CLI. The project source references it, and compile merges it in, where a project item with the same name replaces the library's item entirely. We chose this over copying the library into each project, which is what 1.x packs do, because copies drift silently and can't be upgraded.

## Consequences

- This supersedes ADR-008's use of `.lazyai/` as the sidecar directory. The sidecar (`sidecar.yaml`, `~/.lazyai/`), `secret` and token-rent code move to the library-cut delete inventory, because 2.0 owns no runtime state. Migration warns when it finds an existing `sidecar.yaml`.
- References accept only the bundled library and local-path Agent Plugins. Remote sources (git, registries) are out of scope for 2.0.
