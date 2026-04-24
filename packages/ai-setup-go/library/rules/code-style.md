# Code Style Rule

## Rule

All code must follow the project's established style conventions. Consistency is enforced by automated tools, not manual review.

## Rationale

Consistent code style reduces cognitive load during reviews, prevents bikeshedding, and makes the codebase approachable for new contributors. Automated enforcement eliminates subjective debates.

## Guidelines

### Automated Formatting
- All code must pass the project's configured formatter before commit
- Formatting is non-negotiable — use the tool's defaults or project config
- Never mix formatting changes with functional changes in the same commit

### Naming Conventions
- Follow the language's community conventions (camelCase for JS/TS, snake_case for Ruby/Python/Go)
- Names should be descriptive and self-documenting
- Avoid abbreviations unless they're universally understood in the domain

### File Organization
- One primary concept per file
- Related files grouped by feature or domain, not by type
- Import ordering: external dependencies → internal modules → relative imports

### Code Clarity
- Prefer explicit over clever
- Comments explain WHY, not WHAT — the code should explain what
- Functions should do one thing and be named accordingly
- Maximum function length: if it needs scrolling, it needs splitting

### Commit Hygiene
- Each commit represents one logical change
- Commit messages follow project convention (e.g., `type(scope): description`)
- No TODO/FIXME/HACK committed without a corresponding task or ticket

## Enforcement

- Linter and formatter run on pre-commit hook and CI
- Code review focuses on logic and design, not style (automated tools handle style)
- Reviewer agent flags style issues only when automated tools miss them
