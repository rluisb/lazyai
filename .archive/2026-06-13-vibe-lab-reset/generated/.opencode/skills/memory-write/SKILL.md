---
name: memory-write
description: Write persistent context and decisions for future sessions.
trigger: /memory-write
phase: post-task
techniques: [chain-of-thought, self-consistency]
output: .specify/memory/repos/{repo}/(entities|lessons|decisions)/  OR specs/memory/
output_schema:
  sections:
    - Memory Entry (entity | lesson | decision)
    - Confidence Level (HIGH | MEDIUM | LOW + justification)
    - Origin (what triggered this? ADR/incident/spike/code)
    - Relationships (knows_about, depends_on, prevents, enables, contradicts)
    - Expiration (when to revisit)
    - Related (links to standards, ADRs, code)
consumes:
  - decisions, learnings, entity discoveries
  - library/templates/memory-entity.md, memory-lesson.md (examples)
produces_for:
  - future workflows (searchable, context-aware)
  - constitution amendments (if memory contradicts assumptions)
mcp_tools: [filesystem, qmd]
harness:
  feed_forward: [decisions/learnings from upstream phases]
  contract: [memory-searchable]
  sensors: [memory-consistency]
  memory: [ledger.md]
  anti_slope: [no-silent-discoveries, confidence-leveled]
workspace:
  scope: [project, workspace]
  reads: [decisions, learnings, discoveries]
  writes: [.specify/memory/repos/{repo}/(entities|lessons)/ OR specs/memory/]
  cross_repo: true
---
## Quick Reference

| | |
|---|---|
| **Use when** | [When to use this skill] |
| **Do not use when** | [When NOT to use this skill] |
| **Primary agent** | [Which agent uses this] |
| **Runtime risk** | [Low/Medium/High] |
| **Outputs** | [What this skill produces] |
| **Validation** | [How to validate output] |
| **Deep mode trigger** | [How to trigger full mode] |



# Memory Write Skill

## When to Write Memory
- A non-obvious gotcha was discovered during implementation
- A pattern decision was made that future sessions need to know
- A workaround was applied that should be documented
- A lesson was learned that prevents repeating a mistake
- An entity (concept, component, tool) was created that should be tracked

## YAGNI Gate
Before writing, ask: "Will this help a future session? Or am I just documenting for the sake of it?"
If the insight is obvious from the code itself, skip the memory file.
**Max 30 lines** per memory entry — if you can't summarize it in 30 lines, the concept is too broad.

## Memory Types

### Entity
A concept or component with properties and relationships.
- Stored: `.specify/memory/repos/{repo}/entities/{EntityName}.md`
- Naming: CamelCase (PhotoTag, UserService)
- Confidence: Required (HIGH = tested/proven, MEDIUM = observed, LOW = assumption)

### Lesson
An insight or surprise from spike, incident, or implementation.
- Stored: `.specify/memory/repos/{repo}/lessons/{YYYY-MM-DD}-{lesson-slug}.md`
- Naming: slug format (concurrent-writes-behavior, cache-invalidation-tricky)
- Confidence: Required

### Decision
A choice made with rationale.
- Stored: `.specify/memory/repos/{repo}/decisions/{YYYY-MM-DD}-{decision-slug}.md`
- Links to: ADR if one exists

## Workflow
1. Identify the insight — one concept per memory file
2. Classify: entity / lesson / decision
3. Assess confidence: HIGH / MEDIUM / LOW (with justification)
4. Identify relationships: knows_about, depends_on, prevents, enables, contradicts
5. Cite origin: ADR, incident, spike, or code commit
6. Set expiration: when should this be re-evaluated?
7. Write to appropriate directory (see output_schema)
8. Update ledger: date, who, what, why, origin
9. Check promotion — if pattern repeats 2+ times, promote to specs/standards/

## Promotion Path
- Pattern repeated 2+ times → promote to specs/standards/ → delete memory
- Rule needed to prevent issue → promote to specs/rules/ → delete memory
- Still advisory after expiration → re-evaluate: keep or delete

## Ledger Format
```
| {date} | {agent} | {type} created: {name} | {confidence} | {origin} |
```
Ledger location: `.specify/memory/repos/{repo}/ledger.md` (append-only)

## Confidence Levels
- **HIGH**: Tested in production or spike, proven over time. No contradiction with constitution.
- **MEDIUM**: Observed multiple times, recurring pattern, not yet formally tested.
- **LOW**: Assumption, single observation, unverified. Must set short expiration.

## Relationships Model
- **knows_about**: Entity A knows about entity B
- **depends_on**: Entity A depends on entity B
- **prevents**: Entity A prevents problem B
- **enables**: Entity A enables capability B
- **contradicts**: Entity A contradicts assumption B (escalate if critical)

## Integration
- Agent: any (typically Builder or Documenter)
- Triggered: after completing a task or discovering a non-obvious insight
- Feeds: future research, plan, implement phases
- Handoff: ledger.md is append-only, searched by memory-read skills