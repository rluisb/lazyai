---
name: research
description: Research codebase patterns, dependencies, and conventions.
argument-hint: "[topic-or-question]"
trigger: /research
phase: research
---

# Research Skill

## Workflow
1. Clarify scope — what exactly are we researching and why?
2. Search codebase — identify relevant files, patterns, dependencies
3. Read documentation — check specs/standards/, specs/adrs/, KNOWLEDGE_MAP.md
4. Identify patterns — how does the codebase handle similar things?
5. Assess impact — what could be affected by changes in this area?
6. Produce findings — write structured research output

## Trace Protocol (complex research only)
1. **Thought**: what am I looking for next?
2. **Action**: file read / search / grep
3. **Observation**: what I found
4. **Decision**: continue searching / enough context

## Integration
- Agent: Scout
- Output feeds into: `plan` skill
- Output location: specs/features/NNN-name/research.md
