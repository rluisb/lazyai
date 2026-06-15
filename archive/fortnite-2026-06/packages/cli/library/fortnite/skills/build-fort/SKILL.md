---
name: build-fort
description: UI anti-slop guardrails. Prevents AI from generating low-quality frontend code. 14 slop patterns with severity ratings. 3 tunable design dials (VISUAL_VARIANCE, MOTION_INTENSITY, INFORMATION_DENSITY). 5 UI archetypes with preset configurations.
trigger: /build-fort
triggers:
  - "UI quality check"
  - "review frontend code"
  - "frontend anti-patterns"
  - "design quality"
  - "UI code review"
skill_path: skills/build-fort
---

## Quick Reference

| | |
|---|---|
| **Use when** | Frontend code review, UI anti-pattern detection |
| **Do not use when** | Backend code (use shield-wall), research |
| **Primary agent** | wall-builder |
| **Runtime risk** | Low — advisory guardrails |
| **Outputs** | UI anti-pattern reports, design dial settings, archetype configs |
| **Validation** | Pattern coverage, archetype compliance |
| **Deep mode trigger** | `/build-fort` or frontend quality review |

# build-fort — UI Anti-Slop Guardrails

> *"A wall without mortar is just a pile of rocks. A component without guardrails is just slop waiting to collapse."*

## Purpose

The `build-fort` skill is the **frontend shield** of the Fortnite system. While `build-mode` ensures you construct against spec, and `zero-point` catches drift after the fact, `build-fort` **prevents slop at the source**.

This skill enforces UI quality standards before a single line of frontend code is written. It provides:

- **14 slop patterns** with severity ratings — know exactly what not to build
- **3 tunable design dials** — calibrate visual variance, motion intensity, and information density per project
- **5 UI archetypes** with preset configurations — dashboard, landing page, form/wizard, data table, mobile app
- **3 gate functions** — checkpoints before component creation, animation addition, and UI shipping

Use this skill when:
- Generating React/Vue/Angular/Svelte components
- Adding CSS, animations, or transitions
- Reviewing frontend PRs for quality
- Starting a new frontend feature or page
- Converting designs to code

## Slop Patterns

The following 14 patterns are **forbidden by default**. Each has a severity rating that determines enforcement strictness.

| # | Pattern | Severity | Description | Detection |
|---|---------|----------|-------------|-----------|
| 1 | **Generic Spacing** | MEDIUM | Using arbitrary `margin`/`padding` values without a design system scale (4px, 8px, 16px, 24px, 32px, 48px, 64px) | Any hardcoded pixel value not in the scale |
| 2 | **Missing States** | HIGH | Components without hover, active, disabled, or loading states | Interactive elements missing ≥2 states |
| 3 | **No Accessibility** | CRITICAL | Missing ARIA labels, roles, keyboard navigation, or screen reader support | No `aria-*`, no `role`, no focus management |
| 4 | **Poor Contrast** | CRITICAL | Text/background contrast ratio below WCAG AA (4.5:1 for normal, 3:1 for large) | Any color pair failing contrast check |
| 5 | **Animation Overload** | MEDIUM | More than 3 simultaneous animations or transitions >500ms without user preference respect | `prefers-reduced-motion` not checked |
| 6 | **No Responsive Design** | HIGH | Fixed widths, no breakpoints, or mobile-last approach | No `@media` queries or container queries |
| 7 | **Inconsistent Typography** | MEDIUM | Mixing >2 font families or using raw pixel font sizes outside the type scale | Font stack violations |
| 8 | **Missing Focus States** | HIGH | Interactive elements without visible focus indicators | No `:focus-visible` or `outline` styles |
| 9 | **Div Soup** | MEDIUM | Nesting >4 divs deep without semantic HTML (`<section>`, `<article>`, `<nav>`, etc.) | Excessive `<div>` nesting |
| 10 | **Hardcoded Strings** | LOW | UI text embedded directly in components without i18n/l10n support | Raw strings in JSX/templates |
| 11 | **No Error Boundaries** | HIGH | Components that can crash the entire UI tree without isolation | Missing error boundary wrappers |
| 12 | **Over-fetching** | MEDIUM | Loading data the component doesn't render or loading too early | Unused API calls in component scope |
| 13 | **Missing Skeleton/Loading** | MEDIUM | Async components without loading states or skeleton screens | No `<Suspense>` or skeleton fallback |
| 14 | **No Meta Tags** | LOW | Pages without `<title>`, `<meta description>`, or Open Graph tags | Missing `<Helmet>` or `<Head>` usage |

