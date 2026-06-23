# Research Phase — Cycle 2.1

Inspect the result of Cycle 2.

## Required inspection

```text
docs/concepts/harness-principles.md
docs/concepts/skill-quality.md
docs/concepts/agent-contracts.md
packages/cli/library/templates/skill-quality.md
packages/cli/library/templates/agent-contract.md
packages/cli/library/manifests/curation.yaml
packages/cli/library/manifests/provenance.yaml
packages/cli/internal/validate/validate.go
packages/cli/internal/validate/validate_test.go
README.md
specs/KNOWLEDGE_MAP.md
```

Answer:

```text
- Are skill-quality and agent-contract docs linked from harness-principles?
- Are the new templates referenced by curation/provenance if repo conventions require it?
- Does validate output expose semantic warnings clearly?
- What warnings appear when validating shipped library assets?
- Are any warnings accidental false positives?
- Are any docs duplicated or conflicting?
```

Run if safe:

```bash
go test -v ./packages/cli/internal/validate/...
go test ./packages/cli/...
```
