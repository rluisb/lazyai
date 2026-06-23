# Implementation Scope — Cycle 5

## Priority A — MCP catalog examples and anti-examples

For canonical MCP servers:

```text
ai-memory
filesystem
ripgrep
codegraph
obsidian
```

Add or validate:

```text
purpose
preferred_use
CLI-first vs MCP-first guidance
examples
anti_examples
security_notes
adapter_output_differences
```

Example:

```json
{
  "name": "ripgrep",
  "purpose": "Fast deterministic code search",
  "preferred_use": "Use for exact code/text search before semantic exploration",
  "examples": [
    "Find all references to a function before editing it",
    "Search for TODOs in a package"
  ],
  "anti_examples": [
    "Do not use semantic codegraph when exact symbol search is enough",
    "Do not use MCP filesystem for bulk search when rg is available"
  ]
}
```

Acceptance:

```text
- Catalog remains valid.
- Generated MCP output remains compatible.
- Tool descriptions improve without bloating always-loaded instructions.
```

## Priority B — Adapter capability documentation

Add/update:

```text
docs/adapters/capability-matrix.md
docs/adapters/opencode.md
docs/adapters/claude.md
docs/adapters/copilot.md
docs/adapters/pi.md
docs/adapters/omp.md
docs/adapters/antigravity.md
docs/adapters/kiro.md
```

Each adapter doc should include:

```text
status: stable/beta/partial
generated files
supported asset types
unsupported asset types
MCP behavior
hook behavior
skill behavior
agent behavior
known limitations
tests/fixtures proving behavior
```

Acceptance:

```text
- OMP and Antigravity beta status is visible and justified.
- Pi MCP no-op is explicit and intentional.
- Kiro limitations are explicit.
- Capability matrix matches code.
```

## Priority C — Official compliance matrix refresh

Update existing compliance matrix if present:

```text
docs/lazyai-vibelab-product-spec-pack/03_OFFICIAL_TOOL_COMPLIANCE_MATRIX*
```

Include:

```text
OpenCode
Claude Code
GitHub Copilot
Pi
OMP
Gemini / Antigravity
Kiro
```

If external web access is unavailable, mark:

```text
requires external docs refresh
```

Do not invent official requirements.

Acceptance:

```text
- Matrix matches current code.
- Gaps are explicit.
- Beta graduation criteria are documented.
```

## Priority D — Token-rent and minimality docs

Document:

```text
internal/tokenrent
internal/minimality
curation token_rent flags
minimality report behavior
```

Add or update:

```text
docs/concepts/token-rent.md
docs/concepts/minimality.md
```

Acceptance:

```text
- Users understand why some assets are excluded or progressive-disclosure only.
- Validation/report output is actionable.
- No breaking changes.
```
