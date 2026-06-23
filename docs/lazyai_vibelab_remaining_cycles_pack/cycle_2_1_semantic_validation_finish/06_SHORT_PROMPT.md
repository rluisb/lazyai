# Short Prompt — Cycle 2.1

Finish Cycle 2 cleanly.

Work on:

```text
- link skill-quality and agent-contract docs from harness-principles/README/knowledge map where appropriate
- decide whether new templates must be added to curation/provenance
- run validation against shipped assets
- create a warning inventory report
- refine warning text only if needed
```

Do not rewrite all skills. Do not add runtime behavior.

Run:

```bash
go test -v ./packages/cli/internal/validate/...
go test ./packages/cli/...
```

Final report must confirm LazyAI remains a compiler/asset manager.
