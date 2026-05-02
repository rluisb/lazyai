import { existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { scaffoldOrchestration } from '../scaffold/orchestration.js'
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
    assertCurrentFeatureChainShape(sourceChain)
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
    assertCurrentFeatureChainShape(installedChain)
  })

  it('copies chain definitions without rendering feature-flag template directives', async () => {
    const conditionalLibraryDir = path.join(tempDir, 'conditional-library')
    mkdirSync(path.join(conditionalLibraryDir, 'orchestration', 'chains'), { recursive: true })
    const chainWithTemplateDirectives = `{
  "kind": "chain",
  "name": "feature",
  "entry": "plan-quality",
  "steps": [
    {"id":"plan-quality","agent":"planner","skills":["plan"],"description":"quality","transitions":{"success":"red-team-plan"}},
    {{#if features.adversarialDesign}}
    {"id":"red-team-plan","agent":"reviewer","skills":["red-team-plan"],"description":"red team","transitions":{"success":"plan-gate"}},
    {{/if}}
    {"id":"plan-gate","agent":"planner","skills":[],"description":"gate","gate":"user_approval","transitions":{"approved":"implement","rejected":"plan-quality"}}
  ]
}`
    writeFileSync(path.join(conditionalLibraryDir, 'orchestration', 'chains', 'feature.json'), chainWithTemplateDirectives)

    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir: conditionalLibraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    const installed = readFileSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'), 'utf8')
    expect(installed).toBe(chainWithTemplateDirectives)
    expect(installed).toContain('{{#if features.adversarialDesign}}')
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

function assertCurrentFeatureChainShape(chain: Record<string, unknown>) {
  expect(chain.kind).toBe('chain')
  expect(chain.name).toBe('feature')
  expect(chain).not.toHaveProperty('parallel')
  expect(Array.isArray(chain.steps)).toBe(true)

  const steps = chain.steps as Array<Record<string, unknown>>
  expect(steps.map((step) => step.id)).toEqual(['research', 'plan', 'implement', 'review', 'fix', 'document'])

  const planStep = steps.find((step) => step.id === 'plan')
  expect(planStep).toBeDefined()
  expect(planStep?.gate).toBe('user_approval')
  expect(planStep?.transitions).toMatchObject({ approved: 'implement', rejected: 'research' })
}
