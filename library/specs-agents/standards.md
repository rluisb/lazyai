<rule>
  <scope>auto</scope>
  <globs>specs/standards/**</globs>
  <description>Project standards — real patterns extracted from the codebase. Load before writing code.</description>
</rule>

# Standards Directory

Standards are descriptive: they show HOW we do things here with real code references.
Rules say "validate input." Standards show "here's how we validate input in this project."

**Standards are EXTRACTED, never invented.** Every standard references real code from this codebase.

---

## Categories

```
specs/standards/
├── coding/           ← API patterns, service patterns, error handling, enums
├── architecture/     ← Module boundaries, cross-module communication, state isolation
├── testing/          ← Unit/integration/e2e patterns, fixtures, mocking
├── quality/          ← Naming conventions, file organization, review checklists
├── resilience/       ← Circuit breakers, timeouts, retries, graceful degradation
├── observability/    ← Logging, metrics, health checks, tracing
├── data/             ← Entity patterns, migrations, queries, schema design
└── security/         ← Auth patterns, input validation, secret handling
```

## Progressive Loading — Read Only What Your Task Needs

| Task | Standard to Load |
|------|-----------------|
| Creating an API endpoint | coding/api-patterns.md |
| Writing a service | coding/service-patterns.md |
| Handling errors | coding/error-handling.md |
| Creating a new module | architecture/module-boundaries.md |
| Cross-module communication | architecture/cross-module-communication.md |
| Writing unit tests | testing/unit-test-patterns.md |
| Writing integration tests | testing/integration-test-patterns.md |
| Writing e2e tests | testing/e2e-test-patterns.md |
| Naming anything | quality/naming-conventions.md |
| Organizing files | quality/file-organization.md |
| Adding external API calls | resilience/circuit-breakers.md + resilience/timeouts-retries.md |
| Adding logging | observability/logging-patterns.md |
| Adding metrics | observability/metrics-patterns.md |
| Creating entities/tables | data/entity-patterns.md |
| Writing migrations | data/migration-patterns.md |
| Handling auth | security/auth-patterns.md |
| Validating input | security/input-validation.md |

## Bootstrapping Standards (When This Directory Is Empty)

If standards don't exist yet for this project, follow this process:

### For an existing codebase:
1. **Scout agent** reads the codebase module by module
2. For each module, identify:
   - How are controllers structured? → coding/api-patterns.md
   - How are services structured? → coding/service-patterns.md
   - How are errors handled? → coding/error-handling.md
   - How are tests written? → testing/ standards
   - How are modules organized? → architecture/ standards
   - How is logging done? → observability/ standards
   - How is auth handled? → security/ standards
3. **Documenter agent** extracts patterns into standard files using `specs/templates/standard-template.md`
4. Each standard MUST reference a real file in the codebase — not invented code
5. Submit all standards as a single PR for team review

### For a new project:
1. Standards directory starts empty — that's OK
2. As code is written, patterns emerge
3. After the first 3-5 features are complete, run the bootstrap process above
4. Standards grow organically from real code

### Adding a new standard later:
1. Identify a pattern that's been used in 2+ places consistently
2. Pick the cleanest implementation as the reference
3. Create the standard file using the template
4. Submit via PR

**The rule: if a pattern exists in code but not in standards, the standard is missing — not the code.**

## Self-Improvement Triggers

<!-- When any of these events happen, check if standards need updating. -->

After ANY of these events, the agent MUST check if standards need updating:

| Event | Check |
|-------|-------|
| New pattern introduced in code | Does a standard exist for it? If not → create one. |
| Existing standard's reference file was refactored | Update the standard's example and path. |
| A PR review found a pattern violation | Is the standard clear enough? If not → improve it. |
| A new category of code appears (e.g., first WebSocket handler) | Create standards for the new category. |
| A dependency was replaced (e.g., Jest → Vitest) | Update all testing standards. |
| A tech debt item resolved a pattern issue | Update or create the relevant standard. |

When updating, append to the bottom of the standard file:

```markdown
---
**Change log:**
- [YYYY-MM-DD] [what changed and why]
```
