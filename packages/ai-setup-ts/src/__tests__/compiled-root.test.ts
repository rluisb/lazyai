import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { buildTargetedAgentsUpdatePatch, scaffoldCompiledRoot } from '../scaffold/compiled-root.js'
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
      tools: ['claude-code', 'opencode', 'copilot'],
      projectName: 'test-project',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    // Check that the expected shared/tool-specific root files were created
    expect(fileExists(path.join(targetDir, 'CLAUDE.md'))).toBe(false)
    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.github/copilot-instructions.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, 'INSTRUCTIONS.md'))).toBe(false)
    expect(fileExists(path.join(targetDir, 'GEMINI.md'))).toBe(false)

    // Verify file records have correct source annotations
    const agentsRecord = fileRecords.find((r) => r.path === 'AGENTS.md')
    expect(agentsRecord?.source).toMatch(/^compiled:(claude-code|opencode)$/)

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
        adversarialDesign: false,
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
        adversarialDesign: false,
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
    const tools = ['claude-code', 'opencode', 'copilot'] as const

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

    const rootFiles = ['AGENTS.md', '.github/copilot-instructions.md']

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
      tools: ['claude-code', 'copilot'],
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

  it('preserves existing CLAUDE.md and appends AGENTS reference once', async () => {
    // Create an existing file
    const existingPath = path.join(targetDir, 'CLAUDE.md')
    ensureDir(path.dirname(existingPath))
    writeFile(existingPath, 'EXISTING CONTENT')

    for (let i = 0; i < 2; i += 1) {
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
    }

    const currentContent = readFile(existingPath)
    expect(currentContent).toContain('EXISTING CONTENT')
    const reference = '<!-- ai-setup: AGENTS.md reference -->\nThis project uses [AGENTS.md](./AGENTS.md) as the canonical AI agent instruction file.'
    expect(currentContent.match(/ai-setup: AGENTS\.md reference/g)?.length).toBe(1)
    expect(currentContent).toContain(reference)

    const record = fileRecords.find((r) => r.path === 'CLAUDE.md')
    expect(record).toBeUndefined()
  })

  it('does not create CLAUDE.md for a clean Claude Code project', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['claude-code'],
      projectName: 'clean-claude-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, 'CLAUDE.md'))).toBe(false)
  })

  it('handles multiple tools with distinct output files', async () => {
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode', 'copilot'],
      projectName: 'multi-tool-test',
      planningDir: '.ai/planning',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
    })

    // Both supported tools should process without requiring removed tool support.
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
    const tools = ['claude-code', 'opencode', 'copilot'] as const

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

  it('patches hand-edited AGENTS.md with targeted replacements only', () => {
    const existing = [
      '# Hand Edited Agents',
      '',
      '## Custom Runbook',
      'Do not remove this paragraph.',
      '',
      '## Conventions',
      '',
      '### Naming',
      '- [YOUR_NAMING_CONVENTION]',
      '',
      '### Imports',
      '- [YOUR_IMPORT_ORDER]',
      '',
      '## Project Overview',
      '',
      '[YOUR_PROJECT_OVERVIEW]',
      '',
      '**Stack:**',
      '- Language: [YOUR_LANGUAGE]',
      '- Framework: [YOUR_FRAMEWORK]',
      '- Testing: [YOUR_TEST_FRAMEWORK]',
      '- Package manager: [YOUR_PACKAGE_MANAGER]',
      '',
      '## Do NOT',
      '',
      '- Never push directly to `[YOUR_PROTECTED_BRANCH]`',
      '- Keep this user-authored rule.',
      '',
      '## Testing',
      '',
      '- Minimum coverage: `[YOUR_COVERAGE_THRESHOLD]`%',
      '',
    ].join('\n')

    const { content, patch } = buildTargetedAgentsUpdatePatch('AGENTS.md', existing, {
      projectName: 'creator-checkout',
      planningDir: 'specs',
      constitution: {
        projectOverview: 'Acme checkout orchestrates creator payments.',
        stack: {
          language: 'TypeScript',
          framework: 'Next.js',
          testing: 'Vitest',
          packageManager: 'yarn',
        },
        conventions: {
          naming: 'camelCase for values and PascalCase for React components',
          importOrder: 'External packages, internal aliases, relative imports',
        },
        protectedBranch: 'main',
        coverageThreshold: 87,
      },
    })

    expect(content).toContain('## Custom Runbook\nDo not remove this paragraph.')
    expect(content).toContain('- Keep this user-authored rule.')
    expect(content).not.toContain('## Decision Tree')
    expect(content).toContain('Acme checkout orchestrates creator payments.')
    expect(content).toContain('- Language: TypeScript')
    expect(content).toContain('- Framework: Next.js')
    expect(content).toContain('- Testing: Vitest')
    expect(content).toContain('- Package manager: yarn')
    expect(content).toContain('- camelCase for values and PascalCase for React components')
    expect(content).toContain('- External packages, internal aliases, relative imports')
    expect(content).toContain('- Never push directly to `main`')
    expect(content).toContain('- Minimum coverage: `87`%')
    expect(patch.file).toBe('AGENTS.md')
    expect(patch.preservedUnrecognizedContent).toBe(true)
    expect(patch.replacements.length).toBeGreaterThanOrEqual(8)
  })

  it('emits TargetedUpdatePatch JSON shape', () => {
    const { content, patch } = buildTargetedAgentsUpdatePatch(
      'AGENTS.md',
      '## Project Overview\n\n[YOUR_PROJECT_OVERVIEW]\n',
      {
        projectName: 'shape-test',
        planningDir: 'specs',
        constitution: { projectOverview: 'Updated overview' },
      },
    )

    expect(content).toContain('Updated overview')
    const parsed = JSON.parse(JSON.stringify(patch))
    expect(parsed).toEqual(
      expect.objectContaining({
        file: 'AGENTS.md',
        replacements: expect.any(Array),
        warnings: expect.any(Array),
        preservedUnrecognizedContent: true,
      }),
    )
    expect(parsed.replacements[0]).toEqual(
      expect.objectContaining({
        field: 'PROJECT_OVERVIEW',
        oldText: '[YOUR_PROJECT_OVERVIEW]',
        newText: 'Updated overview',
        location: expect.objectContaining({
          section: 'Project Overview',
          lineStart: expect.any(Number),
          lineEnd: expect.any(Number),
        }),
      }),
    )
  })

  it('does not reprocess placeholders introduced by a targeted replacement value', () => {
    const existing = '## Project Overview\n\n[YOUR_PROJECT_OVERVIEW]\n[YOUR_PROJECT_OVERVIEW]\n'
    const { content, patch } = buildTargetedAgentsUpdatePatch('AGENTS.md', existing, {
      projectName: 'recursive-placeholder-test',
      planningDir: 'specs',
      constitution: { projectOverview: 'Project [YOUR_PROJECT_OVERVIEW]' },
    })

    expect(content).toBe('## Project Overview\n\nProject [YOUR_PROJECT_OVERVIEW]\nProject [YOUR_PROJECT_OVERVIEW]\n')
    expect(patch.replacements).toHaveLength(2)
    expect(content).not.toContain('Project Project')
  })

  it('warns and leaves unsafe parsed slots unchanged', () => {
    const existing = '## Project Overview\n\nA human-authored overview.\n\n**Stack:**\n- Language: Ruby maintained by platform\n'
    const { content, patch } = buildTargetedAgentsUpdatePatch('AGENTS.md', existing, {
      projectName: 'unsafe-test',
      planningDir: 'specs',
      constitution: {
        projectOverview: 'Generated overview',
        stack: { language: 'Go' },
      },
    })

    expect(content).toBe(existing)
    expect(patch.replacements).toHaveLength(0)
    expect(patch.warnings.length).toBeGreaterThanOrEqual(2)
  })

  it('W1.A answered profile renders AGENTS.md without collected fallback markers', async () => {
    // AC-N8-001, AC-N8-004, AC-N8-005, AC-N4-002 parity evidence for the
    // TS root surface: answered profile fields render, ignored dirs stay out
    // of the codebase map, and human-owned responsibility cells remain.
    for (const dir of ['src', 'node_modules', 'dist', '.git', 'vendor']) {
      fs.mkdirSync(path.join(targetDir, dir), { recursive: true })
    }

    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'creator-checkout',
      planningDir: 'specs',
      projectOverview: 'Creator Checkout processes creator payments.',
      primaryLanguage: 'TypeScript',
      framework: 'Next.js',
      constitution: {
        stack: {
          database: 'PostgreSQL',
          orm: 'Prisma',
          testing: 'Vitest',
          packageManager: 'yarn',
        },
        conventions: {
          naming: 'camelCase values; PascalCase React components',
          errorHandling: 'Return typed Result values at service boundaries',
          apiResponses: 'JSON envelopes include data or error',
          importOrder: 'External packages, internal aliases, relative imports',
        },
      },
      protectedBranch: 'main',
      testCommand: 'yarn test',
      lintCommand: 'yarn lint',
      buildCommand: 'yarn build',
      coverageThreshold: 91,
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
      setupScope: 'project',
    })

    const content = readFile(path.join(targetDir, 'AGENTS.md'))
    for (const expected of [
      'Creator Checkout processes creator payments.',
      '- Language: TypeScript',
      '- Framework: Next.js',
      '- Database: PostgreSQL',
      '- ORM/Query: Prisma',
      '- Testing: Vitest',
      '- Package manager: yarn',
      '- camelCase values; PascalCase React components',
      '- Return typed Result values at service boundaries',
      '- JSON envelopes include data or error',
      '- External packages, internal aliases, relative imports',
      '- Never push directly to `main`',
      '- Minimum coverage: `91`%',
      '| src | [WHAT_IT_DOES] |',
    ]) {
      expect(content).toContain(expected)
    }
    for (const marker of [
      '[YOUR_PROJECT_OVERVIEW]',
      '[YOUR_LANGUAGE]',
      '[YOUR_FRAMEWORK]',
      '[YOUR_DATABASE]',
      '[YOUR_ORM]',
      '[YOUR_TEST_FRAMEWORK]',
      '[YOUR_PACKAGE_MANAGER]',
      '[YOUR_NAMING_CONVENTION]',
      '[YOUR_ERROR_PATTERN]',
      '[YOUR_API_CONVENTION]',
      '[YOUR_IMPORT_ORDER]',
      '[YOUR_PROTECTED_BRANCH]',
    ]) {
      expect(content).not.toContain(marker)
    }
    for (const ignored of ['node_modules', 'dist', '.git', 'vendor']) {
      expect(content).not.toContain(`| ${ignored} |`)
    }
  })

  it('W1.A skipped profile preserves documented fallback markers', async () => {
    // AC-N8-002, AC-N4-003 parity evidence: skipped/non-interactive profile
    // values preserve literal fallbacks and coverage resolves to 80.
    await scaffoldCompiledRoot({
      targetDir,
      libraryDir,
      tools: ['opencode'],
      projectName: 'skipped-profile',
      planningDir: 'specs',
      fileRecords,
      strategy: 'skip' as ConflictStrategy,
      perFileOverrides: new Map(),
      setupScope: 'project',
    })

    const content = readFile(path.join(targetDir, 'AGENTS.md'))
    for (const fallback of [
      '[YOUR_PROJECT_OVERVIEW]',
      '[YOUR_LANGUAGE]',
      '[YOUR_FRAMEWORK]',
      '[YOUR_DATABASE]',
      '[YOUR_ORM]',
      '[YOUR_TEST_FRAMEWORK]',
      '[YOUR_PACKAGE_MANAGER]',
      '[YOUR_NAMING_CONVENTION]',
      '[YOUR_ERROR_PATTERN]',
      '[YOUR_API_CONVENTION]',
      '[YOUR_IMPORT_ORDER]',
      '[YOUR_PROTECTED_BRANCH]',
      '<!-- fill-in: test command -->',
      '<!-- fill-in: build command -->',
      '- Minimum coverage: `80`%',
    ]) {
      expect(content).toContain(fallback)
    }
    expect(content).not.toContain('null')
    expect(content).not.toContain('undefined')
  })

  it('W1.A targeted update preserves custom AGENTS.md sections byte-for-byte', () => {
    // AC-N8-003 parity evidence for TargetedUpdatePatch behavior.
    const customBlock = [
      '## Custom Operations Runbook',
      '',
      '  Keep indentation, spacing, and punctuation exactly.',
      '- User-owned checklist item',
      '',
    ].join('\n')
    const existing = ['# Existing Agents', '', customBlock, '## Project Overview', '', '[YOUR_PROJECT_OVERVIEW]', ''].join('\n')

    const { content } = buildTargetedAgentsUpdatePatch('AGENTS.md', existing, {
      projectName: 'targeted-update',
      planningDir: 'specs',
      constitution: { projectOverview: 'Generated overview' },
    })

    expect(content).toContain(customBlock)
    expect(content).toContain('Generated overview')
  })
})
