<rule>
  <scope>auto</scope>
  <description>Documentation structure and standards for this project</description>
</rule>

# Documentation Rules

## Structure

| Directory | Contains | Purpose |
|-----------|----------|---------|
| docs/rules/ | Prescriptive rules (WHAT to do) | Team conventions, enforced by AI |
| docs/standards/ | Descriptive patterns (HOW we do it) | Real code references, examples |
| docs/templates/ | Document templates | Formats for PRD, TechSpec, Tasks, ADR |
| docs/adrs/ | Architecture Decision Records | Permanent decisions (never edit, only supersede) |
| docs/features/ | Feature work artifacts | RPI flow: research → PRD → techspec → tasks |
| docs/bugfixes/ | Bugfix work artifacts | Shortened flow: research → techspec → tasks |
| docs/refactors/ | Refactor work artifacts | Full flow + mandatory ADR |
| docs/tech-debt/ | Tech debt work artifacts | Shortened flow: research → techspec → tasks (no PRD) |
| docs/memory/ | Agent learnings | Patterns discovered during work |

## Rules
- All docs are markdown. No other formats.
- Keep files focused — one topic per file.
- Update KNOWLEDGE_MAP.md when creating new features, bugfixes, or ADRs.
- Never modify accepted ADRs — create new ones that supersede.

## Self-Improvement
Any change to documentation structure (new directories, renamed files, new categories) MUST be reflected in:
1. This file's structure table (above)
2. Root AGENTS.md decision tree
3. docs/KNOWLEDGE_MAP.md
