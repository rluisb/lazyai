# [YOUR_PROJECT_NAME] Constraints

## Hard Constraints (Must Not Violate)
- [ ] No secrets committed to source control
- [ ] All public APIs must have tests
- [ ] No breaking changes without deprecation cycle

<!-- Example hard constraints:
- [ ] All database queries must be parameterized (no string concatenation)
- [ ] Authentication changes require security review
- [ ] No direct DOM manipulation outside React components
- [ ] API response schemas must be validated with zod/joi
-->

## Soft Constraints (Prefer Unless Justified)
- [ ] Prefer composition over inheritance
- [ ] Keep functions under 50 lines
- [ ] Minimize dependencies

## Technology Boundaries
- Approved languages: [list]
- Approved frameworks: [list]
- Forbidden patterns: [list]

## Constraint Techniques
1. **Boundary Fence**: Define what's in/out of scope
2. **Dependency Lock**: Pin approved versions
3. **Style Rail**: Enforce via linter rules
4. **Safety Net**: Required test coverage thresholds
5. **Review Gate**: Human approval for certain changes

<!-- These techniques from Spec-Driven Development help LLMs stay within bounds.
Use them in specifications and task descriptions to prevent scope drift. -->
