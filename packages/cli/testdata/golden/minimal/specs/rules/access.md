# Access Rule

## Rule

Agents and automation must operate only within their designated file and directory scope. Changes outside the task boundary require explicit approval.

## Rationale

Unrestricted file access leads to scope leakage, unintended side effects, and makes changes harder to review. Bounded access ensures each task is self-contained and auditable.

## Guidelines

### Scope Boundaries
- Each task defines its file scope — agents must stay within it
- Shared libraries and configuration files require explicit mention in the task plan
- Cross-service changes are not allowed without plan-level coordination

### Path Access Controls
- Production code: restricted to paths listed in the task
- Test code: restricted to corresponding test directories
- Documentation: restricted to relevant specs/ subdirectories
- Configuration: requires explicit approval for any config file changes

### Violation Detection
- Git diff reviewed against task scope at review time
- Files outside declared scope flagged as access violations
- Shared file modifications require justification in the PR description

### Exceptions
- README and documentation updates related to the change
- Package lock files updated by dependency installation
- Auto-generated files (types, schemas) triggered by in-scope changes

## Enforcement

- Reviewer agent checks diff scope against task file boundaries
- Anti-speculation skill detects scope leakage (Pattern 5)
- CI can enforce path-based ownership rules (CODEOWNERS)
