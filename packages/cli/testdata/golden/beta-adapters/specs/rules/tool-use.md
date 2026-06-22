# Tool Use Discipline

## Purpose
Guide AI agents in selecting, invoking, and handling tools/functions correctly.

## Rules

### Tool Selection
- Always use the narrowest tool that satisfies the need
- Prefer read-only tools before write tools (explore before modifying)
- Use dedicated tools over shell commands when available (e.g., use file-read tool, not `cat`)
- Check tool availability before attempting invocation

### Function Naming & Descriptions
- Tool names should be self-documenting (e.g., `search_files` not `sf`)
- Read tool descriptions fully before invoking — parameter semantics matter
- When multiple tools could work, prefer the one with the most specific description

### Parameter Discipline
- Always provide required parameters — never rely on defaults for critical arguments
- Use exact values from prior tool outputs (file paths, IDs) — do not guess or reconstruct
- Quote file paths containing spaces
- Validate parameter types match expectations (string vs array vs object)

### Error Handling
- Check tool return status before proceeding
- On tool failure: read the error message, adjust parameters, retry once
- After two failures with the same tool: try an alternative approach
- Never silently ignore tool errors — log them in the task journal

### MCP Server Awareness
- Know which MCP servers are available in your session
- Prefer server-native tools over shell equivalents (e.g., `ripgrep.search` over `grep`)
- Remote servers may have latency — batch requests when possible
- API-key-dependent servers may fail silently — verify connectivity on first use

### Anti-Patterns
- ❌ Invoking tools speculatively without reading their description
- ❌ Using write tools before understanding current state via read tools
- ❌ Hardcoding paths or IDs instead of using prior tool output
- ❌ Chaining 5+ tool calls without checking intermediate results
- ❌ Using interactive tools (requiring stdin) in non-interactive contexts

## Enforcement
- Reviewers should check tool invocation patterns in agent traces
- Anti-speculation skill should flag speculative tool usage
- Tool errors in task journals should trigger review of invocation discipline

## Defining Tools for Agent Use

When authoring tool/function definitions that LLMs will consume:

### Schema Quality
- **Names:** Use descriptive verb-noun pairs (`search_repositories`, not `search`)
- **Descriptions:** Explain *when* to call, not just *what* it does ("Use when you need to find files matching a pattern" > "Searches for files")
- **Parameters:** Include type, constraints, and examples in descriptions
- **Required vs optional:** Mark clearly; provide sensible defaults for optional params

### Error Responses
- Return structured error objects, not raw strings
- Include actionable guidance: "File not found at X. Did you mean Y?"
- Distinguish retriable errors (timeout) from permanent errors (not found)

### Example Schema Pattern
```json
{
  "name": "search_code",
  "description": "Search repository code for a pattern. Use when you need to find where a function is defined or where a pattern is used across the codebase.",
  "parameters": {
    "pattern": { "type": "string", "description": "Regex pattern to search for (e.g., 'function handleAuth')" },
    "path": { "type": "string", "description": "Directory to search in. Defaults to repo root.", "default": "." },
    "maxResults": { "type": "number", "description": "Max results to return (1-100)", "default": 20 }
  }
}
```
