---
kind: domain-skill
name: backend
description: Backend systems knowledge for APIs, persistence, queues, and service boundaries
applies_to:
  - scout
  - planner
  - builder
  - reviewer
knowledge_areas:
  - api-design
  - auth-and-authorization
  - database-access
  - caching
  - background-jobs
  - observability
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
  - Bash
model_hint: sonnet
---

# Backend Domain Skill

Focus on service boundaries, data ownership, transport contracts, persistence semantics,
failure modes, idempotency, and operational safety.

When applying this skill:
- trace request flow end to end
- identify schema, tenancy, and migration impact early
- prefer explicit contracts over hidden coupling
- surface performance, reliability, and rollback risks
