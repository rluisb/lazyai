import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { writeToCanonical } from '../migration/canonical-writer.js'
import type { ParsedSetup } from '../migration/types.js'
import type { FileRecord } from '../types.js'

function createParsedSetup(): ParsedSetup {
  return {
    projectName: 'Test Project',
    description: 'migration test',
    agents: [
      {
        id: 'custom-agent',
        name: 'Custom Agent',
        description: 'agent',
        role: 'custom',
        content: '# Custom Agent\n\nPreserved content',
        sourcePath: '.opencode/agents/custom-agent.md',
      },
    ],
    rules: [],
    commands: [
      {
        id: 'custom-command',
        name: 'Custom Command',
        description: 'command',
        content: '# Custom Command\n\nDo custom work',
        sourcePath: '.opencode/commands/custom-command.md',
      },
    ],
    templates: [],
    customSections: [],
    files: [],
    metadata: { adapter: 'opencode' },
  }
}

describe('writeToCanonical', () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'ai-setup-canonical-writer-'))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('writes parsed agents and commands into .ai canonical directories', async () => {
    const fileRecords: FileRecord[] = []
    const result = await writeToCanonical({
      targetDir: tempDir,
      parsedSetup: createParsedSetup(),
      fileRecords,
    })

    expect(result.agents).toContain('.ai/agents/custom-agent.md')
    expect(result.skills).toContain('.ai/skills/custom-command.md')
    expect(existsSync(path.join(tempDir, '.ai', 'agents', 'custom-agent.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'skills', 'custom-command.md'))).toBe(true)
  })

  it('supports dry-run mode without writing files', async () => {
    const fileRecords: FileRecord[] = []
    const result = await writeToCanonical({
      targetDir: tempDir,
      parsedSetup: createParsedSetup(),
      fileRecords,
      dryRun: true,
    })

    expect(result.agents).toContain('.ai/agents/custom-agent.md')
    expect(result.skills).toContain('.ai/skills/custom-command.md')
    expect(existsSync(path.join(tempDir, '.ai', 'agents', 'custom-agent.md'))).toBe(false)
    expect(existsSync(path.join(tempDir, '.ai', 'skills', 'custom-command.md'))).toBe(false)
    expect(fileRecords).toHaveLength(0)
  })

  it('tracks fileRecords with migrated source metadata', async () => {
    const fileRecords: FileRecord[] = []
    await writeToCanonical({
      targetDir: tempDir,
      parsedSetup: createParsedSetup(),
      fileRecords,
    })

    const agentRecord = fileRecords.find((record) => record.path === '.ai/agents/custom-agent.md')
    const commandRecord = fileRecords.find((record) => record.path === '.ai/skills/custom-command.md')

    expect(agentRecord?.source).toBe('migrated:.opencode/agents/custom-agent.md')
    expect(commandRecord?.source).toBe('migrated:.opencode/commands/custom-command.md')
    expect(agentRecord?.hash).toMatch(/^[a-f0-9]{16}$/)

    const agentContent = readFileSync(path.join(tempDir, '.ai', 'agents', 'custom-agent.md'), 'utf-8')
    expect(agentContent).toContain('Preserved content')
  })
})