### Severity Enforcement

| Severity | Action | Override |
|------------|--------|----------|
| CRITICAL | **BLOCK** — Code cannot proceed without fix | Requires human approval + documented exception |
| HIGH | **WARN** — Must fix before merge | Can override with `// build-fort:ignore [reason]` |
| MEDIUM | **SUGGEST** — Should fix, PR comment | Can override with `// build-fort:ignore [reason]` |
| LOW | **NOTE** — Informational, no block | No override needed |

## Design Dials

Three tunable parameters control the visual character of your UI. Set these at the start of a project or per-component.

### VISUAL_VARIANCE

Controls how "loud" or "quiet" the UI is.

| Setting | Description | Use When |
|---------|-------------|----------|
| `minimal` | Flat, single accent color, generous whitespace, no shadows | Enterprise dashboards, admin panels, data-heavy tools |
| `moderate` | 2-3 accent colors, subtle depth, balanced whitespace | SaaS apps, content platforms, standard B2B |
| `expressive` | Bold colors, gradients, pronounced shadows, playful spacing | Marketing sites, creative tools, consumer apps |

**Default:** `moderate`

### MOTION_INTENSITY

Controls animation richness. Always respects `prefers-reduced-motion`.

| Setting | Description | Use When |
|---------|-------------|----------|
| `none` | Zero animations, instant state changes | Accessibility-first, performance-critical, user preference |
| `subtle` | Only opacity and transform transitions, <200ms | Professional tools, data-heavy UIs |
| `standard` | Full transitions, entrance animations, <400ms | Most web applications |
| `rich` | Complex sequences, staggered animations, micro-interactions | Marketing, onboarding, premium experiences |

**Default:** `standard`

### INFORMATION_DENSITY

Controls how much content fits per viewport.

| Setting | Description | Use When |
|---------|-------------|----------|
| `sparse` | Large padding, big typography, lots of breathing room | Landing pages, portfolios, luxury brands |
| `comfortable` | Balanced padding, readable type, clear hierarchy | Most applications, default choice |
| `dense` | Compact spacing, smaller type, more data visible | Data tables, admin panels, power-user tools |

**Default:** `comfortable`

### Dial Configuration

Set dials in your project config or per-component:

```yaml
# .build-fort.yml (project-level)
dials:
  VISUAL_VARIANCE: moderate
  MOTION_INTENSITY: standard
  INFORMATION_DENSITY: comfortable
```

```jsx
// Component-level override
// build-fort: VISUAL_VARIANCE=expressive MOTION_INTENSITY=rich
export function HeroBanner() { ... }
```

## UI Archetypes

Pre-configured dial settings for common UI patterns. Use these as starting points.

### Dashboard

**Key Concerns:** Data density, quick scanning, consistent spacing, error boundaries for widgets

```yaml
archetype: dashboard
dials:
  VISUAL_VARIANCE: minimal
  MOTION_INTENSITY: subtle
  INFORMATION_DENSITY: dense
slop_guards:
  - No Accessibility (CRITICAL)
  - Missing States (HIGH)
  - No Responsive Design (HIGH)
  - Missing Error Boundaries (HIGH)
  - Over-fetching (MEDIUM)
  - Missing Skeleton/Loading (MEDIUM)
```

