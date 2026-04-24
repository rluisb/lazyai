import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { scaffoldAgentsSkillsPrompts } from '../scaffold/agents-skills-prompts.js'
import { scaffoldConstitution } from '../scaffold/constitution.js'
import { scaffoldInfra } from '../scaffold/infra.js'
import { scaffoldRootFiles } from '../scaffold/root-files.js'
import { scaffoldSpecs } from '../scaffold/specs.js'
import { scaffoldTemplatesRules } from '../scaffold/templates-rules.js'
import type { ConflictStrategy, FileRecord } from '../types.js'

const libraryDir = path.resolve(process.cwd(), 'library')

describe('scaffoldSpecs', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'scaffold-specs-'))
    fileRecords = []
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('creates only selected specs directories', async () => {
    await scaffoldSpecs({
      targetDir: tempDir,
      libraryDir,
      specsDirs: ['features', 'bugfixes'],
      specsAgents: [],
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'specs', 'features'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'specs', 'bugfixes'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'specs', 'adrs'))).toBe(false)
    expect(fileRecords).toHaveLength(0)
  })

  it('creates no directories when specsDirs is empty', async () => {
    await scaffoldSpecs({
      targetDir: tempDir,
      libraryDir,
      specsDirs: [],
      specsAgents: [],
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'specs'))).toBe(false)
    expect(fileRecords).toHaveLength(0)
  })

  it('copies AGENTS.md for selected specsAgents', async () => {
    await scaffoldSpecs({
      targetDir: tempDir,
      libraryDir,
      specsDirs: ['features', 'bugfixes'],
      specsAgents: ['features'],
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'specs', 'features', 'AGENTS.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'specs', 'bugfixes', 'AGENTS.md'))).toBe(false)
    expect(fileRecords.length).toBeGreaterThan(0)
  })
})

describe('scaffoldTemplatesRules', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'scaffold-templates-rules-'))
    fileRecords = []
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('creates no template or rule files when templates and rules are empty', async () => {
    await scaffoldTemplatesRules({
      targetDir: tempDir,
      libraryDir,
      templates: [],
      rules: [],
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'specs', 'templates'))).toBe(false)
    expect(existsSync(path.join(tempDir, 'specs', 'rules'))).toBe(false)
    expect(existsSync(path.join(tempDir, 'specs', 'prompts', 'local-examples'))).toBe(true)
    expect(fileRecords.length).toBeGreaterThan(0)
  })

  it('copies only selected templates and rules', async () => {
    await scaffoldTemplatesRules({
      targetDir: tempDir,
      libraryDir,
      templates: ['task'],
      rules: ['security'],
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'specs', 'templates', 'task.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'specs', 'templates', 'prd.md'))).toBe(false)
    expect(existsSync(path.join(tempDir, 'specs', 'rules', 'security.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'specs', 'rules', 'review.md'))).toBe(false)
    expect(fileRecords.length).toBeGreaterThan(0)
  })
})

describe('scaffoldInfra', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'scaffold-infra-'))
    fileRecords = []
    mkdirSync(path.join(tempDir, '.git'), { recursive: true })
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('skips pre-commit when not in infra selection', async () => {
    await scaffoldInfra({
      targetDir: tempDir,
      libraryDir,
      infra: [],
      projectName: 'demo-project',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, '.git', 'hooks', 'pre-commit'))).toBe(false)
    expect(fileRecords).toHaveLength(0)
  })

  it('copies CODEOWNERS when codeowners infra is selected', async () => {
    await scaffoldInfra({
      targetDir: tempDir,
      libraryDir,
      infra: ['codeowners'],
      projectName: 'demo-project',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'CODEOWNERS'))).toBe(true)
    expect(fileRecords.some((record) => record.path === 'CODEOWNERS')).toBe(true)
  })

})

