---
kind: domain-skill
name: data
description: Data systems knowledge for schemas, analytics, ETL, and pipeline correctness
applies_to:
  - scout
  - planner
  - builder
  - reviewer
knowledge_areas:
  - schema-design
  - sql-and-querying
  - migrations
  - etl-pipelines
  - data-quality
  - analytics
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
  - Bash
model_hint: sonnet
---

# Data Domain Skill

Focus on data modeling, transformation correctness, lineage, migrations, and the operational
impact of changing schemas or queries.

When applying this skill:
- identify source-of-truth ownership and downstream consumers
- protect data quality, backfill safety, and reversibility
- check aggregation and null-handling assumptions carefully
- surface cost and performance risks for large datasets
