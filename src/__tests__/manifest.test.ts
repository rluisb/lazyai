import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import type { AiSetupConfig, WizardSelections } from '../types.js'
import { extractSelections, readManifest } from '../utils/manifest.js'

function buildManifest(files: string[], selections?: WizardSelections): AiSetupConfig {
  return {
    version: '1.0.0',
    setupScope: 'project',
    setupType: 'project',
    tools: ['opencode'],
    projectName: 'test-project',
    installedAt: '2026-01-01T00:00:00.000Z',
    files: files.map(filePath => ({
      path: filePath,
      hash: 'hash',
      source: 'source',
    })),
    ...(selections ? { selections } : {}),
  }
}

describe('readManifest', () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-manifest-'))
  })

  afterEach(() => {
    fs.rmSync(tempDir, { recursive: true, force: true })
  })

  it('returns null when .ai-setup.json is missing', async () => {
    await expect(readManifest(tempDir)).resolves.toBeNull()
  })

  it('returns null when file contains malformed JSON', async () => {
    fs.writeFileSync(path.join(tempDir, '.ai-setup.json'), '{ invalid json', 'utf-8')
    await expect(readManifest(tempDir)).resolves.toBeNull()
  })

  it('returns parsed AiSetupConfig for valid manifest JSON', async () => {
    const manifest = buildManifest(['docs/features/example.md'])
    fs.writeFileSync(path.join(tempDir, '.ai-setup.json'), JSON.stringify(manifest), 'utf-8')

    const result = await readManifest(tempDir)
    expect(result?.version).toBe('1.0.0')
    expect(result?.setupScope).toBe('project')
    expect(result?.tools).toEqual(['opencode'])
    expect(result?.projectName).toBe('test-project')
    expect(result?.files).toHaveLength(1)
    expect(result?.files?.[0]?.path).toBe('docs/features/example.md')
  })
})

describe('extractSelections', () => {
  it('returns manifest.selections directly when present', () => {
    const selections: WizardSelections = {
      templates: ['adr'],
      rules: ['cost'],
      agents: ['builder'],
      skills: ['implement'],
      prompts: ['plan'],
      infra: ["pre-commit"],
      constitution: [],
    }
    const manifest = buildManifest(['docs/features/example.md'], selections)

    const result = extractSelections(manifest)

    expect(result).toBe(selections)
  })

  it('infers templates from docs/templates/*.md paths', () => {
    const manifest = buildManifest(['docs/templates/adr.md'])

    const result = extractSelections(manifest)

    expect(result.templates).toEqual(['adr'])
  })

  it('infers rules from docs/rules/*.md paths', () => {
    const manifest = buildManifest(['docs/rules/cost.md'])

    const result = extractSelections(manifest)

    expect(result.rules).toEqual(['cost'])
  })

  it('infers agents from supported tool directories', () => {
    const manifest = buildManifest(['.claude/builder.md'])

    const result = extractSelections(manifest)

    expect(result.agents).toEqual(['builder'])
  })

  it('infers skills from commands/skills/prompts paths', () => {
    const manifest = buildManifest(['.claude/skills/implement/SKILL.md'])

    const result = extractSelections(manifest)

    expect(result.skills).toEqual(['implement'])
  })

  it('infers infra from git hook and compliance files', () => {
    const manifest = buildManifest(['.git/hooks/pre-commit', 'docs/compliance.md'])

    const result = extractSelections(manifest)

    expect(result.infra).toEqual(['pre-commit', 'compliance'])
  })

  it('infers codeowners infra from CODEOWNERS file', () => {
    const manifest = buildManifest(['CODEOWNERS'])

    const result = extractSelections(manifest)

    expect(result.infra).toEqual(['codeowners'])
  })

  it('returns empty partial when manifest has empty files array', () => {
    const manifest = buildManifest([])

    const result = extractSelections(manifest)

    expect(result).toEqual({})
  })
})