describe('scaffoldRootFiles', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'scaffold-root-files-'))
    fileRecords = []
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('creates correct files per tool', async () => {
    await scaffoldRootFiles({
      targetDir: tempDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'my-test-project',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'CLAUDE.md'))).toBe(false)
    expect(fileRecords.length).toBeGreaterThan(0)
  })

  it('replaces [YOUR_PROJECT_NAME] placeholder', async () => {
    await scaffoldRootFiles({
      targetDir: tempDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'placeholder-check',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    const content = readFileSync(path.join(tempDir, 'AGENTS.md'), 'utf-8')
    expect(content).toContain('# placeholder-check — AI Agent Rules')
    expect(content).not.toContain('[YOUR_PROJECT_NAME]')
    expect(fileRecords.length).toBeGreaterThan(0)
  })
})

describe('scaffoldAgentsSkillsPrompts', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'scaffold-adapters-'))
    fileRecords = []
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('installs files for selected tools and records outputs', async () => {
    await scaffoldAgentsSkillsPrompts({
      targetDir: tempDir,
      libraryDir,
      tools: ['opencode'],
      agents: ['builder'],
      skills: ['implement'],
      prompts: ['plan'],
      fileRecords,
      force: true,
    })

    expect(existsSync(path.join(tempDir, '.opencode', 'agents', 'builder.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.opencode', 'skills', 'implement', 'SKILL.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.opencode', 'commands'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.opencode', 'templates'))).toBe(false)
    expect(fileRecords.length).toBeGreaterThan(0)
  })

  it('does nothing when no tools are selected', async () => {
    await scaffoldAgentsSkillsPrompts({
      targetDir: tempDir,
      libraryDir,
      tools: [],
      agents: ['builder'],
      skills: ['implement'],
      prompts: ['plan'],
      fileRecords,
      force: true,
    })

    expect(existsSync(path.join(tempDir, '.opencode'))).toBe(false)
    expect(fileRecords).toHaveLength(0)
  })
})

describe('scaffoldConstitution', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'scaffold-constitution-'))
    fileRecords = []
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('creates .ai/constitution with all 4 files', async () => {
    await scaffoldConstitution({
      targetDir: tempDir,
      libraryDir,
      projectName: 'my-project',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, '.ai', 'constitution', 'constitution.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'constitution', 'constraints.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'constitution', 'quality-gates.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'constitution', 'uncertainty.md'))).toBe(true)
  })

  it('replaces [YOUR_PROJECT_NAME] placeholder', async () => {
    await scaffoldConstitution({
      targetDir: tempDir,
      libraryDir,
      projectName: 'placeholder-check',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    const constitution = readFileSync(path.join(tempDir, '.ai', 'constitution', 'constitution.md'), 'utf-8')
    expect(constitution).toContain('# placeholder-check Constitution')
    expect(constitution).not.toContain('[YOUR_PROJECT_NAME]')
  })

  it('skips existing files with skip strategy', async () => {
    const existingPath = path.join(tempDir, '.ai', 'constitution', 'constitution.md')
    mkdirSync(path.dirname(existingPath), { recursive: true })
    const existingContent = '# Existing Constitution\n'
    writeFileSync(existingPath, existingContent, 'utf-8')

    await scaffoldConstitution({
      targetDir: tempDir,
      libraryDir,
      projectName: 'ignored-name',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    const currentContent = readFileSync(existingPath, 'utf-8')
    expect(currentContent).toBe(existingContent)
  })

  it('records scaffolded files in fileRecords', async () => {
    await scaffoldConstitution({
      targetDir: tempDir,
      libraryDir,
      projectName: 'records-check',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(fileRecords).toHaveLength(4)
    expect(fileRecords.map((r) => r.path).sort()).toEqual([
      '.ai/constitution/constitution.md',
      '.ai/constitution/constraints.md',
      '.ai/constitution/quality-gates.md',
      '.ai/constitution/uncertainty.md',
    ])
  })
})
