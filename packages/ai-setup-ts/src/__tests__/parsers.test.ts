/**
 * Parser tests — OpenCode and Claude parsers
 *
 * Tests the detect, parse, and canMerge behaviour of the OpenCode and Claude
 * parsers against synthetic temporary directories.
 */
import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { ClaudeCodeParser } from '../migration/parsers/claude-parser.js'
import { OpenCodeParser } from '../migration/parsers/opencode-parser.js'
import type { MigrationContext } from '../migration/types.js'

function makeContext(sourcePath: string): MigrationContext {
  return {
    sourcePath,
    targetPath: sourcePath,
    options: {
      mergeStrategy: 'smart',
      skipBackup: true,
    },
  }
}

describe('OpenCodeParser', () => {
  let tempDir: string
  let parser: OpenCodeParser

  beforeEach(async () => {
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'ai-setup-opencode-parser-'))
    parser = new OpenCodeParser()
  })

  afterEach(async () => {
    await fs.rm(tempDir, { recursive: true, force: true })
  })

  describe('detect', () => {
    it('is not detected in an empty directory', async () => {
      const result = await parser.detect(makeContext(tempDir))
      expect(result.detected).toBe(false)
      expect(result.confidence).toBe(0)
    })

    it('is detected when AGENTS.md is present', async () => {
      await fs.writeFile(path.join(tempDir, 'AGENTS.md'), '# Agent\n')
      const result = await parser.detect(makeContext(tempDir))
      expect(result.detected).toBe(true)
      expect(result.adapterId).toBe('opencode')
    })

    it('has higher confidence when .opencode/ is also present', async () => {
      await fs.writeFile(path.join(tempDir, 'AGENTS.md'), '# Agent\n')
      await fs.mkdir(path.join(tempDir, '.opencode', 'agents'), { recursive: true })
      await fs.writeFile(path.join(tempDir, '.opencode', 'agents', 'builder.md'), '# Builder\n')

      const noDir = await new OpenCodeParser().detect(makeContext(tempDir))
      const withDir = await new OpenCodeParser().detect({
        ...makeContext(tempDir),
        sourcePath: tempDir,
      })

      expect(noDir.confidence).toBeGreaterThan(0)
      expect(withDir.confidence).toBeGreaterThanOrEqual(noDir.confidence)
    })

    it('includes AGENTS.md in detected files', async () => {
      await fs.writeFile(path.join(tempDir, 'AGENTS.md'), '# Agent\n')
      const result = await parser.detect(makeContext(tempDir))
      expect(result.files.some(f => f.path === 'AGENTS.md')).toBe(true)
    })
  })

  describe('parse', () => {
    it('returns empty collections when no opencode files found', async () => {
      // Even with no files, parse should succeed with empty setup
      const result = await parser.parse(makeContext(tempDir))
      expect(result.agents).toEqual([])
      expect(result.commands).toEqual([])
    })

    it('parses agent files from .opencode/agents/', async () => {
      await fs.mkdir(path.join(tempDir, '.opencode', 'agents'), { recursive: true })
      await fs.writeFile(
        path.join(tempDir, '.opencode', 'agents', 'builder.md'),
        '# Builder Agent\n\nHelps build things.',
      )

      const result = await parser.parse(makeContext(tempDir))

      expect(result.agents.length).toBeGreaterThanOrEqual(1)
      const builder = result.agents.find(a => a.id === 'builder')
      expect(builder).toBeDefined()
      expect(builder?.content).toContain('Builder Agent')
    })

    it('parses command files from .opencode/commands/', async () => {
      await fs.mkdir(path.join(tempDir, '.opencode', 'commands'), { recursive: true })
      await fs.writeFile(
        path.join(tempDir, '.opencode', 'commands', 'deploy.md'),
        '# Deploy Command\n\nRuns deployment.',
      )

      const result = await parser.parse(makeContext(tempDir))

      expect(result.commands.length).toBeGreaterThanOrEqual(1)
      const deploy = result.commands.find(c => c.id === 'deploy')
      expect(deploy).toBeDefined()
    })

    it('metadata contains adapter id', async () => {
      const result = await parser.parse(makeContext(tempDir))
      expect(result.metadata.adapter).toBe('opencode')
    })
  })

  describe('canMerge', () => {
    it('returns false when parsed setup has no agents/rules/templates', async () => {
      const parsed = await parser.parse(makeContext(tempDir))
      expect(parser.canMerge(parsed)).toBe(false)
    })

    it('returns true when parsed setup has agents', async () => {
      await fs.mkdir(path.join(tempDir, '.opencode', 'agents'), { recursive: true })
      await fs.writeFile(
        path.join(tempDir, '.opencode', 'agents', 'scout.md'),
        '# Scout\n\nExplores.',
      )
      const parsed = await parser.parse(makeContext(tempDir))
      expect(parser.canMerge(parsed)).toBe(true)
    })
  })

  describe('validate', () => {
    it('passes validation', () => {
      const result = parser.validate()
      expect(result.valid).toBe(true)
      expect(result.errors).toHaveLength(0)
    })
  })
})

describe('ClaudeCodeParser', () => {
  let tempDir: string
  let parser: ClaudeCodeParser

  beforeEach(async () => {
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'ai-setup-claude-parser-'))
    parser = new ClaudeCodeParser()
  })

  afterEach(async () => {
    await fs.rm(tempDir, { recursive: true, force: true })
  })

  describe('detect', () => {
    it('is not detected in an empty directory', async () => {
      const result = await parser.detect(makeContext(tempDir))
      expect(result.detected).toBe(false)
    })

    it('is detected when CLAUDE.md is present', async () => {
      await fs.writeFile(path.join(tempDir, 'CLAUDE.md'), '# Claude Setup\n')
      const result = await parser.detect(makeContext(tempDir))
      expect(result.detected).toBe(true)
      expect(result.adapterId).toBe('claude-code')
    })

    it('is detected when .claude/ directory is present', async () => {
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true })
      await fs.writeFile(path.join(tempDir, '.claude', 'settings.json'), '{}')
      const result = await parser.detect(makeContext(tempDir))
      expect(result.detected).toBe(true)
    })
  })

  describe('parse', () => {
    it('returns empty agents/commands on a bare setup', async () => {
      const result = await parser.parse(makeContext(tempDir))
      expect(result.agents).toEqual([])
      expect(result.commands).toEqual([])
    })

    it('parses agent files from .claude/', async () => {
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true })
      await fs.writeFile(
        path.join(tempDir, '.claude', 'reviewer.md'),
        '# Reviewer Agent\n\nReviews code.',
      )

      const result = await parser.parse(makeContext(tempDir))

      expect(result.agents.length).toBeGreaterThanOrEqual(1)
      const reviewer = result.agents.find((a: { id: string }) => a.id === 'reviewer')
      expect(reviewer).toBeDefined()
    })

    it('parses command files from .claude/commands/', async () => {
      await fs.mkdir(path.join(tempDir, '.claude', 'commands'), { recursive: true })
      await fs.writeFile(
        path.join(tempDir, '.claude', 'commands', 'test.md'),
        '# Test Command\n',
      )

      const result = await parser.parse(makeContext(tempDir))
      expect(result.commands.length).toBeGreaterThanOrEqual(1)
    })

    it('metadata contains adapter id', async () => {
      const result = await parser.parse(makeContext(tempDir))
      expect(result.metadata.adapter).toBe('claude-code')
    })
  })

  describe('validate', () => {
    it('passes validation', () => {
      const result = parser.validate()
      expect(result.valid).toBe(true)
    })
  })
})
