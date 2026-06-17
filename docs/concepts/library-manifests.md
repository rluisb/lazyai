# Embedded library manifests

LazyAI embeds a curated library from `packages/cli/library/`. Some assets are derived from vibe-lab principles or current baseline files, but LazyAI ships repository-local copies and does not require a local `~/code/vibe-lab` checkout at runtime.

Two repository manifests document that boundary:

- `packages/cli/library/manifests/provenance.yaml` covers every active canonical file under `packages/cli/library/canonical/` with its local SHA-256 and source notes.
- `packages/cli/library/manifests/curation.yaml` covers the guarded embedded asset families: canonical agents/skills/hooks, full emitted skills, hook-policy docs, runtime hook assets, root templates, rules, standards, templates, tool templates, and the docs-only workflow catalog.

## Update flow

1. **Upstream sync:** compare the upstream or baseline source outside the default check path, copy or compress the intended content into `packages/cli/library/`, then update `source_repo`, `source_ref`, `source_path`, `mode`, `notes`, and `local_sha256`.
2. **Intentional LazyAI curation:** edit the embedded asset directly, set provenance `mode` to `curated`, `compressed`, or `LazyAI-authored`, and update the curation rationale fields.
3. **Coverage guard:** run the targeted manifest tests before review; they fail when a guarded asset is missing curation/provenance coverage or a canonical hash is stale.

The token-rent budget remains a separate canonical-library size guard; these manifests do not change token-rent behavior.

## Runtime hook assets

Hook policy markdown under `packages/cli/library/hooks/` is the durable policy source.

Generated runtime assets live alongside the supported tool surfaces:

- `packages/cli/library/claudecode/hooks/*.sh`
- `packages/cli/library/copilot/hooks/*.{json,sh}`
- `packages/cli/library/opencode/plugins/vibe-lab-hooks.js`

These are embedded runtime artifacts, not external dependencies.

## Workflow catalog policy

`packages/cli/library/workflows/` is a documentation-only parity surface. It exists so LazyAI can preserve baseline workflow references without reintroducing retired runtime `task`, `workflow`, or orchestration command families.

## Canonical command assets

The files `packages/cli/library/canonical/commands/graphify.md` and `packages/cli/library/canonical/commands/handoff.md` are provenance-covered canonical inventory assets only. They are documented command blueprints and are not emitted as runtime or default slash command definitions.
Actual emitted slash commands continue to come from the tool-specific command directories:
`packages/cli/library/claudecode/commands/` and `packages/cli/library/opencode/commands/`.

Any change to command emission behavior must stay in a separate implementation branch with explicit approval and dedicated verification.
## MCP catalog policy

`packages/cli/library/mcp/catalog.json` now carries only the actively shipped MCP inventory. Do not keep dormant or speculative server placeholders in the catalog; add a server only when the product intends to ship and verify it.

Figma and Slack remain documented exclusions. Do not add catalog placeholders for them until an authoritative server URL/command, authentication shape, and supported client behavior are verified and separately approved.
