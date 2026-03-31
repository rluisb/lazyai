<rule>
  <scope>auto</scope>
  <globs>docs/rules/**</globs>
  <description>Project rules — prescriptive conventions the AI must follow</description>
</rule>

# Rules Directory

Rules are prescriptive: they tell the AI WHAT to do and what NOT to do.
They exist independent of any AI tool — these are team conventions.

## Loading Rules

| Rule File | When to Load |
|-----------|-------------|
| code-style.md | Writing or modifying code |
| testing.md | Writing or modifying tests |
| workflow.md | Starting any task (always relevant) |
| security.md | Touching auth, secrets, user input, or APIs |
| access.md | Reading or writing any file (path checks) |
| review.md | Reviewing code or preparing for review |
| cost.md | Choosing models or managing sessions |

## Maintenance
- Rules change via PR. Reviewed by team.
- If the AI does something wrong twice → write a rule.
- Keep each file under 100 lines. Split if longer.

## Self-Improvement
When a new rule is added or an existing one changes:
- Update the loading table above if a new rule file was created
- Update root AGENTS.md if the rule affects the decision tree
- If the rule originated from a docs/memory/ note → delete the memory note after promotion
