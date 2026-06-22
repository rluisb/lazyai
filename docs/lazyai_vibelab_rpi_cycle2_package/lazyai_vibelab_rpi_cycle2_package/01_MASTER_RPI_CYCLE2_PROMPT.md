# Mission: RPI Cycle 2 — Semantic validation for skills and agent contracts

You are running locally inside the LazyAI repository.

Repo context:

```text
Repo root: /Users/ricardo/projects/teachable/lazyai
Current work branch should be based on the result of RPI Cycle 1.
Product: lazyai-cli, a Go CLI harness asset manager/compiler.
LazyAI owns: .ai source, manifests, schemas, validation, adapter compilation, generated output.
vibe-lab owns: process/quality/taste assets, RPI workflow, human gates, skills, hooks, rubrics, handoff discipline.
Host tools execute generated assets.
LazyAI must not become a runtime or orchestrator.
```

Cycle 1 already completed:

```text
- Added adapter conformance fixtures and golden tests.
- Added packages/cli/testdata/.
- Added docs/concepts/harness-principles.md.
- Fixed 029 status drift.
- Confirmed no runtime/orchestration surface was added.
```

Your mission in Cycle 2 is:

> Improve LazyAI validation so skills and agents are not only structurally valid, but also behaviorally useful, scoped, safe, evidence-driven, and aligned with vibe-lab quality principles.

Use RPI:

```text
R = Research current skill/agent formats and validation paths.
P = Plan a small, safe, non-breaking implementation.
I = Implement semantic validation with tests and clear output.
```

---

## Non-negotiable boundaries

Do **not** add or reintroduce:

```text
lazyai run workflow
lazyai autonomous task
lazyai orchestrate
lazyai judge
lazyai eval runner
mandatory trace daemon
mandatory RAG core
background subagent runtime
LangChain/LangGraph/CrewAI dependency in core
old task/workflow/orchestration/eval commands
Codex adapter
```

Allowed work:

```text
- skill semantic validation
- agent contract validation
- validation warnings/errors
- schema or lint helpers
- tests and fixtures
- docs/templates explaining skill and agent quality
- small asset updates required to satisfy validation
```

Prefer warnings over breaking errors for subjective quality checks.

Only use hard errors for objective problems:

```text
- invalid schema/frontmatter
- broken references
- unsupported adapter target
- missing required structural fields
- malformed file
```

Use warnings for semantic quality issues:

```text
- weak trigger guidance
- missing non-trigger guidance
- missing evidence requirements
- missing output contract
- missing human gate guidance
- ambiguous role/scope
- no examples or anti-examples
- no progressive-disclosure guidance
```

---

# Safety rules

Before modifying files:

```bash
pwd
git rev-parse --show-toplevel
git status --short
git branch --show-current
git log -1 --oneline
```

Do not expose secret values.

Do not run destructive commands.

Never run:

```bash
rm -rf
git clean
git reset
git push
git commit
npm publish
go install
curl | sh
brew install
docker compose up
terraform apply
kubectl apply
lazyai update
lazyai eject
lazyai compile --write
lazyai init --force
```

Safe commands:

```bash
find packages/cli/library/skills -maxdepth 2 -type f | sort
find packages/cli/library/canonical/agents -maxdepth 2 -type f | sort
rg "validate|skill|agent|frontmatter|contract|evidence|trigger|when to use|handoff|human" packages/cli/internal packages/cli/library docs specs
go test ./packages/cli/internal/compiler/...
go test ./packages/cli/internal/validate/...
go test ./packages/cli/cmd/...
go test ./...
```

---

# R — Research phase

First inspect the current implementation. Do not edit yet.

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

For representative skills, inspect at least:

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

---

# P — Plan phase

After research, produce a concise implementation plan before editing.

The plan must include:

```text
1. Current skill/agent validation state
2. Proposed semantic validation checks
3. Which checks will be warnings vs errors
4. Files to change
5. Tests to add/update
6. Asset/doc updates needed
7. Compatibility risks
8. Validation commands to run
```

Keep the plan small enough to complete in this cycle.

---

# I — Implementation scope

## Priority A — Skill semantic validation

Implement conservative semantic validation for skills.

Target behavior:

