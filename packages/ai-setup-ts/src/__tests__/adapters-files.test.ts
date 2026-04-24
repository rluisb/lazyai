import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { beforeEach, describe, expect, it } from 'vitest'
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
    writeFile(
      path.join(libraryDir, 'agents/orchestrator.md'),
      [
        '---',
        'name: Orchestrator',
        'model: opus',
        'tools: list_catalog compose_agent start_chain advance_chain get_status get_budget retry_step escalate_step handoff',
        '---',
        '',
        '# Orchestrator',
        '',
        'Use start_chain, advance_chain, and get_status.',
      ].join('\n'),
    )
    writeFile(path.join(libraryDir, 'agents/reviewer.md'), '# reviewer')
    writeFile(path.join(libraryDir, 'prompts/plan.md'), '# plan')
    writeFile(path.join(libraryDir, 'tool-agents/agents-dir.md'), '# agents context')
    writeFile(path.join(libraryDir, 'tool-agents/skills-dir.md'), '# skills context')
    writeFile(path.join(libraryDir, 'tool-agents/templates-dir.md'), '# templates context')
    writeFile(path.join(libraryDir, 'tool-agents/root-dir.md'), '# root context')
    writeFile(path.join(libraryDir, 'root/AGENTS.template.md'), '# [YOUR_PROJECT_NAME]\nRoot agent instructions')
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
    expect(fileExists(path.join(targetDir, 'opencode.jsonc'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/commands'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/templates'))).toBe(false)
    expect(fileExists(path.join(targetDir, '.opencode/agents/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/skills/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.ai-setup-backup/.opencode/agents/builder.md'))).toBe(true)

    const config = JSON.parse(readFile(path.join(targetDir, 'opencode.jsonc')))
    expect(config.$schema).toBe('https://opencode.ai/config.json')
    expect(config.instructions).toContain('AGENTS.md')

    expect(fileRecords.map((f) => f.path).sort()).toEqual([
      '.opencode/AGENTS.md',
      '.opencode/agents/AGENTS.md',
      '.opencode/agents/builder.md',
      '.opencode/agents/reviewer.md',
      '.opencode/skills/AGENTS.md',
      '.opencode/skills/implement/SKILL.md',
      'opencode.jsonc',
    ])
    expect(fileRecords.some((f) => f.path === 'opencode.jsonc')).toBe(true)
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

  it('adds orchestration guidance file when orchestrator is enabled', async () => {
    const opencodeTarget = makeTempDir('ai-setup-opencode-orchestrator-')

    await new OpenCodeAdapter().install({
      targetDir: opencodeTarget,
      libraryDir,
      fileRecords: [],
      force: true,
      enableServers: ['orchestrator'],
    })

    const opencodeOut = readFile(path.join(opencodeTarget, '.opencode/agents/orchestrator.md'))

    expect(opencodeOut).toContain('<!-- allowed-tools: list_catalog, compose_agent, start_chain, advance_chain, get_status, get_budget, retry_step, escalate_step, handoff -->')
    expect(opencodeOut).toContain('<!-- Recommended model: opus -->')
  })
})
