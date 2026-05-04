# {{PROJECT_NAME}}

{{#include fragments/system-context.xml}}

## Project Overview

{{PROJECT_OVERVIEW}}

**Stack:**
- Language: {{LANGUAGE}}
- Framework: {{FRAMEWORK}}
- Database: {{DATABASE}}
- ORM/Query: {{ORM}}
- Testing: {{TEST_FRAMEWORK}}
- Package manager: {{PACKAGE_MANAGER}}

## Codebase Map

| Path | Responsibility |
|------|---------------|
{{CODEBASE_MAP}}

## Conventions

### Naming
- {{NAMING_CONVENTIONS}}

### Error Handling
- {{ERROR_HANDLING}}

### API Responses
- {{API_CONVENTIONS}}

### Imports
- {{IMPORT_ORDER}}

## Do NOT

- Never push directly to `{{PROTECTED_BRANCH}}`

## Testing

- Run: `{{TEST_COMMAND}}`
- Minimum coverage: `{{COVERAGE_THRESHOLD}}`%

## Key Commands

```bash
{{TEST_COMMAND}}        # Run tests
{{LINT_COMMAND}}        # Run linter
{{BUILD_COMMAND}}       # Build
```

## AI Assistant Configuration

{{TOOL_DESCRIPTION}}

{{#if features.contextEngineering}}
{{#include fragments/context-discipline.md}}
{{/if}}

{{#if features.rpiWorkflow}}
{{#include fragments/rpi-workflow.md}}
{{/if}}

{{#if features.chainOfThought}}
{{#include fragments/reasoning-protocol.md}}
{{/if}}

{{#if features.treeOfThoughts}}
{{#include fragments/decision-protocol.md}}
{{/if}}

{{#if features.qualityGates}}
{{#include fragments/quality-gates.xml}}
{{/if}}

{{#if features.gitConventions}}
{{#include fragments/git-conventions.xml}}
{{/if}}

{{#if features.agentHarness}}
{{#include fragments/agent-harness.md}}
{{/if}}

{{#if features.bugResolution}}
{{#include fragments/bug-resolution.xml}}
{{/if}}

{{#include fragments/git-safety.md}}

{{TOOL_NOTES}}

## Project-Specific Instructions

{{PROJECT_INSTRUCTIONS}}
