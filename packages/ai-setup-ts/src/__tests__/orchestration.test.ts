import { existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { scaffoldOrchestration } from '../scaffold/orchestration.js'
import type { FeatureFlags } from '../store/schema.js'
import type { FileRecord } from '../types.js'
import { findMonorepoLibraryDir } from './test-helpers.js'

const libraryDir = findMonorepoLibraryDir()

describe('scaffoldOrchestration', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'ai-setup-orchestration-'))
    fileRecords = []
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('keeps the source feature chain sequential with the current explicit plan gate', () => {
    const sourceChain = readFeatureChain(path.join(libraryDir, 'orchestration', 'chains', 'feature.json'))
    assertFeatureChainShape(sourceChain, { adversarialDesign: false })
  })

  it('keeps the adversarial source feature chain sequential with red-team before the plan gate', () => {
    const sourceChain = readFeatureChain(path.join(libraryDir, 'orchestration', 'chains', 'feature-adversarial.json'))
    assertFeatureChainShape(sourceChain, { adversarialDesign: true })
  })

  it('copies the orchestration library tree into .ai/orchestration', async () => {
    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'teams', 'review-team.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'workflows', 'rpi.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'skills', 'domains', 'backend.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'skills', 'modes', 'senior.md'))).toBe(true)
    expect(fileRecords.some((record) => record.path === '.ai/orchestration/chains/feature.json')).toBe(true)

    const installedChain = readFeatureChain(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))
    assertFeatureChainShape(installedChain, { adversarialDesign: false })
  })

  it('installs the adversarial source as feature.json when adversarialDesign is enabled', async () => {
    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
      features: { adversarialDesign: true } as Partial<FeatureFlags>,
    })

    const installedChain = readFeatureChain(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))
    assertFeatureChainShape(installedChain, { adversarialDesign: true })
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature-adversarial.json'))).toBe(false)
  })

  it('installs the base feature source as feature.json when adversarialDesign is disabled', async () => {
    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
      features: { adversarialDesign: false } as Partial<FeatureFlags>,
    })

    const installedChain = readFeatureChain(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))
    assertFeatureChainShape(installedChain, { adversarialDesign: false })
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature-adversarial.json'))).toBe(false)
  })

  it('selects explicit chain sources instead of rendering feature-flag template directives', async () => {
    const explicitLibraryDir = path.join(tempDir, 'explicit-library')
    mkdirSync(path.join(explicitLibraryDir, 'orchestration', 'chains'), { recursive: true })
    writeFileSync(path.join(explicitLibraryDir, 'orchestration', 'chains', 'feature.json'), JSON.stringify(minimalFeatureChain(false), null, 2))
    writeFileSync(path.join(explicitLibraryDir, 'orchestration', 'chains', 'feature-adversarial.json'), JSON.stringify(minimalFeatureChain(true), null, 2))

    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir: explicitLibraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
      features: { adversarialDesign: true } as Partial<FeatureFlags>,
    })

    const installed = readFeatureChain(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))
    assertFeatureChainShape(installed, { adversarialDesign: true })
  })

  it('scaffolds the expected top-level orchestration directories', async () => {
    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    const entries = readdirSync(path.join(tempDir, '.ai', 'orchestration')).sort()
    expect(entries).toEqual(['chains', 'skills', 'teams', 'workflows'])
  })

  it('adds orchestration content from discovered extensions', async () => {
    const extensionDir = path.join(tempDir, '.ai', 'extensions', 'team-pack')
    mkdirSync(path.join(extensionDir, 'skills'), { recursive: true })
    mkdirSync(path.join(extensionDir, 'orchestration', 'chains'), { recursive: true })
    mkdirSync(path.join(extensionDir, 'orchestration', 'workflows'), { recursive: true })
    writeFileSync(path.join(extensionDir, 'skills', 'custom-skill.md'), '# Custom Skill')
    writeFileSync(path.join(extensionDir, 'orchestration', 'chains', 'release.json'), '{"name":"release"}')
    writeFileSync(path.join(extensionDir, 'orchestration', 'workflows', 'deploy.json'), '{"name":"deploy"}')

    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'release.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'workflows', 'deploy.json'))).toBe(true)
    expect(fileRecords.some((record) => record.path === '.ai/orchestration/chains/release.json')).toBe(true)
  })
})

function readFeatureChain(filePath: string): Record<string, unknown> {
  return JSON.parse(readFileSync(filePath, 'utf8')) as Record<string, unknown>
}

