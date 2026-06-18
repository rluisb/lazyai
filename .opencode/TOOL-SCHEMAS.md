# Tool Schemas — Required Fields

Reference for correct tool usage. Prevents schema errors.

## todowrite

```json
{
  "todos": [
    {
      "content": "Brief task description",
      "status": "pending",
      "priority": "medium"
    }
  ]
}
```

**Fields:**
- `todos`: array, required
- `todos[].content`: string, required
- `todos[].status`: string, required — `pending|in_progress|completed|cancelled`
- `todos[].priority`: string, required — `high|medium|low`

**Common mistake:** Using `text` instead of `content`.

---

## bash

```json
{
  "command": "git status",
  "description": "Shows working tree status",
  "timeout": 120000,
  "workdir": "/path"
}
```

**Fields:**
- `command`: string, required
- `description`: string, required (5-10 words)
- `timeout`: integer, optional (ms)
- `workdir`: string, optional

**Common mistake:** Omitting `description`.

---

## task

```json
{
  "description": "Short task label",
  "prompt": "Detailed task instructions",
  "subagent_type": "wall-builder"
}
```

**Fields:**
- `description`: string, required
- `prompt`: string, required
- `subagent_type`: string, required
- `command`: string, optional (metadata)

**Common mistake:** Using `mode` or `text` as top-level fields.

---

## read

```json
{
  "filePath": "/absolute/path/to/file",
  "offset": 1,
  "limit": 100
}
```

**Fields:**
- `filePath`: string, required (absolute path)
- `offset`: integer, optional
- `limit`: integer, optional

**Common mistake:** Using relative paths.

---

## filesystem_edit_file

```json
{
  "path": "/absolute/path",
  "edits": [
    {
      "oldText": "exact existing text",
      "newText": "replacement text"
    }
  ],
  "dryRun": false
}
```

**Fields:**
- `path`: string, required
- `edits`: array, required
- `edits[].oldText`: string, required
- `edits[].newText`: string, required
- `dryRun`: boolean, optional

**Common mistake:** Using `oldString`/`newString`.

---

## morph-mcp_edit_file

```json
{
  "path": "/absolute/path",
  "instruction": "I am changing X",
  "code_edit": "// ... existing code ...\nchanged lines\n// ... existing code ...",
  "dryRun": false
}
```

**Fields:**
- `path`: string, required
- `instruction`: string, required
- `code_edit`: string, required
- `dryRun`: boolean, optional

**Common mistake:** Omitting `instruction`.

---

## compress

```json
{
  "topic": "Short label",
  "content": [
    {
      "startId": "m0001",
      "endId": "m0005",
      "summary": "Dense technical summary"
    }
  ]
}
```

**Fields:**
- `topic`: string, required
- `content`: array, required
- `content[].startId`: string, required
- `content[].endId`: string, required
- `content[].summary`: string, required

**Common mistake:** Using `text` instead of `topic`.

---

## Validation Checklist

Before calling tools, verify:
- [ ] `bash`: `description` field present
- [ ] `todowrite`: `content`, `status`, `priority` all present
- [ ] `task`: `description`, `prompt`, `subagent_type` all present
- [ ] `read`: `filePath` is absolute path
- [ ] `edit`: `oldText`/`newText` used (not `oldString`/`newString`)
