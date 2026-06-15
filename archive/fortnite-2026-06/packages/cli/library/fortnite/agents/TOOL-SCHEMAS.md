# Tool Schemas — Required Fields

Central reference for tool parameter schemas. Use this to avoid common schema errors when dispatching agents or calling tools directly.

---

## todowrite

```json
{
  "todos": [
    {
      "content": "string (required)",
      "status": "pending|in_progress|completed|cancelled",
      "priority": "high|medium|low"
    }
  ]
}
```

**Required fields:** `content`, `status`, `priority`

---

## bash

```json
{
  "command": "string (required)",
  "description": "string (required)",
  "timeout": 120000,
  "workdir": "string"
}
```

**Required fields:** `command`, `description`

---

## task

```json
{
  "description": "string (required)",
  "prompt": "string (required)",
  "subagent_type": "wall-builder|loot-hawk|...",
  "command": "string (optional)"
}
```

**Required fields:** `description`, `prompt`, `subagent_type`

---

## read

```json
{
  "filePath": "string (required, absolute path)",
  "offset": 1,
  "limit": 100
}
```

**Required fields:** `filePath` (must be absolute)

---

## filesystem_edit_file

```json
{
  "path": "string (required)",
  "edits": [
    {
      "oldText": "string (required)",
      "newText": "string (required)"
    }
  ],
  "dryRun": false
}
```

**Required fields:** `path`, `edits` (array of `{oldText, newText}`)

---

## morph-mcp_edit_file

```json
{
  "path": "string (required)",
  "instruction": "string (required)",
  "code_edit": "string (required)",
  "dryRun": false
}
```

**Required fields:** `path`, `instruction`, `code_edit`

---

## compress

```json
{
  "topic": "string (required)",
  "content": [
    {
      "startId": "m0001",
      "endId": "m0005",
      "summary": "string (required)"
    }
  ]
}
```

**Required fields:** `topic`, `content` (array of `{startId, endId, summary}`)

---

## Common Mistakes

| Tool | Mistake | Correct |
|------|---------|---------|
| `todowrite` | ❌ Using `text` instead of `content` | ✅ Use `content` |
| `bash` | ❌ Omitting `description` | ✅ Always include `description` |
| `task` | ❌ Using `mode` or `text` as top-level fields | ✅ Use `description`, `prompt`, `subagent_type` |
| `read` | ❌ Using relative paths for `filePath` | ✅ Use absolute paths |
| `filesystem_edit_file` | ❌ Using `oldString`/`newString` | ✅ Use `oldText`/`newText` |
| `morph-mcp_edit_file` | ❌ Omitting `instruction` | ✅ Always include `instruction` |
| `compress` | ❌ Using `text` instead of `topic` | ✅ Use `topic` |

---

## Quick Validation Checklist

Before calling tools, verify:
- [ ] `bash`: `description` field present
- [ ] `todowrite`: `content`, `status`, `priority` all present
- [ ] `task`: `description`, `prompt`, `subagent_type` all present
- [ ] `read`: `filePath` is absolute path
- [ ] `filesystem_edit_file`: `edits` is array with `oldText`/`newText`
- [ ] `morph-mcp_edit_file`: `instruction` and `code_edit` present
