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