```text
A high-quality skill should clearly state:
- when to use it
- when not to use it
- required evidence
- expected output or done criteria
- required tools or dependencies, if any
- human gate requirements, if any
- examples or anti-examples when useful
- progressive-disclosure behavior where relevant
```

Suggested checks:

### Error-level checks

Use errors only for objective issues:

```text
- malformed frontmatter if frontmatter exists
- invalid schema if schema exists
- broken referenced skill/fragment/tool path
- unsupported adapter target
- empty skill body
- missing title/name if current format requires it
```

### Warning-level checks

Use warnings for quality issues:

```text
- missing “when to use” / trigger guidance
- missing “when not to use” / misuse guidance
- missing evidence requirement
- missing expected output / done criteria
- missing human gate guidance for risky skills
- missing examples/anti-examples for broad skills
- too-broad name or description
- duplicated always-on rule content
- excessive always-loaded content / token-rent concern
- no progressive-disclosure cue where the skill is large or procedural
```

Important:

```text
Do not force a breaking schema migration unless the repo already has a schema model for skills.
If skills are Markdown-only, implement heading/pattern linting.
Recognize reasonable heading variants instead of requiring exact wording.
```

Acceptable heading variants:

```text
when to use:
- When to use
- Trigger
- Triggers
- Use when
- Invocation
- Activation

when not to use:
- When not to use
- Do not use
- Non-goals
- Misuse
- Avoid

evidence:
- Evidence
- Required evidence
- Verification
- Validation
- Proof
- Done criteria

output:
- Output
- Expected output
- Deliverable
- Result
- Response format
- Done
```

Potential files:

```text
packages/cli/internal/compiler/skill_validate.go
packages/cli/internal/compiler/skill_validate_test.go
packages/cli/internal/validate/
packages/cli/internal/schema/
packages/cli/library/templates/skill-quality.md
docs/concepts/skill-quality.md
```

Acceptance criteria:

```text
- Validation detects clearly bad skills.
- Existing shipped skills either pass or produce intentional, actionable warnings.
- Tests cover good skill, missing trigger, missing non-trigger, missing evidence, empty body, and broken reference if references exist.
- No unnecessary mass rewrite of all skills.
```

---

## Priority B — Agent role contract validation

Implement or improve semantic validation for canonical agents.

A high-quality agent should define:

```text
- role / purpose
- when to use
- when not to use
- expected workflow
- referenced skills/tools/fragments
- evidence requirements
- human gates
- output format
- handoff behavior
- safety boundaries
```

Default agents to validate:

```text
guide
implementer
researcher
planner
reviewer
deployer
responder
evidence-verifier
```

### Error-level checks

Use errors for objective issues:

```text
- missing/empty agent file
- malformed metadata/frontmatter if required
- invalid adapter target if declared
- broken referenced skill/fragment/tool
- missing required structural fields if the existing schema already requires them
```

### Warning-level checks

Use warnings for semantic quality:

```text
- missing role/purpose
- missing when-to-use guidance
- missing when-not-to-use guidance
- missing workflow
- missing evidence requirements
- missing human gate guidance
- missing output format
- missing handoff behavior
- ambiguous overlap with another agent
```

Potential files:

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/compiler/agent_validate_test.go
packages/cli/library/templates/agent-contract.md
docs/concepts/agent-contracts.md
```

Acceptance criteria:

```text
- Default agents have explicit contracts or actionable warnings.
- Tests cover valid agent, missing role, missing evidence, missing handoff, and broken reference if references exist.
- No runtime behavior added.
```

---

## Priority C — Validation result quality

Improve validation output so local AI agents and humans can fix issues quickly.

Each semantic validation message should include:

```text
- asset type: skill or agent
- asset path
- severity: error/warning/info
- rule id
- short message
- fix suggestion
```

Example:

```text
warning skill.missing_non_trigger packages/cli/library/skills/review.md
Skill does not explain when not to use it.
Add a “When not to use” or “Misuse” section to prevent accidental invocation.
```

Suggested rule IDs:

```text
skill.missing_trigger
skill.missing_non_trigger
skill.missing_evidence
skill.missing_output
skill.missing_progressive_disclosure
skill.too_broad
skill.empty_body
skill.broken_reference

