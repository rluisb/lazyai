---
kind: domain-skill
name: frontend
description: Frontend systems knowledge for UI architecture, accessibility, and browser behavior
applies_to:
  - scout
  - planner
  - builder
  - reviewer
knowledge_areas:
  - component-architecture
  - state-management
  - accessibility
  - css-and-layout
  - browser-apis
  - performance
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
  - Bash
model_hint: sonnet
---

# Frontend Domain Skill

Focus on user flows, state transitions, rendering behavior, accessibility, responsive layout,
and browser-specific edge cases.

When applying this skill:
- preserve clear UX intent and interaction flow
- check keyboard, screen-reader, and focus behavior
- look for hydration, caching, and client/server boundary issues
- favor maintainable component seams over deeply nested logic