### Landing Page

**Key Concerns:** Visual impact, clear CTA, performance, SEO meta tags

```yaml
archetype: landing-page
dials:
  VISUAL_VARIANCE: expressive
  MOTION_INTENSITY: rich
  INFORMATION_DENSITY: sparse
slop_guards:
  - No Accessibility (CRITICAL)
  - Poor Contrast (CRITICAL)
  - Animation Overload (MEDIUM)
  - No Responsive Design (HIGH)
  - No Meta Tags (LOW)
```

### Form / Wizard

**Key Concerns:** Focus management, validation states, progress indication, keyboard navigation

```yaml
archetype: form-wizard
dials:
  VISUAL_VARIANCE: moderate
  MOTION_INTENSITY: subtle
  INFORMATION_DENSITY: comfortable
slop_guards:
  - No Accessibility (CRITICAL)
  - Missing States (HIGH)
  - Missing Focus States (HIGH)
  - Hardcoded Strings (LOW)
  - No Error Boundaries (HIGH)
```

### Data Table

**Key Concerns:** Density, sorting/filtering states, empty states, performance

```yaml
archetype: data-table
dials:
  VISUAL_VARIANCE: minimal
  MOTION_INTENSITY: none
  INFORMATION_DENSITY: dense
slop_guards:
  - No Accessibility (CRITICAL)
  - Missing States (HIGH)
  - No Responsive Design (HIGH)
  - Missing Skeleton/Loading (MEDIUM)
  - Over-fetching (MEDIUM)
```

### Mobile App

**Key Concerns:** Touch targets, gesture support, offline states, viewport adaptation

```yaml
archetype: mobile-app
dials:
  VISUAL_VARIANCE: moderate
  MOTION_INTENSITY: standard
  INFORMATION_DENSITY: comfortable
slop_guards:
  - No Accessibility (CRITICAL)
  - Missing States (HIGH)
  - No Responsive Design (HIGH)
  - Animation Overload (MEDIUM)
  - Missing Focus States (HIGH)
```

## Gate Functions

Three mandatory checkpoints. Do not proceed until the gate is satisfied.

### Gate 1: Before Writing a Component

**Trigger:** You are about to generate a new component file.

**Checklist:**
- [ ] Which archetype does this belong to? (dashboard, landing-page, form-wizard, data-table, mobile-app)
- [ ] Are the design dials set? (VISUAL_VARIANCE, MOTION_INTENSITY, INFORMATION_DENSITY)
- [ ] What slop patterns are most relevant? (refer to archetype slop_guards)
- [ ] Is there a design system or component library to extend?
- [ ] What states does this component need? (default, hover, active, disabled, loading, error, empty)
- [ ] What accessibility requirements apply? (ARIA roles, keyboard nav, screen reader)

**Output:** Brief component contract — archetype, dials, states, a11y notes.

### Gate 2: Before Adding Animation

**Trigger:** You are about to add CSS transitions, keyframes, or JS animation libraries.

**Checklist:**
- [ ] Does the MOTION_INTENSITY dial permit this level of animation?
- [ ] Is `prefers-reduced-motion` respected?
- [ ] Are animation durations under the threshold? (subtle: 200ms, standard: 400ms, rich: 800ms)
- [ ] Will this animation run on the compositor? (transform, opacity only)
- [ ] Is there a fallback for browsers without animation support?

**Output:** Animation spec — trigger, duration, easing, reduced-motion fallback.

### Gate 3: Before Shipping UI

**Trigger:** You are about to mark a UI task as complete or open a PR.

**Checklist:**
- [ ] All CRITICAL slop patterns resolved
- [ ] All HIGH slop patterns resolved or explicitly overridden
- [ ] Responsive design verified at 320px, 768px, 1440px
- [ ] Accessibility audit: keyboard navigation, screen reader, color contrast
- [ ] Performance check: no layout thrashing, images optimized, lazy loading where appropriate
- [ ] Cross-browser check: latest Chrome, Firefox, Safari, Edge
- [ ] Meta tags present (for pages)
- [ ] Error boundaries in place (for component trees)

