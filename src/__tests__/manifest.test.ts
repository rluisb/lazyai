import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { extractSelections, readManifest } from '../utils/manifest.js'
import type { AiSetupConfig, WizardSelections } from '../types.js'

function buildManifest(files: string[], selections?: WizardSelections): AiSetupConfig {
  return {
    version: '1.0.0',
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

    await expect(readManifest(tempDir)).resolves.toEqual(manifest)
  })
})

describe('extractSelections', () => {
  it('returns manifest.selections directly when present', () => {
    const selections: WizardSelections = {
      docsDirs: ['features'],
      docsAgents: ['features'],
      templates: ['adr'],
      rules: ['cost'],
      agents: ['builder'],
      skills: ['implement'],
      prompts: ['plan'],
      infra: ['CODEOWNERS'],
    }
    const manifest = buildManifest(['docs/features/example.md'], selections)

    const result = extractSelections(manifest)

    expect(result).toBe(selections)
  })

  it('infers docsDirs from docs directory file paths', () => {
    const manifest = buildManifest(['docs/features/some-file.txt'])

    const result = extractSelections(manifest)

    expect(result.docsDirs).toEqual(['features'])
  })

  it('infers docsAgents from docs/*/AGENTS.md file paths', () => {
    const manifest = buildManifest(['docs/features/AGENTS.md'])

    const result = extractSelections(manifest)

    expect(result.docsAgents).toEqual(['features'])
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
    const manifest = buildManifest(['.claude/commands/implement.md'])

    const result = extractSelections(manifest)

    expect(result.skills).toEqual(['implement'])
  })

  it('infers infra from CODEOWNERS and git hook files', () => {
    const manifest = buildManifest(['CODEOWNERS', '.git/hooks/pre-commit'])

    const result = extractSelections(manifest)

    expect(result.infra).toEqual(['CODEOWNERS', 'pre-commit'])
  })

  it('returns empty partial when manifest has empty files array', () => {
    const manifest = buildManifest([])

    const result = extractSelections(manifest)

    expect(result).toEqual({})
  })
})
