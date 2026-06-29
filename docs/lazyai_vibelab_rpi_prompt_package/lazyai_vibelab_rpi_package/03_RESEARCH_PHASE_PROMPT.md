# LazyAI / vibe-lab — RPI Research Phase Prompt

You are running locally inside the LazyAI repository.

Your job in this phase is **research only**. Do not modify files.

Use evidence from files, symbols, tests, docs, and command output. Do not guess.

## Research checklist

### 1. Compile contract

Inspect:

```text
packages/cli/internal/aimanifest/
packages/cli/internal/lockfile/
packages/cli/internal/writer/
packages/cli/internal/compiler/
packages/cli/cmd/compile.go
packages/cli/internal/schema/
```

Answer:

```text
- How compile selects targets
- How lockfile works
- How managed-region drift is detected
- How dry-run and force behave
- Where contract validation lives
- Where agent validation lives
```

### 2. Adapter support

Inspect:

```text
packages/cli/internal/adapter/
```

Confirm all 8 targets:

```text
opencode
claude
copilot
pi
omp
kiro
antigravity
codex
```

For each adapter, identify:

```text
- output paths
- MCP behavior
- hook behavior
- skill behavior
- agent behavior
- tests
- beta/stable status
```

Special attention:

```text
- OMP beta
- Antigravity beta
- Pi MCP no-op
- Codex rejection
```

### 3. Test/conformance state

Inspect:

```text
packages/cli/testdata/
packages/cli/internal/adapter/*test*
packages/cli/internal/compiler/*test*
packages/cli/cmd/*test*
```

Answer:

```text
- Does packages/cli/testdata exist?
- Are adapter fixtures inline or file-based?
- Are golden outputs present?
- Which adapter behaviors are currently regression-tested?
- Which behaviors are missing fixture coverage?
```

### 4. Docs and spec status drift

Inspect:

```text
README.md
packages/cli/KNOWLEDGE_MAP.md
specs/029-lazyai-v2/spec.md
specs/adrs/
docs/concepts/
docs/lazyai-vibelab-product-spec-pack/
```

Answer:

```text
- Which feature/status source is authoritative?
- Is 029 marked Draft or Done?
- Are README, spec, ADRs, and knowledge map consistent?
- Are retired features still presented as active?
```

### 5. Embedded library assets

Inspect:

```text
packages/cli/library/
packages/cli/library/manifests/curation.yaml
packages/cli/library/manifests/provenance.yaml
packages/cli/internal/library/
```

Pay attention to:

```text
skills/
canonical/agents/
rules/
fragments/
prompts/
hooks/
templates/
standards/
mcp/catalog.json
opencode/
claudecode/
copilot/
pi/
omp/
antigravity/
```

Answer:

```text
- Which asset categories exist?
- How many active/excluded assets are declared?
- Which assets represent vibe-lab quality/process?
- Which assets are compiled to which surfaces?
```

### 6. Skills and agent validation

Inspect:

```text
packages/cli/internal/compiler/agent_validate.go
packages/cli/internal/validate/
packages/cli/library/skills/
packages/cli/library/canonical/agents/
```

Answer:

```text
- What skill validation exists today?
- What agent validation exists today?
- Are triggers validated?
- Are non-triggers/misuse guidance validated?
- Are evidence requirements validated?
- Is progressive disclosure validated?
- Are missing references detected?
- Are broad/ambiguous skills warned?
```

### 7. Hooks / middleware / policy

Inspect:

```text
packages/cli/library/hooks/
packages/cli/library/claudecode/hooks/
packages/cli/library/opencode/
packages/cli/library/antigravity/
packages/cli/internal/adapter/
packages/cli/library/cupcake/
```

Answer:

```text
- What hook assets exist?
- Are hooks mapped to a neutral lifecycle?
- Which adapters support actual hooks?
- Which adapters are instruction-only?
- Is human approval modeled?
- Is destructive command blocking modeled?
```

### 8. Trace/eval/rubric state

Inspect:

```text
packages/cli/internal/evals/
packages/cli/internal/schema/
packages/cli/library/templates/
packages/cli/library/skills/
packages/cli/library/rules/
packages/cli/internal/runtime/
packages/cli/internal/db/
```

Answer:

```text
- Are eval cases supported?
- Are holdouts supported?
- Are rubrics supported?
- Is there a trace taxonomy?
- Is there a harness-change report template?
- Is there LLM-as-judge behavior?
- Are session/ledger/memory/metrics/cost optional?
```

## Required research output

Produce this report:

```markdown
# LazyAI / vibe-lab Research Report

## Repo state

- Root:
- Branch:
- HEAD:
- Dirty state:

## Existing implementation confirmed

## Missing pieces confirmed

## Important constraints

## Evidence table

| Finding | Evidence |
|---|---|

## Recommended implementation plan

Do not implement yet. Wait for confirmation or continue only if the original instruction explicitly asked for implementation.
```
