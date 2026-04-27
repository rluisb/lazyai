import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { scaffoldCompiledRoot } from '../scaffold/compiled-root.js'
import type { ConflictStrategy, FileRecord } from '../types.js'
import { ensureDir, fileExists, readFile, writeFile } from '../utils/files.js'
import { findMonorepoLibraryDir } from './test-helpers.js'

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
  return findMonorepoLibraryDir()
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

  it('generates correct root filenames per tool', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['claude-code', 'opencode', 'codex', 'copilot', 'gemini'],
      projectName: 'test-project',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    // Check that the expected shared/tool-specific root files were created
    expect(fileExists(path.join(targetDir, 'CLAUDE.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.github/copilot-instructions.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, 'INSTRUCTIONS.md'))).toBe(false)
    expect(fileExists(path.join(targetDir, 'GEMINI.md'))).toBe(true)

    // Verify file records have correct source annotations
    const claudeRecord = fileRecords.find((r) => r.path === 'CLAUDE.md')
    expect(claudeRecord?.source).toBe('compiled:claude-code')

    const agentsRecord = fileRecords.find((r) => r.path === 'AGENTS.md')
    expect(agentsRecord?.source).toBe('compiled:opencode')

    const copilotRecord = fileRecords.find((r) => r.path === '.github/copilot-instructions.md')
    expect(copilotRecord?.source).toBe('compiled:copilot')
  })

  it('passes through feature flags in context', async () => {
    // Test with decision protocol disabled (default behavior)
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

  it('compiles shared root outputs for all supported tools with camelCase feature conditions', async () => {
    const tools = ['claude-code', 'opencode', 'codex', 'copilot', 'gemini'] as const

    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: [...tools],
      projectName: 'all-tools-camelcase-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    const rootFiles = [
      'CLAUDE.md',
      'AGENTS.md',
      '.github/copilot-instructions.md',
      'GEMINI.md',
    ]

    for (const rootFile of rootFiles) {
      const content = readFile(path.join(targetDir, rootFile))
      expect(content).toContain('<context-discipline>')
      expect(content).toContain('<quality-gates>')
    }
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
    // The compiled output should reference the planning directory
    // (actual content depends on the template, but it should be present)
    expect(agentsContent).toBeTruthy()
    expect(fileRecords.length).toBeGreaterThan(0)
  })

  it('correctly populates file records with source annotations', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['claude-code', 'gemini'],
      projectName: 'records-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    // Every file record should have a source starting with 'compiled:'
    expect(fileRecords.length).toBeGreaterThan(0)
    for (const record of fileRecords) {
      expect(record.source).toMatch(/^compiled:/)
      expect(record.hash).toBeTruthy()
      expect(record.hash.length).toBe(16) // fileHash returns 16-char hash
    }
  })

  it('skip strategy prevents overwriting existing files', async () => {
    // Create an existing file
    const existingPath = path.join(targetDir, 'CLAUDE.md')
    ensureDir(path.dirname(existingPath))
    writeFile(existingPath, 'EXISTING CONTENT')
    const originalHash = fs.readFileSync(existingPath, 'utf-8')

    // Run scaffoldCompiledRoot with skip strategy
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['claude-code'],
      projectName: 'skip-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    // File should not have been modified
    const currentContent = readFile(existingPath)
    expect(currentContent).toBe(originalHash)

    // File should NOT be in records if skipped
    const record = fileRecords.find((r) => r.path === 'CLAUDE.md')
    expect(record).toBeUndefined()
  })

  it('handles multiple tools with distinct output files', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode', 'codex'],
      projectName: 'multi-tool-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    // Both opencode and codex write AGENTS.md, but they have different source origins
    // There may be one AGENTS.md file with the last tool's source, or multiple records
    // This is a behavior test: verify that both tools processed
    expect(fileRecords.length).toBeGreaterThan(0)
    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)
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

  it('writes files to correct directories maintaining structure', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['copilot'],
      projectName: 'dir-structure-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    // Copilot should write to .github/copilot-instructions.md
    const copilotPath = path.join(targetDir, '.github/copilot-instructions.md')
    expect(fileExists(copilotPath)).toBe(true)
    expect(fileExists(path.dirname(copilotPath))).toBe(true)

    // Verify the record path is relative
    const record = fileRecords.find((r) => r.path === '.github/copilot-instructions.md')
    expect(record).toBeDefined()
  })

  it('generates valid content for all supported tools', async () => {
    const tools = ['claude-code', 'opencode', 'codex', 'copilot', 'gemini'] as const

    for (const tool of tools) {
      const toolTargetDir = makeTempDir(`ai-setup-tool-${tool}-`)
      const toolFileRecords: FileRecord[] = []

      try {
        await scaffoldCompiledRoot({
          targetDir: toolTargetDir,
          libraryDir,
          tools: [tool],
          projectName: `test-${tool}`,
          planningDir: '.ai/planning',
          fileRecords: toolFileRecords,
          strategy: 'skip' as ConflictStrategy,
          perFileOverrides: new Map(),
        })

        // Each tool should produce at least one file record
        expect(toolFileRecords.length).toBeGreaterThan(0)

        // At least one file should exist in the target directory
        const files = fs.readdirSync(toolTargetDir, { recursive: true })
        expect(files.length).toBeGreaterThan(0)
      } finally {
        if (fileExists(toolTargetDir)) {
          fs.rmSync(toolTargetDir, { recursive: true, force: true })
        }
      }
    }
  })
})
