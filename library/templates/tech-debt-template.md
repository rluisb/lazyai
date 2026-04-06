<template>
  <name>Tech Debt — Technical Debt Assessment</name>
  <output>specs/tech-debt/NNN-debt-name/techspec.md</output>
  <input>Research + codebase observation</input>
  <phase>Plan (no PRD — debt is technical, not product)</phase>
</template>

# Tech Debt: [Title]

**ID:** NNN-debt-name
**Date:** YYYY-MM-DD
**Status:** Identified | Planned | In Progress | Resolved
**Author:** [name]
**Priority:** Low | Medium | High | Critical
**Research:** [link to research.md]

---

## Debt Description

<!-- One paragraph. What the debt is. Factual, no blame. -->

[DESCRIPTION]

## Risk Assessment

<!-- What happens if we DON'T fix this. Be specific. -->

| Risk | Likelihood | Impact | Timeframe |
|------|-----------|--------|-----------|
| [what goes wrong] | Low/Med/High | Low/Med/High | [when it becomes a problem] |

## Root Cause

<!-- How this debt accumulated. No blame — just facts. -->

- [cause]

## Current Impact

<!-- How this debt affects the team TODAY. Measurable if possible. -->

- [impact]

## Proposed Fix

### Simplicity Gate
- What is the MINIMUM fix that reduces the risk? [describe]
- Can this be done incrementally? [YES → prefer incremental / NO → justify]

### Approach
[Brief description of the fix]

### Patterns to Follow

| Pattern | Standard | Reference File |
|---------|----------|---------------|
| [pattern] | [specs/standards/file.md] | [src/path/to/reference] |

### Files Affected
- [paths]

## Dependencies
- **Blocks:** [what this debt blocks — or "Nothing currently"]
- **Blocked by:** [what must happen first — or "Nothing"]

## ADR Required?
- [ ] Yes → [decision to capture]
- [ ] No

---

<!-- PRINCIPLES CHECK
- [ ] Risk is real, not theoretical
- [ ] Impact is measurable
- [ ] Fix is the simplest that reduces risk (YAGNI)
- [ ] Incremental approach preferred over big-bang
- [ ] Priority reflects actual risk, not annoyance level
- [ ] Patterns to Follow references existing standards
-->
