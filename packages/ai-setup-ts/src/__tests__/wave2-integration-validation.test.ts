import { existsSync, readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { findMonorepoLibraryDir } from './test-helpers.js'

const wave2AcceptanceCriteria = [
  'AC-D4-001',
  'AC-D4-002',
  'AC-D4-003',
  'AC-D4-004',
  'AC-N12-001',
  'AC-N12-002',
  'AC-N12-003',
  'AC-N2-001',
  'AC-D14-001',
  'AC-D14-002',
  'AC-D9-001',
  'AC-D9-002',
  'AC-D9-003',
  'AC-D13-001',
  'AC-D13-002',
  'AC-D5-001',
  'AC-D5-002',
  'AC-D5-003',
  'AC-D8-001',
  'AC-D8-002',
  'AC-D8-003',
]

describe('Wave 2 integration validation and scope audit', () => {
  const libraryDir = findMonorepoLibraryDir()
  const repoRoot = findRepoRoot(libraryDir)

  it('records a T024 acceptance trace for every Wave 2 acceptance criterion', () => {
    const t024 = readRepoFile(repoRoot, 'specs/features/001-ai-techniques-integration/tasks/024-wave2-integration-validation.md')

    expect(t024).toMatch(/Status:\*\* DONE/)
    expect(t024).toContain('## Acceptance Trace Evidence')
    for (const ac of wave2AcceptanceCriteria) {
      expect(t024).toContain(ac)
    }
    expect(t024).toContain('## Scope Audit')
    expect(t024).toContain('## Validation Evidence')
  })

  it('keeps feature chains sequential without unsupported runtime constructs', () => {
    for (const chainFile of ['feature.json', 'feature-adversarial.json']) {
      const raw = readFileSync(path.join(libraryDir, 'orchestration', 'chains', chainFile), 'utf8')
      const chain = JSON.parse(raw) as { parallel?: unknown; steps?: Array<Record<string, unknown>> }

      expect(raw).not.toContain('{{#if')
      expect(raw).not.toContain('{{/if}}')
      expect(raw).not.toContain('optionalByFeature')
      expect(raw).not.toContain('condition')
      expect(chain).not.toHaveProperty('parallel')
      expect(Array.isArray(chain.steps)).toBe(true)
      for (const step of chain.steps ?? []) {
        expect(step).not.toHaveProperty('parallel')
        expect(step).not.toHaveProperty('condition')
        expect(step).not.toHaveProperty('optionalByFeature')
      }
    }
  })

  it('does not add unapproved Wave 3/4 or D5/D8 runtime artifacts', () => {
    const forbiddenArtifacts = [
      'packages/orchestrator/src/recovery-classifier.ts',
      'packages/orchestrator/src/model-router.ts',
      'packages/orchestrator/src/telemetry.ts',
      'packages/orchestrator/src/rag.ts',
      'packages/orchestrator/src/lifecycle-state.ts',
      'packages/ai-setup-go/library/rules/rag.md',
      'packages/ai-setup-go/library/rules/model-routing.md',
      'packages/ai-setup-go/library/rules/learning.md',
      'packages/ai-setup-go/library/orchestration/workflows/debate.json',
    ]

    for (const artifact of forbiddenArtifacts) {
      expect(existsSync(path.join(repoRoot, artifact))).toBe(false)
    }

    const runtimeTypes = readRepoFile(repoRoot, 'packages/orchestrator/src/types.ts')
    for (const lifecycleLabel of ['loading_context', 'planning', 'awaiting_approval', 'executing', 'verifying', 'done', 'error']) {
      expect(runtimeTypes).not.toMatch(new RegExp(`['"]${lifecycleLabel}['"]`))
    }

    const chainMachine = readRepoFile(repoRoot, 'packages/orchestrator/src/chain-machine.ts')
    expect(chainMachine).not.toMatch(/recoveryClassifier|autoRecovery|lifecycleLabel|loading_context|awaiting_approval/i)
  })
})

function readRepoFile(repoRoot: string, repoRelativePath: string): string {
  return readFileSync(path.join(repoRoot, repoRelativePath), 'utf8')
}

function findRepoRoot(startPath: string): string {
  let dir = startPath
  for (let i = 0; i < 20; i++) {
    if (existsSync(path.join(dir, 'package.json')) && existsSync(path.join(dir, 'specs'))) {
      return dir
    }
    const parent = path.dirname(dir)
    if (parent === dir) break
    dir = parent
  }
  throw new Error(`Could not find repo root from: ${startPath}`)
}
