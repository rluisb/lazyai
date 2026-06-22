# RPI Cycle 2 — Research Phase

Do not edit files until this research is complete.

## 1. Current validation paths

Inspect:

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/compiler/
packages/cli/internal/validate/
packages/cli/cmd/validate.go
packages/cli/cmd/compile.go
packages/cli/internal/schema/
```

Answer:

```text
- Where are agents validated today?
- Where are skills validated today?
- Does validation run during compile, validate, or both?
- Are warnings supported?
- Are severity levels supported?
- Are validation results machine-readable or text-only?
- Are there existing tests for validation behavior?
```

## 2. Current skill format

Inspect:

```text
packages/cli/library/skills/
packages/cli/library/manifests/curation.yaml
packages/cli/library/manifests/provenance.yaml
packages/cli/internal/library/
```

Inspect representative skills:

```text
rpi
tdd-loop
plan
implement
research
diagnose
review
chain-verify
anti-speculation
handoff
parallel-execution
memory-write
self-improve
no-workarounds
```

Answer:

```text
- Are skills Markdown-only, frontmatter-based, schema-based, or mixed?
- Which headings/patterns already exist?
- Do skills already contain triggers?
- Do they contain non-trigger or misuse guidance?
- Do they contain evidence requirements?
- Do they define expected outputs?
- Do they reference tools/fragments/agents?
- Do they follow progressive disclosure?
```

## 3. Current agent format

Inspect:

```text
packages/cli/library/canonical/agents/
packages/cli/library/canonical/agents/guide*
packages/cli/library/canonical/agents/implementer*
packages/cli/library/canonical/agents/researcher*
packages/cli/library/canonical/agents/planner*
packages/cli/library/canonical/agents/reviewer*
packages/cli/library/canonical/agents/deployer*
packages/cli/library/canonical/agents/responder*
packages/cli/library/canonical/agents/evidence-verifier*
```

Answer for each default agent:

```text
- Role/purpose
- When to use
- When not to use
- Workflow
- Evidence requirements
- Human gates
- Output format
- Handoff behavior
- Referenced skills/tools/fragments
- Gaps
```

## 4. Existing test style

Inspect:

```text
packages/cli/internal/compiler/*test.go
packages/cli/internal/validate/*test.go
packages/cli/cmd/*test.go
packages/cli/testdata/
```

Answer:

```text
- What test style should this cycle follow?
- Are fixtures preferred?
- Are golden tests relevant here?
- How are validation failures currently asserted?
```

## Research output

Before implementation, produce:

```markdown
## Research summary

- Repo state:
- Validation paths:
- Skill format:
- Agent format:
- Existing tests:
- Important risks:
- Suggested implementation approach:
```
