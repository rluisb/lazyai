import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { beforeEach, describe, expect, it } from 'vitest'
import { ClaudeCodeAdapter } from '../adapters/claude-code.js'
import { CodexAdapter } from '../adapters/codex.js'
import { CopilotAdapter } from '../adapters/copilot.js'
import { GeminiAdapter } from '../adapters/gemini.js'
import { OpenCodeAdapter } from '../adapters/opencode.js'
import type { FileRecord } from '../types.js'
import {
  copyDir,
  copyFile,
  ensureDir,
  fileExists,
  fileHash,
  listDir,
  readFile,
  writeFile,
} from '../utils/files.js'

function makeTempDir(prefix: string): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), prefix))
}

describe('file utils', () => {
  it('creates directories and writes/reads files', () => {
    const tempDir = makeTempDir('ai-setup-files-')
    const nested = path.join(tempDir, 'a/b/c')
    const filePath = path.join(nested, 'note.txt')

    ensureDir(nested)
    expect(fileExists(nested)).toBe(true)

    writeFile(filePath, 'hello world')
    expect(readFile(filePath)).toBe('hello world')
    expect(fileExists(filePath)).toBe(true)
  })

  it('copies files and directories recursively', () => {
    const tempDir = makeTempDir('ai-setup-copy-')
    const src = path.join(tempDir, 'src')
    const dest = path.join(tempDir, 'dest')
    ensureDir(path.join(src, 'nested'))
    writeFile(path.join(src, 'root.txt'), 'root')
    writeFile(path.join(src, 'nested/child.txt'), 'child')

    copyDir(src, dest)
    expect(readFile(path.join(dest, 'root.txt'))).toBe('root')
    expect(readFile(path.join(dest, 'nested/child.txt'))).toBe('child')

    const copiedSingle = path.join(tempDir, 'single/one.txt')
    copyFile(path.join(src, 'root.txt'), copiedSingle)
    expect(readFile(copiedSingle)).toBe('root')
  })

  it('hashes deterministically and listDir handles missing directories', () => {
    const tempDir = makeTempDir('ai-setup-hash-')
    const filePath = path.join(tempDir, 'hash.txt')
    writeFile(filePath, 'same content')

    const hash1 = fileHash(filePath)
    const hash2 = fileHash(filePath)
    expect(hash1).toBe(hash2)
    expect(hash1).toHaveLength(16)

    writeFile(filePath, 'different content')
    expect(fileHash(filePath)).not.toBe(hash1)

    expect(listDir(path.join(tempDir, 'does-not-exist'))).toEqual([])
  })

  it('throws when copying from a missing source directory', () => {
    const tempDir = makeTempDir('ai-setup-copy-error-')
    expect(() => copyDir(path.join(tempDir, 'missing'), path.join(tempDir, 'dest'))).toThrow(
      'Directory not found',
    )
  })
})