function assertFeatureChainShape(chain: Record<string, unknown>, opts: { adversarialDesign: boolean }) {
  expect(chain.kind).toBe('chain')
  expect(chain.name).toBe('feature')
  expect(chain).not.toHaveProperty('parallel')
  expect(Array.isArray(chain.steps)).toBe(true)

  const steps = chain.steps as Array<Record<string, unknown>>
  const raw = JSON.stringify(chain)
  expect(raw).not.toContain('{{#if')
  expect(raw).not.toContain('{{/if}}')
  expect(raw).not.toContain('optionalByFeature')
  expect(raw).not.toContain('condition')

  const expectedStepIDs = [
    'research',
    'plan',
    'plan-quality',
    ...(opts.adversarialDesign ? ['red-team-plan'] : []),
    'plan-gate',
    'implement',
    'review',
    'fix',
    'document',
  ]
  expect(steps.map((step) => step.id)).toEqual(expectedStepIDs)
  for (const step of steps) {
    expect(step).not.toHaveProperty('condition')
    expect(step).not.toHaveProperty('optionalByFeature')
    expect(step).not.toHaveProperty('parallel')
  }

  const planStep = steps.find((step) => step.id === 'plan')
  expect(planStep).toBeDefined()
  expect(planStep).not.toHaveProperty('gate')
  expect(planStep?.transitions).toMatchObject({ success: 'plan-quality' })

  const planQualityStep = steps.find((step) => step.id === 'plan-quality')
  expect(planQualityStep).toBeDefined()
  const stepAfterQuality = opts.adversarialDesign ? 'red-team-plan' : 'plan-gate'
  expect(planQualityStep?.transitions).toMatchObject({
    success: stepAfterQuality,
    pass: stepAfterQuality,
    warn: stepAfterQuality,
    fail: stepAfterQuality,
  })
  expect((planQualityStep?.transitions as Record<string, unknown>).fail).not.toBe('plan')

  const redTeamStep = steps.find((step) => step.id === 'red-team-plan')
  if (opts.adversarialDesign) {
    expect(redTeamStep).toBeDefined()
    expect(redTeamStep?.skills).toEqual(['red-team-plan'])
    expect(redTeamStep?.transitions).toMatchObject({ success: 'plan-gate', soft_fail: 'plan-gate', failure: 'plan-gate' })
  } else {
    expect(redTeamStep).toBeUndefined()
  }

  const stepsBeforeImplementation = steps.slice(0, steps.findIndex((step) => step.id === 'implement'))
  const approvalGatesBeforeImplementation = stepsBeforeImplementation.filter((step) => step.gate === 'user_approval')
  expect(approvalGatesBeforeImplementation.map((step) => step.id)).toEqual(['plan-gate'])

  const planGateStep = steps.find((step) => step.id === 'plan-gate')
  expect(planGateStep).toBeDefined()
  expect(planGateStep?.gate).toBe('user_approval')
  expect(planGateStep?.transitions).toMatchObject({ approved: 'implement', rejected: 'plan' })
}

function minimalFeatureChain(adversarialDesign: boolean): Record<string, unknown> {
  const planQualityTransitions = adversarialDesign
    ? { success: 'red-team-plan', pass: 'red-team-plan', warn: 'red-team-plan', fail: 'red-team-plan' }
    : { success: 'plan-gate', pass: 'plan-gate', warn: 'plan-gate', fail: 'plan-gate' }
  return {
    kind: 'chain',
    name: 'feature',
    entry: 'research',
    steps: [
      { id: 'research', agent: 'scout', skills: ['research'], transitions: { success: 'plan' } },
      { id: 'plan', agent: 'planner', skills: ['plan'], transitions: { success: 'plan-quality' } },
      { id: 'plan-quality', agent: 'planner', skills: ['plan'], transitions: planQualityTransitions },
      ...(adversarialDesign
        ? [{ id: 'red-team-plan', agent: 'red-team', skills: ['red-team-plan'], transitions: { success: 'plan-gate', soft_fail: 'plan-gate', failure: 'plan-gate' } }]
        : []),
      { id: 'plan-gate', agent: 'planner', skills: [], gate: 'user_approval', transitions: { approved: 'implement', rejected: 'plan' } },
      { id: 'implement', agent: 'builder', skills: ['implement'], transitions: { success: 'review' } },
      { id: 'review', agent: 'reviewer', skills: ['extract-standards'], transitions: { pass: 'document' } },
      { id: 'fix', agent: 'builder', skills: ['iterate'], transitions: { success: 'review' } },
      { id: 'document', agent: 'documenter', skills: [], transitions: { success: 'done' } },
    ],
  }
}
