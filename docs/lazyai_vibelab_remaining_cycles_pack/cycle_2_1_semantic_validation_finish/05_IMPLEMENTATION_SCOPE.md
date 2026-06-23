# Implementation Scope — Cycle 2.1

## Priority A — Link semantic validation docs

Add links from:

```text
docs/concepts/harness-principles.md
README.md if appropriate
specs/KNOWLEDGE_MAP.md if appropriate
```

To:

```text
docs/concepts/skill-quality.md
docs/concepts/agent-contracts.md
```

Acceptance:

```text
- Docs are discoverable.
- No duplicate/conflicting doctrine.
- LazyAI still described as compiler/asset manager.
```

## Priority B — Curate templates if appropriate

Inspect existing curation rules before editing.

If project conventions require templates in curation/provenance, add:

```text
packages/cli/library/templates/skill-quality.md
packages/cli/library/templates/agent-contract.md
```

Acceptance:

```text
- Manifest validation still passes.
- Templates are discoverable by library tooling.
```

## Priority C — Warning inventory

Run validation against shipped assets if there is an existing safe way.

Capture a warning inventory such as:

```text
rule id
asset path
count
likely true positive / likely false positive
recommended future cleanup
```

Store it in one of:

```text
docs/reports/semantic-validation-warning-inventory.md
specs/029-lazyai-v2/semantic-validation-warning-inventory.md
```

Follow existing repo conventions.

Acceptance:

```text
- No mass asset rewrite.
- Warning inventory gives future cleanup path.
- False positives are noted.
```

## Priority D — Minor wording refinements

Only adjust semantic validation wording if needed for clarity.

Acceptance:

```text
- Rule IDs stay stable.
- Tests still pass.
```
