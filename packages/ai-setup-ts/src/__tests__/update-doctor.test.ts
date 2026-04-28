import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, } from 'vitest'
import { createProgram } from '../cli.js'
import type { AiSetupConfig, SetupConfig } from '../types.js'
import { runWizard } from '../wizard/index.js'

describe('update and doctor commands', () => {
  let originalCwd: string

  beforeEach(() => {
    originalCwd = process.cwd()
  })

  afterEach(() => {
    process.chdir(originalCwd)
  })

  async function setupProject(): Promise<{ tempDir: string; config: AiSetupConfig }> {
    const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-test-'))
    fs.mkdirSync(path.join(tempDir, '.git'), { recursive: true })
    process.chdir(tempDir)

    const config: SetupConfig = {
      setupScope: 'project',
      setupType: 'project',
      tools: ['opencode', 'claude-code'],
      projectName: 'test-project',
      targetDir: tempDir,
    }

    await runWizard({
      interactive: false,
      cliOverrides: {
        scope: config.setupScope,
        tools: config.tools,
        name: config.projectName,
      },
      targetDir: config.targetDir,
      ...(config.force !== undefined ? { force: config.force } : {}),
    })

    const aiSetup = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as AiSetupConfig
    return { tempDir, config: aiSetup }
  }

  it('update --force backs up customized files and adds missing untracked expected files', async () => {
    const { tempDir, config } = await setupProject()

    expect(config.files.length).toBeGreaterThan(2)

    // The `update` command only restores files it knows how to rebuild from the
    // library (templates/, rules/, root files, agents, skills, prompts). Pick
    // tracked files whose paths fall under that coverage so the restore path
    // is exercised end-to-end.
    const isManagedByUpdate = (relPath: string): boolean =>
      relPath.startsWith('specs/templates/')
      || relPath.startsWith('specs/rules/')
      || relPath.startsWith('.opencode/agents/')
      || relPath.startsWith('.claude/agents/')
    const managedFiles = config.files.filter((f) => isManagedByUpdate(f.path))
    const customized = managedFiles[0]
    const reAddTarget = managedFiles[1]
    if (!customized || !reAddTarget) {
      throw new Error('Expected scaffold to create files managed by update')
    }

    const customizedPath = path.join(tempDir, customized.path)
    const originalCustomizedContent = fs.readFileSync(customizedPath, 'utf-8')
    fs.writeFileSync(customizedPath, `${fs.readFileSync(customizedPath, 'utf-8')}\ncustom change\n`, 'utf-8')

    const reduced = {
      ...config,
      files: config.files.filter((f) => f.path !== reAddTarget.path),
    }
    fs.unlinkSync(path.join(tempDir, reAddTarget.path))
    fs.writeFileSync(path.join(tempDir, '.ai-setup.json'), JSON.stringify(reduced, null, 2), 'utf-8')

    const program = createProgram()
    await program.parseAsync(['node', 'ai-setup', 'update', '--force'])

    const nextConfig = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as AiSetupConfig

    expect(fs.readFileSync(customizedPath, 'utf-8')).toBe(originalCustomizedContent)
    expect(nextConfig.files.find((f) => f.path === reAddTarget.path)).toBeTruthy()
    expect(fs.existsSync(path.join(tempDir, reAddTarget.path))).toBe(true)

    const backupPath = path.join(tempDir, '.ai-setup-backup', customized.path)
    expect(fs.existsSync(backupPath)).toBe(true)
    expect(fs.readFileSync(backupPath, 'utf-8')).toContain('custom change')
  })

  it('doctor exits non-zero when tracked files are modified', async () => {
    const { tempDir, config } = await setupProject()

    expect(config.files.length).toBeGreaterThan(0)

    const target = config.files[0]
    if (!target) {
      throw new Error('Expected at least one tracked file')
    }

    const targetPath = path.join(tempDir, target.path)
    fs.writeFileSync(targetPath, `${fs.readFileSync(targetPath, 'utf-8')}\nmutated\n`, 'utf-8')

    const program = createProgram()
    await expect(program.parseAsync(['node', 'ai-setup', 'doctor'])).rejects.toThrow('Doctor found 1 integrity issue(s)')
  })
})
