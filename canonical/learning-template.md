# Learning Template

Use after `task-to-issues`, `issue-triage`, `diagnose`, or any workflow that discovers reusable knowledge.

## Storage Ladder

1. Session note: `.vibe-lab/sessions/learning-YYYY-MM-DD-<slug>.md` for raw findings.
2. Memory proposal: `memory-promotion` when the finding is likely to recur.
3. Canonical rule/template: only after explicit approval or repeated evidence.

## Classification

- `rule`: stable behavior agents must always follow.
- `template`: reusable output or artifact shape.
- `trap`: sharp edge, false assumption, or recurring failure mode.
- `pattern`: repeated workflow that may deserve a skill or hook.

## Capture Block

```markdown
### <Skill> — <short descriptor>
- Trigger: <why this skill/process was used>
- Source: <issue, note, error, file, or conversation reference>
- Decision: <what changed or was classified>
- Evidence: <observed proof, not inference>
- Reusable: <yes/no and why>
- Promotion target: <none | memory | canonical rule | template | issue>
```

Do not write durable memory silently. Use `memory-promotion` for approval.
