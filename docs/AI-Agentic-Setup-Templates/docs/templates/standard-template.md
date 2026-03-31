<template>
  <name>Standard — Project Standard</name>
  <output>docs/standards/[category]/[pattern-name].md</output>
  <input>Extracted from existing codebase patterns — never invented</input>
  <phase>Any — created when a pattern is identified in the codebase</phase>
</template>

<!-- ============================================================
  HOW TO USE THIS TEMPLATE

  1. Pick your concern category (coding/architecture/testing/quality/
     resilience/observability/data/security)
  2. Keep ALL base sections (they are always required)
  3. Add the concern-specific sections that apply to your category
  4. Delete concern sections that don't apply
  5. Fill with REAL code from the codebase — never invent examples
  6. Submit via PR for team review
============================================================ -->

<rule>
  <scope>auto</scope>
  <globs>[GLOB_PATTERNS_WHERE_THIS_APPLIES]</globs>
  <description>[ONE_LINE — when the AI should load this standard]</description>
</rule>

# Standard: [Pattern Name]

**Category:** [coding | architecture | testing | quality | resilience | observability | data | security]
**Purpose:** [One sentence — what pattern this documents.]
**Reference:** `[path/to/real/file/in/codebase.ext]`
**Extracted from:** [commit, PR, or "initial codebase analysis"]

---

<!-- ============================================================
  BASE SECTIONS — Always present in every standard
============================================================ -->

## Rules

<!-- Imperative. MUST/NEVER language. ✅ what to do, ❌ what not to do. -->

- ✅ [do this — with brief reason]
- ✅ [do this]
- ❌ [never this — with consequence]
- ❌ [never this]

## Example

<!-- REAL code from the codebase. Under 30 lines.
     Always show ✅ GOOD vs ❌ BAD when possible.
     Include file path and approximate line numbers. -->

<!-- From [path/to/file.ext] — lines NN-NN -->

```[language]
// ✅ GOOD
[real code from codebase]
```

```[language]
// ❌ BAD
[anti-pattern — real or realistic]
```

## Anti-Patterns

- ❌ [bad practice] — [why it's wrong or what to do instead]
- ❌ [bad practice] — [why]

## When to Apply

- [condition — e.g. "Creating a new API endpoint"]
- [condition — e.g. "Modifying an existing service"]

<!-- ============================================================
  CONCERN-SPECIFIC SECTIONS — Include only what applies.
  Delete entire sections that don't match your category.
============================================================ -->

<!-- ===================== ARCHITECTURE ===================== -->
<!-- Include when: module boundaries, composition, communication patterns -->

## Violation Severity

<!-- 🔴 CRITICAL = data corruption, security hole
     🟠 HIGH = architecture degradation
     🟡 MEDIUM = code quality -->

**Severity:** 🔴 CRITICAL | 🟠 HIGH | 🟡 MEDIUM

## Diagram

```
[ASCII component or boundary diagram]
```

## Detection

<!-- Bash commands to find violations. Run these to verify compliance. -->

```bash
[command to find violations in the codebase]
```

**Expected output:** [empty = no violations, or describe what good looks like]

<!-- ===================== TESTING ===================== -->
<!-- Include when: test structure, fixtures, mocking, lifecycle -->

## File Location

```
[path pattern for test files — e.g. src/modules/{module}/__tests__/{name}.test.ts]
```

## Setup Template

<!-- Boilerplate that every test of this type needs. -->

```[language]
[setup/lifecycle code]
```

<!-- ===================== QUALITY ===================== -->
<!-- Include when: naming, file organization, checklists -->

## Naming Table

| Category | Convention | Example |
|----------|-----------|---------|
| [what] | [rule] | [example] |

## Checklist

- [ ] [verification item]
- [ ] [verification item]

<!-- ===================== RESILIENCE ===================== -->
<!-- Include when: failure handling, circuit breakers, retries, degradation -->

## Failure Scenario

<!-- What goes wrong WITHOUT this pattern. Be specific. -->

[SCENARIO]

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| [setting] | [value] | [reason] |

<!-- ===================== OBSERVABILITY ===================== -->
<!-- Include when: logging, metrics, health checks, tracing -->

## Log Format

```
[structured log example with fields]
```

## Metric Naming

Pattern: `[prefix]_[module]_[operation]_[unit]`
Example: `billing_subscription_creates_total`

## What to Log / Not Log

- ✅ Log: [what]
- ❌ Never log: [what — e.g. PII, secrets]

<!-- ===================== DATA ===================== -->
<!-- Include when: entities, migrations, queries, schemas -->

## Schema

```sql
[table/entity definition or TypeORM entity example]
```

## Migration Checklist

- [ ] [step — e.g. "Migration is reversible"]
- [ ] [step]

<!-- ===================== SECURITY ===================== -->
<!-- Include when: auth, validation, secrets, threat prevention -->

## Threat

<!-- What attack or risk this standard prevents. -->

[THREAT_DESCRIPTION]

## Validation Rule

```[language]
[validation/sanitization code]
```

---

<!-- STANDARD RULES
- One pattern per file. Do one thing well.
- Reference REAL files. Not hypothetical code.
- Example under 30 lines. Show the shape, not the whole file.
- If referenced file changes → update this standard.
- New standards go through PR.
- If violated repeatedly → promote key rule to docs/rules/.
- If a detection command exists → include it. Verifiable > aspirational.
-->
