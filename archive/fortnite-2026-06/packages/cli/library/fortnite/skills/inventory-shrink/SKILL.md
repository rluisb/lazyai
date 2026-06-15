---
name: inventory-shrink
description: Compress MCP tool descriptions to save tokens. Wraps any MCP server, compresses tool schemas. ~60% fewer tokens per tool description.
trigger: /inventory-shrink
skill_path: skills/inventory-shrink
scripts:
  - name: tool-shrink.sh
    description: Compress MCP tool descriptions
    path: scripts/tool-shrink.sh
---

# Inventory Shrink — MCP Tool Description Compression

## Purpose
MCP tool descriptions cost tokens every turn. Inventory Shrink compresses tool schemas — drops filler words, keeps substance. Same functionality, ~60% fewer tokens per description.

**Use when:**
- Setting up new MCP server — "shrink the tool descriptions"
- Before long session — "compress all tool schemas"
- Token budget tight — "shrink inventory"

## Scripts

| Script | Purpose | Key Flags |
|--------|---------|-----------|
| `tool-shrink.sh` | Compress MCP tool descriptions | `--input <json>`, `--output <json>`, `--level lite|full|ultra` |

## Workflow

### Step 1: Export Tool Descriptions
Get current tool schemas from MCP server:
```bash
# Example: export from morph-mcp
morph-mcp tools list --json > tools.json
```

### Step 2: Compress
```bash
./scripts/tool-shrink.sh --input tools.json --output tools-shrunk.json --level full
```

### Step 3: Use Compressed Schema
Replace original tool descriptions with compressed versions in MCP config.

## Compression Levels

| Level | Description | Savings |
|-------|-------------|---------|
| lite | Drop filler words, keep structure | ~30% |
| full | Fragments, no articles, direct | ~60% |
| ultra | Telegraphic, symbols only | ~80% |

## Example

**Before (69 tokens):**
```json
{
  "description": "The reason your React component is re-rendering is likely because you're creating a new object reference on each render cycle."
}
```

**After (19 tokens):**
```json
{
  "description": "New object ref each render. Inline object prop = new ref = re-render. Wrap in useMemo."
}
```

## Integration with Other Skills

- **caveman**: Uses same compression grammar
- **zero-point**: Validates compressed tools still work
- **battle-bus**: Compressed tools save tokens in long workflows

## Tips

- Test compressed tools before deploying
- Keep code examples, URLs, paths byte-preserved
- Use `--level lite` for critical tools, `full` for others
- Compression is reversible — original descriptions preserved
- Run shrink after MCP server updates
