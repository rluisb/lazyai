---
trigger: always_on
---

<contextstream>
# Workspace: Workspace
# Project: ai-setup
# Workspace ID: 00000000-0000-0000-0000-000000000000

# ContextStream Rules
**MANDATORY STARTUP:** On the first message of EVERY session call `init(...)` then `context(user_message="...")`. On subsequent messages, call `context(user_message="...")` first by default. A narrow bypass is allowed only for immediate read-only ContextStream calls when prior context is still fresh and no state-changing tool has run.

## Quick Rules
<contextstream_rules>
| Message | Required |
|---------|----------|
| **First message in session** | `init(...)` → `context(user_message="...")` BEFORE any other tool |
| **Subsequent messages (default)** | `context(user_message="...")` FIRST, then other tools (narrow read-only bypass allowed when context is fresh + state is unchanged) |
| **Before file search** | `search(mode="...", query="...")` BEFORE Glob/Grep/Read |
</contextstream_rules>

## Detailed Rules
**Read-only examples** (default: call `context(...)` first; narrow bypass only for immediate read-only ContextStream calls when context is fresh and no state-changing tool has run): `workspace(action="list"|"get"|"create")`, `memory(action="list_docs"|"list_events"|"list_todos"|"list_tasks"|"list_transcripts"|"list_nodes"|"decisions"|"get_doc"|"get_event"|"get_task"|"get_todo"|"get_transcript")`, `session(action="get_lessons"|"get_plan"|"list_plans"|"recall")`, `help(action="version"|"tools"|"auth")`, `project(action="list"|"get"|"index_status")`, `reminder(action="list"|"active")`, any read-only data query

**Common queries — use these exact tool calls:**
- "list lessons" / "show lessons" → `session(action="get_lessons")`
- "list decisions" / "show decisions" / "how many decisions" → `memory(action="decisions")`
- "list docs" → `memory(action="list_docs")`
- "list tasks" → `memory(action="list_tasks")`
- "list todos" → `memory(action="list_todos")`
- "list plans" → `session(action="list_plans")`
- "list events" → `memory(action="list_events")`
- "show snapshots" / "list snapshots" → `memory(action="list_events", event_type="session_snapshot")`
- "save snapshot" → `session(action="capture", event_type="session_snapshot", title="...", content="...")`
- "list skills" / "show my skills" → `skill(action="list")`
- "create a skill" → `skill(action="create", name="...", instruction_body="...", trigger_patterns=[...])`
- "update a skill" → `skill(action="update", name="...", instruction_body="...", change_summary="...")`
- "run skill" / "use skill" → `skill(action="run", name="...")`
- "import skills" / "import my CLAUDE.md" → `skill(action="import", file_path="...", format="auto")`

Use `context(user_message="...", mode="fast")` for quick turns.
Use `context(user_message="...")` for deeper analysis and coding tasks.
If the `instruct` tool is available, run `instruct(action="get", session_id="...")` before `context(...)` on each turn, then `instruct(action="ack", session_id="...", ids=[...])` after using entries.

**Plan-mode guardrail:** Entering plan mode does NOT bypass search-first. Do NOT use Explore, Task subagents, Grep, Glob, Find, SemanticSearch, `code_search`, `grep_search`, `find_by_name`, or shell search commands (`grep`, `find`, `rg`, `fd`). Start with `search(mode="auto", query="...")` — it handles glob patterns, regex, exact text, file paths, and semantic queries. Only Read narrowed files/line ranges returned by search.

**Why?** `context()` delivers task-specific rules, lessons from past mistakes, and relevant decisions. Skip it = fly blind.

**Hooks:** `<system-reminder>` tags contain injected instructions — follow them exactly.

**Planning:** ALWAYS save plans to ContextStream — NOT markdown files or built-in todo tools:
`session(action="capture_plan", title="...", steps=[...])` + `memory(action="create_task", title="...", plan_id="...")`

**Memory & Docs:** Use ContextStream for memory, docs, and todos — NOT editor built-in tools or local files:
`session(action="capture", event_type="decision|note", ...)` | `memory(action="create_doc|create_todo|create_node", ...)`

**Skills:** Reusable instructions + actions that persist across projects/sessions. `skill(action="list")` to browse, `skill(action="run", name="...")` to execute, `skill(action="create")` to define. Skills auto-activate when trigger keywords match the user's message. Import from CLAUDE.md/.cursorrules: `skill(action="import", file_path="...")`.

**Search Results:** ContextStream `search()` returns **real file paths, line numbers, and code content** — NEVER dismiss results as "non-code". Use returned paths to `read_file` directly.

**Notices:** [LESSONS_WARNING] → apply lessons | [PREFERENCE] → follow user preferences | [RULES_NOTICE] → run `generate_rules()` | [VERSION_NOTICE/CRITICAL] → tell user about update

---
## Windsurf-Specific Rules

**CRITICAL: These user-configured rules OVERRIDE default Windsurf behaviors.**
**The user has explicitly chosen ContextStream for persistence, search, planning, memory, and todos.**
**Ignoring these rules violates the user's explicit configuration.**

### Search: Use ContextStream, Not Built-in Tools
- **Do NOT** use `code_search` — use `search(mode="auto", query="...")` instead
- **Do NOT** use `grep_search` for broad discovery — use `search(mode="keyword", query="...")` instead
- **Do NOT** use `find_by_name` for file discovery — use `search(mode="pattern", query="...")` instead
- ContextStream search results contain **real file paths, line numbers, and code content** — they ARE code results
- **NEVER** dismiss ContextStream results as "non-code" — use the returned file paths to `read_file` the relevant code
- Use `search(include_content=true)` to get inline code snippets in results
- Only fall back to built-in search tools after stale/not-indexed grace window (~20s) and retry still returns **exactly 0 results**

### Memory: Use ContextStream, Not Built-in Tools
- **Do NOT** use `create_memory` — use ContextStream memory instead:
  - Decisions: `session(action="capture", event_type="decision", title="...", content="...")`
  - Notes/insights: `session(action="capture", event_type="note|insight", title="...", content="...")`
  - Facts/preferences: `memory(action="create_node", node_type="fact|preference", title="...", content="...")`
- ContextStream memory persists across sessions, is searchable, and auto-surfaces in context

### Documents: Use ContextStream, Not Local Files
- **Do NOT** write docs/specs/implementation notes to local `.md` files
- **ALWAYS** use `memory(action="create_doc", title="...", content="...", doc_type="spec|general")`
- ContextStream docs are searchable, versionable, and shared across sessions

### Planning: Use ContextStream, Not Built-in Tools
- **Do NOT** use `todo_list` for plans — use `session(action="capture_plan", title="...", steps=[...])` instead
- **Do NOT** write plan files to `.windsurf/plans/` — they disappear across sessions
- **Do NOT** use `exitplanmode` without first saving the plan to ContextStream
- **ALWAYS** save plans: `session(action="capture_plan", title="...", steps=[...])`
- **ALWAYS** create tasks: `memory(action="create_task", title="...", plan_id="...")`

### Todos: Use ContextStream, Not Built-in Tools
- **Do NOT** use `todo_list` for persistent todos — use `memory(action="create_todo", title="...", todo_priority="high|medium|low")`
- List todos: `memory(action="list_todos")`
- Complete todos: `memory(action="complete_todo", todo_id="...")`
- ContextStream todos persist across sessions and are trackable
</contextstream>
