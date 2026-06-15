# Embedded library manifests

LazyAI embeds a curated library from `packages/cli/library/`. Some assets are derived from vibe-lab principles or historical assets, but LazyAI ships repository-local copies and does not require a local `~/code/vibe-lab` checkout at runtime.

Two repository manifests document that boundary:

- `packages/cli/library/manifests/provenance.yaml` covers every active file under `packages/cli/library/canonical/` with its local SHA-256 and source notes.
- `packages/cli/library/manifests/curation.yaml` covers the guarded embedded asset families: agents, skills, hooks, root templates, rules, standards, templates, and tool templates.

## Update flow

1. **Upstream sync:** compare the upstream source outside the default check path, copy or compress the intended content into `packages/cli/library/`, then update `source_repo`, `source_ref`, `source_path`, `mode`, `notes`, and `local_sha256`.
2. **Intentional LazyAI curation:** edit the embedded asset directly, set provenance `mode` to `curated`, `compressed`, or `LazyAI-authored`, and update the curation rationale fields.
3. **Coverage guard:** run the targeted manifest tests before review; they fail when a guarded asset is missing curation/provenance coverage or a canonical hash is stale.

The token-rent budget remains a separate canonical-library size guard; these manifests do not change token-rent behavior.

## Canonical command assets

The provenance manifest hash-covers `packages/cli/library/canonical/commands/graphify.md` and `packages/cli/library/canonical/commands/handoff.md`. These files are library inventory assets — they document compressed command prompts available for reference — but they are not active command emission sources. The adapter does not emit them as slash commands to generated tool setups by default.

Command emission changes require a separate focused implementation branch with its own verification.

## MCP catalog policy

`packages/cli/library/mcp/catalog.json` may include opt-in MCP servers that are not active setup defaults. These entries must keep `enabled: false` unless a separate product decision promotes them.

Plan C adds disabled remote entries for Context7 and GitHub MCP. The GitHub MCP entry supplements the existing `gh` CLI catalog entry; it does not replace CLI-first GitHub workflows.

Figma and Slack remain documented exclusions for now. Do not add catalog placeholders for them until an authoritative server URL/command, authentication shape, and supported client behavior are verified and separately approved.
