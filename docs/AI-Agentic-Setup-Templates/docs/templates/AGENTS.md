<rule>
  <scope>manual</scope>
  <description>Template usage rules — load only when creating documents from templates</description>
</rule>

# Template Rules

## Purpose
These templates define the structure for project documents. They are NOT examples — they are the required format.

## How to Use
1. Copy the template to the target location
2. Fill every section — delete sections only if marked "delete if not applicable"
3. Replace all `[PLACEHOLDER]` values
4. Verify the principles check at the bottom before marking as complete
5. Never modify the templates themselves — modify the output files

## Template Inventory

| Template | Used For | Output Location |
|----------|---------|----------------|
| prd-template.md | Product requirements (WHAT/WHY) | docs/features/NNN-*/prd.md |
| techspec-template.md | Technical specification (HOW) | docs/features/NNN-*/techspec.md |
| tasks-template.md | Ordered task list with phases | docs/features/NNN-*/tasks/tasks.md |
| task-template.md | Individual task files | docs/features/NNN-*/tasks/NNN-*.md |
| adr-template.md | Architecture decision records | docs/adrs/NNN-*.md |
| tech-debt-template.md | Technical debt assessment | docs/tech-debt/NNN-*/techspec.md |
| standard-template.md | Project coding standards | docs/standards/[pattern-name].md |
| progress-template.md | Feature trace log | docs/features/NNN-*/progress.md |

## Self-Improvement
When a new template is created or existing one updated:
- Update the inventory table above
- If the template adds a new document type → update docs/features/AGENTS.md workflow
- If the template changes output location → update all AGENTS.md files that reference it