agent.missing_role
agent.missing_trigger
agent.missing_non_trigger
agent.missing_workflow
agent.missing_evidence
agent.missing_handoff
agent.missing_output
agent.broken_reference
```

Acceptance criteria:

```text
- Validation output is actionable.
- Tests assert rule IDs.
- Existing CLI output style is preserved where possible.
```

---

## Priority D — Docs/templates for future authors

Add or update docs/templates explaining how to write good assets.

Suggested files:

```text
docs/concepts/skill-quality.md
docs/concepts/agent-contracts.md
packages/cli/library/templates/skill-quality.md
packages/cli/library/templates/agent-contract.md
```

Content should explain:

```text
- skill vs agent distinction
- progressive disclosure
- trigger and non-trigger guidance
- evidence requirements
- human gates
- output contracts
- examples and anti-examples
- how validation works
```

Acceptance criteria:

```text
- Docs are linked from harness principles if appropriate.
- Templates are included in curation/manifest if required by repo conventions.
- No duplicate or conflicting docs are created if equivalent docs already exist.
```

---

# Implementation constraints

## Be conservative

Do not try to perfect every skill/agent in this cycle.

Prefer:

```text
- validation infrastructure
- representative tests
- docs/templates
- minimal updates to shipped assets where necessary
```

Avoid:

```text
- huge rewrites of all 51 skills
- changing public asset formats unnecessarily
- strict failures for subjective quality
- breaking existing compile behavior
```

## Preserve existing architecture

Before adding new packages, check whether these already exist:

```text
internal/compiler
internal/validate
internal/schema
internal/library
```

Use existing conventions.

## Keep LazyAI as compiler

Validation is allowed.

Runtime execution is not.

This cycle should only make assets more trustworthy before host tools execute them.

---

# Required tests

Run targeted tests first:

```bash
go test -v ./packages/cli/internal/compiler -run 'Test.*Skill|Test.*Agent|Test.*Validate'
go test -v ./packages/cli/internal/validate/...
go test -v ./packages/cli/cmd/...
```

Then run broader tests if safe:

```bash
go test ./...
go vet ./...
```

If `go vet` or `go test ./...` fails for unrelated existing reasons, report the exact failure and whether your targeted tests passed.

Also check working tree:

```bash
git status --short
```

---

# Final report format

When finished, produce:

```markdown
# LazyAI / vibe-lab RPI Cycle 2 Report — Semantic Skill and Agent Validation

## 1. Research summary

- Repo state:
- Current validation paths:
- Current skill format:
- Current agent format:
- Existing tests:
- Constraints confirmed:

## 2. Plan executed

| Item | Status | Notes |
|---|---|---|

## 3. Changes made

| Area | Files changed | Summary |
|---|---|---|

## 4. Skill validation

- New checks:
- Warning rules:
- Error rules:
- Tests:
- Remaining gaps:

## 5. Agent validation

- New checks:
- Warning rules:
- Error rules:
- Tests:
- Remaining gaps:

## 6. Validation output quality

- Rule IDs added:
- Example output:
- Fix suggestions:

## 7. Docs/templates

- Added/updated:
- Linked from:
- Remaining gaps:

## 8. Tests and validation

Commands run:

```bash
...
```

Results:

```text
...
```

## 9. Changed files

```text
...
```

## 10. Risks

| Risk | Impact | Mitigation |
|---|---|---|

## 11. Remaining work

| Item | Why not completed | Suggested next step |
|---|---|---|

## 12. Product boundary confirmation

Confirm explicitly:

- No runtime/orchestration surface was added.
- No old task/workflow/eval command was reintroduced.
- No mandatory judge/scoring engine was added.
- No mandatory trace daemon was added.
- No mandatory RAG core was added.
- LazyAI remains a harness asset manager/compiler.
- vibe-lab remains the process/quality asset layer.
```

---

# Definition of done

Cycle 2 is complete when:

```text
- Skill semantic validation exists or is clearly improved.
- Agent contract validation exists or is clearly improved.
- Validation messages are actionable and include rule IDs.
- Tests cover valid and invalid skill/agent cases.
- Docs/templates explain skill and agent quality.
- Existing shipped assets are not broken.
- Targeted tests pass.
- LazyAI remains a compiler/asset manager, not a runtime.
```

Remember:

```text
Small, safe, test-backed improvements.
Prefer warnings for quality.
Use errors only for objective breakage.
Do not expand product scope.
Make local AI agents better at producing and maintaining high-quality LazyAI/vibe-lab assets.
```
