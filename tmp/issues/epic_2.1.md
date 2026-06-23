**Epic — RPI Cycle 2.1: Finish semantic validation integration**

Goal: Make the Cycle 2 semantic-validation docs/templates discoverable and produce a cleanup path for shipped assets. Small cycle — no mass asset rewrite.

### Why
Cycle 2 added semantic validation (`internal/validate/validate.go`), concept docs (`docs/concepts/skill-quality.md`, `docs/concepts/agent-contracts.md`), and templates (`library/templates/skill-quality.md`, `agent-contract.md`). They are not yet linked or curated, and shipped assets have not been inventoried for warnings.

### Tasks (sub-issues)
- Link semantic-validation docs from harness-principles / README / KNOWLEDGE_MAP
- Decide + apply template curation/provenance for the two new templates
- Produce semantic-validation warning inventory for shipped assets
- Minor semantic-validation wording refinements (stable rule IDs)

### Boundary (02_SHARED_BOUNDARIES)
LazyAI stays a canonical-`.ai` manager / validation layer / compiler. No runtime/orchestration, no judge/eval runner, no Codex adapter. Final report must confirm boundary intact.

### Definition of done
Docs discoverable + non-conflicting; templates curated per repo convention; warning inventory committed with true/false-positive notes; `go test ./packages/cli/...` green.
