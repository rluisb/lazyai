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
