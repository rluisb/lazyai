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
| plan-template.md | Main plan document for features and refactors | specs/features/NNN-*/plan.md |
| spec-template.md | Detailed specification for complex changes | specs/features/NNN-*/spec.md |
| checklist-template.md | Verification criteria and release checks | specs/features/NNN-*/checklists/requirements.md |
| task.md | Individual task files | specs/features/NNN-*/tasks/NNN-*.md |
| adr.md | Architecture decision records | specs/adrs/NNN-*.md |
| tech-debt-template.md | Technical debt assessment | specs/tech-debt/NNN-*/techspec.md |
| bugfix-rca-template.md | Bugfix root cause analysis | specs/bugfixes/NNN-*/techspec.md |
| standard.md | Project standards | specs/standards/[pattern-name].md |
| code-review-template.md | Structured review notes | specs/reviews/[topic].md |
| postmortem-template.md | Incident or failure retrospectives | specs/postmortems/[topic].md |

## Self-Improvement
When a new template is created or existing one updated:
- Update the inventory table above
- If the template adds a new document type → update specs workflow guidance
- If the template changes output location → update all AGENTS.md files that reference it
