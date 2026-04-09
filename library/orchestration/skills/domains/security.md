---
kind: domain-skill
name: security
description: Security knowledge for threat modeling, trust boundaries, and defensive controls
applies_to:
  - scout
  - planner
  - builder
  - reviewer
knowledge_areas:
  - threat-modeling
  - auth-and-session-security
  - secrets-and-key-management
  - input-validation
  - encryption-and-data-protection
  - auditability
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
  - Bash
model_hint: opus
---

# Security Domain Skill

Focus on abuse cases, trust boundaries, privilege checks, secret exposure, and safe defaults.

When applying this skill:
- assume hostile inputs and curious users
- flag missing authorization and audit trails early
- review data handling for leaks at rest and in transit
- prioritize fixes that reduce blast radius and increase observability
