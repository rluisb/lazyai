---
name: investigate
description: Open-ended investigation for when you need to understand something broadly before knowing the specific question.
argument-hint: "[topic-or-domain]"
trigger: /investigate
phase: research
techniques: [chain-of-thought, breadth-first]
output: .specify/investigate/{YYYY-MM-DD}-{topic-slug}.md
output_schema:
  sections:
    - Investigation Topic (what we're exploring)
    - Initial Hypotheses (what we think we might find)
    - Sources Consulted (docs, code, web, experiments)
    - Key Findings (discovery log)
    - Remaining Questions (what we still don't know)
    - Confidence Assessment (how certain are we?)
    - Next Steps (what to do with this knowledge)
consumes:
  - broad topic description
  - optional: specific files or docs to examine
produces_for:
  - research-rpi (if bounded question emerges)
  - memory (if significant discoveries)
  - plan (if investigation clarifies scope)
mcp_tools: [filesystem, ripgrep, qmd, web-search]
harness:
  feed_forward: [topic description]
  contract: [findings-documented, questions-identified]
  sensors: []
  memory: [ledger.md for significant discoveries]
  anti_slope: [no-rabbit-holes, time-boxed]
workspace:
  scope: [project, workspace]
  reads: [broad codebase, documentation, web]
  writes: [.specify/investigate/{YYYY-MM-DD}-{topic-slug}.md]
  cross_repo: true
---

# Investigate Skill

## When to Use
- Not a feature research task — broader exploration
- Don't know the specific question yet ("understand the auth system")
- Need to map unfamiliar territory before bounded research
- Technical deep-dive on an unfamiliar subsystem

## When NOT to Use
- Feature RPI workflow → use `research-rpi` instead
- Bug investigation → use `diagnose` skill
- Quick code lookup → use grep/read directly

## Workflow
1. **Frame the topic** — what broad area do we need to understand?
2. **Form hypotheses** — what do we think we'll find?
3. **Explore breadth-first** — scan docs, read key files, check patterns
4. **Drill on interesting findings** — follow promising leads
5. **Document** — write up findings, remaining questions, confidence

## Time Box
- Default: 30-60 minutes
- If more time needed, document progress and escalate

## Integration
- Agent: Scout (or any agent needing broad understanding)
- Feeds: `research-rpi` (if bounded question emerges), `plan`, `memory`
- Output: `.specify/investigate/{YYYY-MM-DD}-{topic-slug}.md`
- Duration: Time-boxed (typically 30-60 min); if more needed, escalate