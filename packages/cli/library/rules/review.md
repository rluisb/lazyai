# Rule: Code Review

**Category:** Process
**Status:** Active

---

## Rule

All code changes require review before merging. Reviews must verify both quality AND spec compliance.

## Rationale

Reviews catch bugs early, share knowledge, enforce quality standards, and ensure implementation matches what was specified — not more, not less.

## Review Dimensions

### 1. Spec Traceability
- Every changed file maps to a task in the plan
- Every new behavior maps to a spec requirement
- No behaviors added that aren't in the spec (anti-speculation check)
- No spec behaviors missing from the implementation

### 2. Correctness
- Does the code do what the spec says it should do?
- Are edge cases handled as specified?
- Are error conditions handled appropriately?

### 3. Test Coverage
- Does every spec behavior have a corresponding test?
- Do tests verify behavior, not implementation details?
- Are test files committed before or with production code?

### 4. Maintainability
- Is the code easy to read and understand?
- Does it follow existing project conventions?
- Are there no unnecessary abstractions or premature generalizations?

### 5. Security
- Are there input validation gaps?
- Are authentication and authorization properly enforced?
- Is sensitive data handled securely?

### 6. Performance
- Are there obvious performance bottlenecks?
- Are database queries efficient?
- Is there unnecessary computation or duplication?

## Detection Heuristics

During review, check:

- [ ] Every new public function/method maps to a spec behavior
- [ ] Every new endpoint maps to a spec requirement
- [ ] No abstract base classes for single implementations
- [ ] Diff stays within the scope of the task
- [ ] Test files exist for all new behaviors
- [ ] No TODO/FIXME without a corresponding ticket

## Enforcement

- PR review checklist includes all 6 dimensions
- CI gates for tests, linting, and type checking
- Reviewer agent runs detection heuristics automatically
- Anti-speculation skill's Halt Protocol invoked on scope violations
