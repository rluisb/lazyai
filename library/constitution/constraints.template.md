# [YOUR_PROJECT_NAME] Constraints

## Hard Constraints (Must Not Violate)
- [ ] No secrets committed to source control
- [ ] All public APIs must have tests
- [ ] No breaking changes without deprecation cycle

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
