# LazyAI / vibe-lab Research Report

## Repo state

- Root: `/Users/ricardo/projects/teachable/lazyai`
- Branch: `rpi-task`
- HEAD: `f144afd Merge pull request #313 from rluisb/docs/library-reference`
- Dirty state: `docs/lazyai_vibelab_rpi_prompt_package/` and `tmp/` are untracked; working tree is otherwise clean.

## Existing implementation confirmed

1. **Compile Contract:** `lazyai.json` targets are parsed in `aimanifest.go` using `ResolveTargets()`. The lockfile tracks hashes for output idempotency. The writer detects drift and refuses overwrites without `--force` (e.g. `writer.go`, `TestDriftRefusedWithoutForce`). `DryRun` successfully skips disk writes (`TestDryRunWritesNothing`).
2. **Adapter Support:** 7 target adapters are present (`opencode`, `claude`, `copilot`, `pi`, `omp`, `antigravity`, `kiro`). `packages/cli/internal/adapter/capabilities.go` defines supported surfaces for each adapter (e.g., hooks, skills, MCP, plugins).
3. **Embedded Library Assets:** Assets exist under `packages/cli/library/`, with `curation.yaml` tracking inclusions/exclusions and ensuring `vibe-lab` assets are present (e.g. baseline-facing canonical agents and workflow skills).
4. **Validation:** Agent validation (`ValidateAgentResolutions`) checks against tool capabilities. Contract validation (`ValidateChain`) uses TypeScript-compatible semantics to detect duplicate schemas, missing downstream contracts, and missing producers.

## Missing pieces confirmed

1. **Adapter Conformance Fixtures:** `packages/cli/testdata/` is entirely missing. Golden outputs are not established; existing adapter tests rely on inline string assertions (`fstest.MapFS`) instead of file-based fixtures.
2. **Status Drift:** `packages/cli/KNOWLEDGE_MAP.md` claims `specs/029-lazyai-v2/` is `Done`, but `specs/029-lazyai-v2/spec.md` declares itself as `Draft`.
3. **Semantic Skill Validation:** Current skill validation is mostly structural; there are no linter checks enforcing evidence requirements, misuse prevention, or triggering guidance within skill markdown.
4. **Agent Role Contract Validation:** Agents are mostly checked for resolving models or valid fields, not for semantic role boundaries or progressive-disclosure instructions.

## Important constraints

- LazyAI must strictly remain a **compiler** and **asset manager** for harness surfaces (i.e. generating `.ai/` into tool-native outputs).
- No new orchestration engines, default RAG cores, or execution runtimes may be implemented.
- Any new skill or agent validation must avoid breaking schema changes if Markdown format suffices.

## Evidence table

| Finding | Evidence |
|---|---|
| Target selection | `packages/cli/internal/aimanifest/aimanifest.go` handles target parsing |
| Drift detection | `packages/cli/internal/writer/writer.go` and `writer_test.go` (`TestDriftRefusedWithoutForce`) |
| Adapter capabilities | `packages/cli/internal/adapter/capabilities.go` maps surfaces across 7 targets |
| Asset presence | `packages/cli/library/manifests/curation.yaml` records assets like skills, hooks, agents |
| Missing testdata | `ls packages/cli/testdata` fails with `No such file or directory` |
| Status drift | `packages/cli/KNOWLEDGE_MAP.md` (`Done`) vs `specs/029-lazyai-v2/spec.md` (`Draft`) |
| Contract validation | `packages/cli/internal/compiler/contract_validator.go` validates dependencies between skills/agents |

## Recommended implementation plan

1. **Priority 1: Adapter Conformance Fixtures & Golden Tests**
   - Create `packages/cli/testdata/` structure (projects and golden directories).
   - Write file-based tests covering minimal manifest, 7-target manifest, drift conflicts, and adapter-specific exceptions (e.g., Pi MCP no-op, Codex rejection).
2. **Priority 2: Consolidate harness principles**
   - Consolidate principles into `docs/concepts/harness-principles.md` or equivalent, defining the boundary between LazyAI (compile) and vibe-lab (quality/taste).
3. **Priority 3: Fix product/spec status drift**
   - Resolve the discrepancy in `specs/029-lazyai-v2/` and update `KNOWLEDGE_MAP.md` and related documents accurately.
4. **Priority 4: Semantic Skill Validation**
   - Enhance `skill_validate.go` to emit warnings for missing semantic elements in Markdown skills (trigger guidance, misuse guidance, required tools).
5. **Priority 5: Agent Role Contract Validation**
   - Enhance `agent_validate.go` to check agent capabilities, workflow contracts, and handoff instructions without introducing breaking schemas.

Do not implement yet. Wait for confirmation or continue only if the original instruction explicitly asked for implementation.