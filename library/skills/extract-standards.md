---
name: extract-standards
description: Extract coding standards and conventions from a codebase.
trigger: /extract-standards
phase: post-task
preset: standard
---

# Extract Standards Skill

## Purpose
Scan the codebase and extract real patterns into specs/standards/ files. Standards are descriptive — they show HOW we do things here with real code references.

## Usage

```
/extract-standards [category]           # Extract standards for a specific category
/extract-standards                       # Suggest which categories to extract first
/extract-standards --refresh             # Check existing standards for drift
/extract-standards api-service testing   # Workspace: extract from specific repo
```

## Workflow

### Without category — suggest extraction order
1. Scan project structure (file types, directories, counts)
2. Suggest categories by code volume:
   - test files found → testing
   - API routes/controllers → coding
   - auth middleware → security
   - data models/migrations → data
   - module boundaries → architecture
3. Ask user which category to start with

### With category — extract patterns
1. Scan files matching the category (e.g., `**/*.test.*` for testing)
2. Identify recurring patterns (2+ files using the same approach)
3. For each pattern found, present:
   - **Pattern name**: descriptive title
   - **Reference**: real file path (canonical example)
   - **Description**: 2-3 sentences of what the pattern does
4. Ask user: accept / skip / edit for each pattern
5. Write accepted patterns to `specs/standards/{category}/` using the standard template

### With --refresh — check for drift
1. Read existing standards from specs/standards/
2. For each standard, check if the reference file still exists
3. If moved: suggest updated path
4. If deleted: flag as potentially outdated
5. If changed significantly: flag for review
6. Report: ✅ current / ⚠️ drifted / ❌ reference missing

## Categories

| Category | What to scan | Example patterns |
|----------|-------------|-----------------|
| testing | `**/*.test.*`, `**/*.spec.*` | Unit test structure, mock conventions, fixture patterns |
| coding | Controllers, services, utils | API patterns, service layer, error handling |
| security | Auth middleware, validation | Auth flow, input validation, secret handling |
| architecture | Module boundaries, imports | Module structure, cross-module communication |
| data | Models, migrations, schemas | Entity patterns, query patterns, migration style |
| quality | Naming, file organization | Naming conventions, file structure |
| observability | Logging, metrics, health | Logging patterns, metric naming |
| resilience | Retries, circuit breakers | Timeout handling, retry policies |

## Output Format

Each extracted standard follows specs/templates/standard-template.md:

```markdown
# [Category] — [Pattern Name]

## Pattern
[2-3 sentence description]

## Reference
- `path/to/canonical/example.ts` (canonical example)

## Rules
- [Convention 1]
- [Convention 2]

## Extracted
- Date: [today]
- Source: automated extraction, reviewed by [user]
```

## Workspace Mode

For workspace scope, specify the repo:
```
/extract-standards api-service testing
```

User chooses where each pattern is stored:
- **Repo-specific**: `specs/standards/{repo-name}/{category}/`
- **Cross-repo**: `specs/standards/cross-repo/{category}/`

## YAGNI Gate

Before extracting, ask:
- Does this category have enough code to extract patterns from? (minimum 3+ files)
- Are the patterns consistent enough to be standards? (2+ files following the same approach)
- If the codebase is new (<5 features), skip — standards grow organically from real code.

## Integration
- Agent: Scout (reads code) + Documenter (writes standards)
- Requires: existing codebase with real code
- Output: specs/standards/{category}/*.md
- Related: specs/standards/AGENTS.md (progressive loading table — update after extraction)
