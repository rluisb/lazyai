import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { mkdtempSync, rmSync, existsSync, mkdirSync, readFileSync } from 'node:fs'
import path from 'node:path'
import { tmpdir } from 'node:os'
import type { FileRecord, ConflictStrategy } from '../types.js'

import { scaffoldDocs } from '../scaffold/docs.js'
import { scaffoldTemplatesRules } from '../scaffold/templates-rules.js'
import { scaffoldInfra } from '../scaffold/infra.js'
import { scaffoldRootFiles } from '../scaffold/root-files.js'
import { scaffoldAgentsSkillsPrompts } from '../scaffold/agents-skills-prompts.js'

const libraryDir = path.resolve(process.cwd(), 'library')

describe('scaffoldDocs', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'scaffold-docs-'))
    fileRecords = []
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('creates only selected docs directories', async () => {
    await scaffoldDocs({
      targetDir: tempDir,
      libraryDir,
      docsDirs: ['features', 'bugfixes'],
      docsAgents: [],
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'docs', 'features'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'docs', 'bugfixes'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'docs', 'adrs'))).toBe(false)
    expect(fileRecords).toHaveLength(0)
  })

  it('creates no directories when docsDirs is empty', async () => {
    await scaffoldDocs({
      targetDir: tempDir,
      libraryDir,
      docsDirs: [],
      docsAgents: [],
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'docs'))).toBe(false)
    expect(fileRecords).toHaveLength(0)
  })

  it('copies AGENTS.md for selected docsAgents', async () => {
    await scaffoldDocs({
      targetDir: tempDir,
      libraryDir,
      docsDirs: ['features', 'bugfixes'],
      docsAgents: ['features'],
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'docs', 'features', 'AGENTS.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'docs', 'bugfixes', 'AGENTS.md'))).toBe(false)
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

    expect(existsSync(path.join(tempDir, 'docs', 'templates'))).toBe(false)
    expect(existsSync(path.join(tempDir, 'docs', 'rules'))).toBe(false)
    expect(existsSync(path.join(tempDir, 'docs', 'prompts', 'local-examples'))).toBe(true)
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

    expect(existsSync(path.join(tempDir, 'docs', 'templates', 'task.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'docs', 'templates', 'prd.md'))).toBe(false)
    expect(existsSync(path.join(tempDir, 'docs', 'rules', 'security.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'docs', 'rules', 'review.md'))).toBe(false)
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
      tools: ['opencode', 'copilot'],
      projectName: 'my-test-project',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.github', 'copilot-instructions.md'))).toBe(true)
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
    expect(existsSync(path.join(tempDir, '.opencode', 'commands', 'implement.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.opencode', 'templates', 'plan.md'))).toBe(true)
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