**Output:** Ship readiness report — pass/fail per category with evidence.

## Integration

### With build-mode

When `build-mode` is active, `build-fort` automatically:
- Reads the spec for UI requirements and maps them to archetypes
- Injects slop guards into the implementation checklist
- Validates each component against the gate functions before marking tasks complete

**Integration point:** `build-mode` task list includes `build-fort:gate-1` before component tasks and `build-fort:gate-3` before final verification.

### With zero-point

When `zero-point` runs post-implementation, `build-fort` provides:
- The slop pattern checklist for frontend-specific verification
- Contrast ratio calculations for color pairs used
- Animation audit for `prefers-reduced-motion` compliance
- Accessibility tree validation

**Integration point:** `zero-point` frontend checks call `build-fort:verify` with the component source and dial configuration.

### With shield-wall (Backend Equivalent)

`shield-wall` is the backend quality gate (API design, database schema, security). `build-fort` is its frontend counterpart.

| Concern | shield-wall | build-fort |
|---------|-------------|------------|
| Data contracts | API schema validation | PropTypes / TypeScript interfaces |
| Security | AuthZ, input sanitization | XSS prevention, CSP headers |
| Performance | Query optimization, N+1 | Bundle size, image optimization, lazy loading |
| Reliability | Circuit breakers, retries | Error boundaries, loading states |
| Monitoring | Logs, metrics, tracing | User analytics, error reporting |

**Integration point:** Full-stack features run both `shield-wall:gate` and `build-fort:gate-3` before shipping.

## Rules

### Mandatory Enforcement

1. **CRITICAL slop patterns are non-negotiable.** No code with accessibility violations or poor contrast can ship. Ever.
2. **Design dials must be explicit.** Defaulting is allowed, but the choice must be documented in the component contract.
3. **Archetypes are starting points, not prisons.** Override dials per component with inline comments.
4. **Gate functions are blocking.** Do not skip gates. Do not fake gate output.
5. **Overrides require justification.** Every `// build-fort:ignore` must include a reason and a ticket reference.
6. **Motion must respect user preference.** All animations must have a `prefers-reduced-motion: reduce` fallback.
7. **Responsive is not optional.** Every component must work at 320px minimum width.
8. **Semantic HTML is mandatory.** `<div>` is a last resort, not a default.

### Override Syntax

```jsx
// build-fort:ignore [Missing Focus States] [HIGH]
// Reason: Custom focus ring handled by parent component
// Ticket: PROJ-1234
<button className="custom-focus">Click me</button>
```

### Project Setup

Add to your project:

```bash
# .build-fort.yml in project root
dials:
  VISUAL_VARIANCE: moderate
  MOTION_INTENSITY: standard
  INFORMATION_DENSITY: comfortable

archetype: default

# Optional: custom slop patterns
slop_patterns:
  - name: "Custom Pattern"
    severity: MEDIUM
    description: "Your team-specific anti-pattern"
    detection: "regex or heuristic"
```

## Quick Reference

```
/build-fort check <file>          # Run slop detection on a file
/build-fort dial <name> <value>   # Set a design dial
/build-fort archetype <name>      # Show archetype configuration
/build-fort gate <1|2|3>          # Run a specific gate
/build-fort ignore <pattern>      # Add an override with reason
/build-fort report                # Generate ship readiness report
```

## Fortnite Lore

> *"The storm corrupts everything it touches — code, design, reason itself. The build-fort skill is your blueprint for constructing interfaces that survive the storm. Every slop pattern is a crack in the wall. Every design dial is a reinforcement. Every gate is a checkpoint before the storm hits."*

*"Build smart. Build accessible. Build fort."*
