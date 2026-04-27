---
kind: domain-skill
name: devops
description: DevOps knowledge for CI/CD, infrastructure, deployment, and runtime operations
applies_to:
  - scout
  - planner
  - builder
  - reviewer
knowledge_areas:
  - ci-cd
  - infrastructure-as-code
  - containers
  - deployment-strategies
  - monitoring-and-alerting
  - secrets-management
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
  - Bash
model_hint: sonnet
---

# DevOps Domain Skill

Focus on build pipelines, release safety, infrastructure drift, configuration management,
runtime observability, and recovery paths.

When applying this skill:
- verify deployment assumptions and rollback options
- flag secret handling and environment drift risks
- prefer repeatable automation over manual operational steps
- highlight monitoring gaps that would hinder incident response
