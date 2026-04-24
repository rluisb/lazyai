<rule>
  <scope>auto</scope>
  <globs>templates/**</globs>
  <description>Template usage standards for consistent prompt and task framing in the templates directory</description>
</rule>

# Templates Directory

Templates in this directory standardize how work requests are framed. They reduce ambiguity, improve handoffs, and keep planning and implementation consistent.

## Template Inventory

| Template | Purpose | When to Use |
|----------|---------|-------------|
| feature-prompt | Define end-to-end scope for new capabilities | New feature delivery with explicit user or business outcomes |
| bugfix-prompt | Isolate failure symptoms, causes, and validation | Reproducible defects, regressions, or production incidents |
| refactor-prompt | Improve structure without changing behavior | Complexity reduction, readability, modularity, maintainability work |
| tech-debt-prompt | Prioritize and execute deferred quality improvements | Known debt items with measurable impact on speed or reliability |
| rpi-prompt | Guide research → plan → implement execution | Multi-stage tasks requiring clear evidence and stepwise delivery |

## Usage Rules

1. Always start from the closest matching template.
2. Fill all required sections before execution begins.
3. Do not skip sections unless explicitly marked optional.
4. Replace vague language with concrete constraints and acceptance criteria.
5. Keep outputs aligned to the template structure for reviewability.

## Template Customization

- Templates are starting points, not rigid scripts.
- Adapt wording, section detail, and sequencing to team workflow.
- Preserve core intent: scope clarity, verification strategy, and handoff quality.
- If a template repeatedly mismatches real work, update it rather than bypassing it.

## Quality Checklist

- The selected template matches task intent.
- All mandatory fields are complete and specific.
- Success criteria are testable and unambiguous.
- Risks and dependencies are documented where relevant.

## Self-Improvement

When patterns evolve:
- Update the inventory table whenever templates are added, removed, or renamed.
- Introduce new templates when recurring workflows are not well covered.
- Refine section guidance based on review feedback and execution outcomes.
- Retire obsolete templates to prevent inconsistent task framing.
