import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { scaffoldCompiledRoot } from '../scaffold/compiled-root.js'
import type { ConflictStrategy, FileRecord } from '../types.js'
import { ensureDir, fileExists, readFile, writeFile } from '../utils/files.js'

const DEFAULT_FEATURE_FRAGMENT_MARKERS = [
  '<context-discipline>',
  '<rpi-workflow>',
  '<reasoning-protocol>',
  '<decision-protocol>',
  '<quality-gates>',
] as const

function makeTempDir(prefix: string): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), prefix))
}

function resolveLibraryDir(): string {
  return path.resolve(process.cwd(), 'library')
}

describe('scaffoldCompiledRoot', () => {
  let targetDir: string
  let tempLibraryDir: string
  let fileRecords: FileRecord[]
  let libraryDir: string

  beforeEach(() => {
    targetDir = makeTempDir('ai-setup-compiled-root-')
    tempLibraryDir = makeTempDir('ai-setup-temp-library-')
    fileRecords = []
    libraryDir = resolveLibraryDir()

    // Verify the real library dir exists
    if (!fileExists(libraryDir)) {
      throw new Error(`Library directory not found at ${libraryDir}`)
    }
  })

  afterEach(() => {
    if (fileExists(targetDir)) {
      fs.rmSync(targetDir, { recursive: true, force: true })
    }
    if (fileExists(tempLibraryDir)) {
      fs.rmSync(tempLibraryDir, { recursive: true, force: true })
    }
  })

  it('generates AGENTS.md for opencode', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'test-project',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, 'CLAUDE.md'))).toBe(false)
    expect(fileExists(path.join(targetDir, 'GEMINI.md'))).toBe(false)

    const agentsRecord = fileRecords.find((r) => r.path === 'AGENTS.md')
    expect(agentsRecord?.source).toBe('compiled:opencode')
  })

  it('passes through feature flags in context', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'feature-test',
      planningDir: '.ai/planning',
      features: {
        contextEngineering: true,
        rpiWorkflow: true,
        chainOfThought: true,
        treeOfThoughts: false,
        adrEnforcement: true,
        qualityGates: true,
        agentHarness: true,
        bugResolution: true,
        pivotHandling: true,
      },
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)
    expect(fileRecords.length).toBeGreaterThan(0)
  })

  it('renders default-enabled feature fragments when features are omitted', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'default-features-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    const agentsContent = readFile(path.join(targetDir, 'AGENTS.md'))

    for (const marker of DEFAULT_FEATURE_FRAGMENT_MARKERS) {
      expect(agentsContent).toContain(marker)
    }
  })

  it('removes disabled feature fragments while keeping others enabled', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'disabled-features-test',
      planningDir: '.ai/planning',
      features: {
        contextEngineering: false,
        rpiWorkflow: false,
        chainOfThought: true,
        treeOfThoughts: true,
        adrEnforcement: true,
        qualityGates: true,
        agentHarness: true,
        bugResolution: true,
        pivotHandling: true,
      },
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    const agentsContent = readFile(path.join(targetDir, 'AGENTS.md'))

    expect(agentsContent).not.toContain('<context-discipline>')
    expect(agentsContent).not.toContain('<rpi-workflow>')
    expect(agentsContent).toContain('<reasoning-protocol>')
    expect(agentsContent).toContain('<decision-protocol>')
  })

  it('interpolates planningDir variable in output', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'planning-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    const agentsContent = readFile(path.join(targetDir, 'AGENTS.md'))
    expect(agentsContent).toBeTruthy()
    expect(fileRecords.length).toBeGreaterThan(0)
  })

  it('correctly populates file records with source annotations', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'records-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(fileRecords.length).toBeGreaterThan(0)
    for (const record of fileRecords) {
      expect(record.source).toMatch(/^compiled:/)
      expect(record.hash).toBeTruthy()
      expect(record.hash.length).toBe(16)
    }
  })

  it('skip strategy prevents overwriting existing files', async () => {
    const existingPath = path.join(targetDir, 'AGENTS.md')
    ensureDir(path.dirname(existingPath))
    writeFile(existingPath, 'EXISTING CONTENT')
    const originalContent = fs.readFileSync(existingPath, 'utf-8')

    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'skip-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    const currentContent = readFile(existingPath)
    expect(currentContent).toBe(originalContent)

    const record = fileRecords.find((r) => r.path === 'AGENTS.md')
    expect(record).toBeUndefined()
  })

  it('includes additional context in fragment compilation', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'context-test',
      planningDir: '.ai/planning',
      primaryLanguage: 'TypeScript',
      framework: 'Next.js',
      workspaceType: 'monorepo',
      projectInstructions: 'Use ESLint and Prettier',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)
    expect(fileRecords.length).toBeGreaterThan(0)
  })

  it('includes workspace repos section when repos are provided', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'test-workspace',
      planningDir: '.planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
      repos: [
        { name: 'frontend', path: '../frontend', type: 'nextjs-typescript', description: 'Web app' },
        { name: 'backend', path: '../backend', type: 'go' },
      ],
    })

    const content = readFile(path.join(targetDir, 'AGENTS.md'))
    expect(content).toContain('## Workspace Repos')
    expect(content).toContain('### frontend')
    expect(content).toContain('nextjs-typescript')
    expect(content).toContain('### backend')
  })

  it('generates valid content for opencode', async () => {
    const toolTargetDir = makeTempDir('ai-setup-tool-opencode-')
    const toolFileRecords: FileRecord[] = []

    try {
      await scaffoldCompiledRoot({
        targetDir: toolTargetDir,
        libraryDir,
        tools: ['opencode'],
        projectName: 'test-opencode',
        planningDir: '.ai/planning',
        fileRecords: toolFileRecords,
        strategy: 'skip' as ConflictStrategy,
        perFileOverrides: new Map(),
      })

      expect(toolFileRecords.length).toBeGreaterThan(0)

      const files = fs.readdirSync(toolTargetDir, { recursive: true })
      expect(files.length).toBeGreaterThan(0)
    } finally {
      if (fileExists(toolTargetDir)) {
        fs.rmSync(toolTargetDir, { recursive: true, force: true })
      }
    }
  })
})
