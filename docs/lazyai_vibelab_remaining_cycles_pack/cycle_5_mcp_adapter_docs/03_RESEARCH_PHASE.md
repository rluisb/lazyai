# Research Phase — Cycle 5

Inspect MCP, adapter docs, compliance matrix, and minimality/token-rent.

## Required inspection

```text
packages/cli/library/mcp/catalog.json
packages/cli/internal/adapter/mcp_compiler.go
packages/cli/internal/scaffold/mcp.go
packages/cli/internal/adapter/capabilities.go
packages/cli/internal/adapter/output_mapping.go
packages/cli/internal/tokenrent/
packages/cli/internal/minimality/
packages/cli/library/manifests/curation.yaml
docs/lazyai-vibelab-product-spec-pack/
docs/adapters/
specs/029-lazyai-v2/
```

Answer:

```text
- Does MCP catalog include examples and anti-examples?
- How does each adapter emit MCP config?
- Which adapters are beta?
- What does Pi intentionally not emit?
- What does Kiro intentionally omit?
- Is there an adapter capability matrix?
- Is there an official tool compliance matrix?
- How are token-rent and minimality documented?
```