describe('tool adapters', () => {
  let libraryDir: string
  let targetDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    libraryDir = makeTempDir('ai-setup-library-')
    targetDir = makeTempDir('ai-setup-target-')
    fileRecords = []

    ensureDir(path.join(libraryDir, 'root'))
    ensureDir(path.join(libraryDir, 'agents'))
    ensureDir(path.join(libraryDir, 'prompts'))
    ensureDir(path.join(libraryDir, 'tool-agents'))
    writeFile(path.join(libraryDir, 'agents/builder.md'), '# builder')
    writeFile(path.join(libraryDir, 'agents/reviewer.md'), '# reviewer')
    writeFile(path.join(libraryDir, 'prompts/plan.md'), '# plan')
    writeFile(path.join(libraryDir, 'tool-agents/agents-dir.md'), '# agents context')
    writeFile(path.join(libraryDir, 'tool-agents/skills-dir.md'), '# skills context')
    writeFile(path.join(libraryDir, 'tool-agents/templates-dir.md'), '# templates context')
    writeFile(path.join(libraryDir, 'tool-agents/root-dir.md'), '# root context')
    writeFile(path.join(libraryDir, 'root/AGENTS.template.md'), '# [YOUR_PROJECT_NAME]\nRoot agent instructions')
    writeFile(path.join(libraryDir, 'root/GEMINI.template.md'), '# GEMINI root')
    writeFile(path.join(libraryDir, 'root/CLAUDE.template.md'), '# CLAUDE root')
    writeFile(path.join(libraryDir, 'root/copilot-instructions.template.md'), '# Copilot repo instructions')
  })

  it('OpenCode adapter installs agents and force-overwrites existing files with backup', async () => {
    const existingPath = path.join(targetDir, '.opencode/agents/builder.md')
    ensureDir(path.dirname(existingPath))
    writeFile(existingPath, 'customized')
    ensureDir(path.join(libraryDir, 'skills'))
    writeFile(path.join(libraryDir, 'skills', 'implement.md'), '---\nname: implement\ndescription: Implement\n---\n\n# implement')

    const adapter = new OpenCodeAdapter()
    await adapter.install({ targetDir, libraryDir, fileRecords, force: true })

    expect(readFile(existingPath)).toBe('# builder')
    expect(fileExists(path.join(targetDir, '.opencode/agents/reviewer.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/skills/implement/SKILL.md'))).toBe(true)
    expect(readFile(path.join(targetDir, '.opencode/skills/implement/SKILL.md'))).toContain('name: implement')
    expect(fileExists(path.join(targetDir, 'opencode.json'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/commands'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/templates'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.opencode/agents/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/skills/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.ai-setup-backup/.opencode/agents/builder.md'))).toBe(true)

    const config = JSON.parse(readFile(path.join(targetDir, 'opencode.json')))
    expect(config.$schema).toBe('https://opencode.ai/config.json')
    expect(config.instructions).toContain('AGENTS.md')

    expect(fileRecords.map((f) => f.path).sort()).toEqual([
      '.opencode/AGENTS.md',
      '.opencode/agents/AGENTS.md',
      '.opencode/agents/builder.md',
      '.opencode/agents/reviewer.md',
      '.opencode/skills/AGENTS.md',
      '.opencode/skills/implement/SKILL.md',
      'opencode.json',
    ])
    expect(fileRecords.some((f) => f.path === 'opencode.json')).toBe(true)
  })

  it('OpenCode adapter uses skills directory for global scope', async () => {
    const adapter = new OpenCodeAdapter()
    const globalTargetDir = path.join(targetDir, '.config', 'opencode')
    ensureDir(path.join(libraryDir, 'skills'))
    writeFile(path.join(libraryDir, 'skills', 'implement.md'), '# implement')

    await adapter.install({
      targetDir: globalTargetDir,
      setupScope: 'global',
      libraryDir,
      fileRecords,
      force: true,
    })

    expect(fileExists(path.join(globalTargetDir, 'agents', 'builder.md'))).toBe(true)
    expect(fileExists(path.join(globalTargetDir, 'skills', 'implement', 'SKILL.md'))).toBe(true)
    expect(fileExists(path.join(globalTargetDir, 'skills', 'AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(globalTargetDir, 'skills'))).toBe(true)
    expect(fileExists(path.join(globalTargetDir, 'command'))).toBe(false)
    expect(fileExists(path.join(globalTargetDir, 'commands'))).toBe(true)
  })

  it('Copilot adapter writes repo instructions, prompt files, and root AGENTS.md', async () => {
    const adapter = new CopilotAdapter()

    ensureDir(path.join(libraryDir, 'skills'))
    writeFile(path.join(libraryDir, 'skills', 'implement.md'), '# implement')

    await adapter.install({
      targetDir,
      libraryDir,
      fileRecords,
      force: true,
      selections: {
        agents: ['builder'],
        prompts: ['plan'],
        skills: ['implement'],
      },
    })

    expect(fileExists(path.join(targetDir, '.github/instructions'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.github/prompts/plan.prompt.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.github/prompts/implement.prompt.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.github/copilot-instructions.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)

    expect(fileExists(path.join(targetDir, '.github/agents'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.github/AGENTS.md'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.github/prompts/AGENTS.md'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.github/templates/plan.md'))).toBe(false)

    expect(fileRecords.some((f) => f.path === '.github/prompts/plan.prompt.md')).toBe(true)
    expect(fileRecords.some((f) => f.path === '.github/prompts/implement.prompt.md')).toBe(true)
    expect(fileRecords.some((f) => f.path === '.github/copilot-instructions.md')).toBe(true)
    expect(fileRecords.some((f) => f.path === 'AGENTS.md')).toBe(true)
  })

  it('Gemini adapter installs .gemini skills/<name>/SKILL.md and root GEMINI.md (no agents, no templates)', async () => {
    const adapter = new GeminiAdapter()

    ensureDir(path.join(libraryDir, 'skills'))
    writeFile(path.join(libraryDir, 'skills', 'implement.md'), '# implement')

    await adapter.install({
      targetDir,
      libraryDir,
      fileRecords,
      force: true,
      selections: {
        skills: ['implement'],
      },
    })

    // Skills use directory format: .gemini/skills/<name>/SKILL.md
    expect(fileExists(path.join(targetDir, '.gemini/skills/implement/SKILL.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.gemini/settings.json'))).toBe(true)
    expect(fileExists(path.join(targetDir, 'GEMINI.md'))).toBe(true)

    const settings = JSON.parse(readFile(path.join(targetDir, '.gemini/settings.json')))
    expect(settings.model.name).toBe('gemini-2.5-pro')

    // Gemini has NO agents or templates concepts
    expect(fileExists(path.join(targetDir, '.gemini/agents'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.gemini/templates'))).toBe(false)

    expect(fileRecords.some((f) => f.path === '.gemini/skills/implement/SKILL.md')).toBe(true)
    expect(fileRecords.some((f) => f.path === '.gemini/settings.json')).toBe(true)
    expect(fileRecords.some((f) => f.path === 'GEMINI.md')).toBe(true)
  })

  it('Claude Code adapter installs .claude layout and root CLAUDE.md', async () => {
    const adapter = new ClaudeCodeAdapter()

    ensureDir(path.join(libraryDir, 'skills'))
    writeFile(path.join(libraryDir, 'skills', 'implement.md'), '# implement')

    await adapter.install({
      targetDir,
      libraryDir,
      fileRecords,
      force: true,
      selections: {
        agents: ['builder'],
        skills: ['implement'],
        prompts: ['plan'],
      },
    })

    expect(fileExists(path.join(targetDir, '.claude/agents/builder.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.claude/skills/implement/SKILL.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.claude/settings.json'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.claude/commands'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.claude/templates'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.claude/rules'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.claude/rules/typescript.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, 'CLAUDE.md'))).toBe(true)

    expect(JSON.parse(readFile(path.join(targetDir, '.claude/settings.json')))).toEqual({
      permissions: {
        allow: [],
        deny: [],
      },
    })
    expect(readFile(path.join(targetDir, '.claude/rules/typescript.md'))).toContain('paths:')

    expect(fileRecords.some((f) => f.path === '.claude/agents/builder.md')).toBe(true)
    expect(fileRecords.some((f) => f.path === '.claude/settings.json')).toBe(true)
    expect(fileRecords.some((f) => f.path === '.claude/skills/implement/SKILL.md')).toBe(true)
    expect(fileRecords.some((f) => f.path === '.claude/rules/typescript.md')).toBe(true)
    expect(fileRecords.some((f) => f.path === 'CLAUDE.md')).toBe(true)
  })

  it('Codex adapter installs .agents skills directory-per-skill and root AGENTS.md', async () => {
    const adapter = new CodexAdapter()

    ensureDir(path.join(libraryDir, 'skills'))
    writeFile(path.join(libraryDir, 'skills', 'implement.md'), '# implement')
    writeFile(path.join(libraryDir, 'root', 'AGENTS.template.md'), '# [YOUR_PROJECT_NAME]\nCodex agent instructions')

    await adapter.install({
      targetDir,
      libraryDir,
      fileRecords,
      force: true,
      selections: {
        skills: ['implement'],
      },
    })

    // Codex uses AgentSkills standard: .agents/skills/<name>/SKILL.md
    expect(fileExists(path.join(targetDir, '.agents/skills/implement/SKILL.md'))).toBe(true)

    // Root AGENTS.md should exist
    expect(fileExists(path.join(targetDir, 'AGENTS.md'))).toBe(true)

    // No .codex directory should exist
    expect(fileExists(path.join(targetDir, '.codex'))).toBe(false)

    // No agents or templates directories (Codex has no separate agents dir)
    expect(fileExists(path.join(targetDir, '.agents/agents'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.agents/templates'))).toBe(false)

    expect(fileRecords.some((f) => f.path === '.agents/skills/implement/SKILL.md')).toBe(true)
    expect(fileRecords.some((f) => f.path === 'AGENTS.md')).toBe(true)
  })

  it('Codex adapter uses global scope correctly', async () => {
    const adapter = new CodexAdapter()
    const globalTargetDir = path.join(targetDir, '.config', 'codex')
    const globalAgentsDir = path.join(targetDir, '.config', '.agents')
    ensureDir(path.join(libraryDir, 'skills'))
    writeFile(path.join(libraryDir, 'skills', 'implement.md'), '# implement')
    writeFile(path.join(libraryDir, 'root', 'AGENTS.template.md'), '# [YOUR_PROJECT_NAME]\nCodex agent instructions')

    await adapter.install({
      targetDir: globalTargetDir,
      setupScope: 'global',
      libraryDir,
      fileRecords,
      force: true,
      selections: {
        skills: ['implement'],
      },
    })

    // Global scope: config stays in ~/.codex, shared skills go in ~/.agents/skills
    expect(fileExists(path.join(globalAgentsDir, 'skills', 'implement', 'SKILL.md'))).toBe(true)
    expect(fileExists(path.join(globalAgentsDir, 'skills', 'AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(globalAgentsDir, 'AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(globalTargetDir, 'AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(globalTargetDir, 'skills'))).toBe(false)
    expect(fileExists(path.join(globalTargetDir, '.agents'))).toBe(false)
  })
})
