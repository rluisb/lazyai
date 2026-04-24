import { chmodSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { type CmdRunner, validateOpenCodeInstall } from '../adapters/opencode-validate.js'

describe('validateOpenCodeInstall', () => {
  let tempDirs: string[]
  let originalPath: string | undefined

  let originalSkip: string | undefined

  beforeEach(() => {
    tempDirs = []
    originalPath = process.env.PATH
    originalSkip = process.env.AI_SETUP_SKIP_VALIDATION
    delete process.env.AI_SETUP_SKIP_VALIDATION
  })

  afterEach(() => {
    for (const dir of tempDirs) {
      rmSync(dir, { recursive: true, force: true })
    }
    if (originalPath === undefined) {
      delete process.env.PATH
    } else {
      process.env.PATH = originalPath
    }
    if (originalSkip === undefined) {
      delete process.env.AI_SETUP_SKIP_VALIDATION
    } else {
      process.env.AI_SETUP_SKIP_VALIDATION = originalSkip
    }
  })

  const makeTempDir = (prefix: string): string => {
    const dir = mkdtempSync(path.join(tmpdir(), prefix))
    tempDirs.push(dir)
    return dir
  }

  const stubOpenCodeOnPath = (): void => {
    const binDir = makeTempDir('ai-setup-validate-bin-')
    const fakeBin = path.join(binDir, 'opencode')
    writeFileSync(fakeBin, '#!/bin/sh\nexit 0\n')
    chmodSync(fakeBin, 0o755)
    process.env.PATH = binDir
  }

  it('returns no warnings when the opencode binary is absent', () => {
    process.env.PATH = ''
    const ocDir = makeTempDir('ai-setup-validate-ocdir-')

    const stubRunner: CmdRunner = () => {
      throw new Error('unexpected call — binary should be absent')
    }

    const warnings = validateOpenCodeInstall({ ocDir }, stubRunner)
    expect(warnings).toEqual([])
  })

  it('emits a config warning when opencode debug config fails', () => {
    stubOpenCodeOnPath()
    const ocDir = makeTempDir('ai-setup-validate-ocdir-')

    const stubRunner: CmdRunner = (_cmd, args) => {
      if (args[0] === 'debug' && args[1] === 'config') {
        throw new Error('config parse error')
      }
      return 'ok'
    }

    const warnings = validateOpenCodeInstall({ ocDir }, stubRunner)
    expect(warnings.some((w) => w.scope === 'config')).toBe(true)
  })

  it('emits an agent warning when opencode debug agent fails for an installed agent', () => {
    stubOpenCodeOnPath()
    const ocDir = makeTempDir('ai-setup-validate-ocdir-')
    const agentsDir = path.join(ocDir, 'agents')
    mkdirSync(agentsDir, { recursive: true })
    writeFileSync(path.join(agentsDir, 'builder.md'), '# Builder')

    const stubRunner: CmdRunner = (_cmd, args) => {
      if (args[0] === 'debug' && args[1] === 'agent') {
        throw new Error('parse error in frontmatter')
      }
      return '{"mcp": {}}'
    }

    const warnings = validateOpenCodeInstall({ ocDir }, stubRunner)
    expect(warnings.some((w) => w.scope === 'agent' && w.item === 'builder')).toBe(true)
  })

  it('returns no warnings when config and all agents pass', () => {
    stubOpenCodeOnPath()
    const ocDir = makeTempDir('ai-setup-validate-ocdir-')
    const agentsDir = path.join(ocDir, 'agents')
    mkdirSync(agentsDir, { recursive: true })
    writeFileSync(path.join(agentsDir, 'builder.md'), '# Builder')

    const stubRunner: CmdRunner = () => '{"mcp": {"servers": {}}}'

    const warnings = validateOpenCodeInstall({ ocDir }, stubRunner)
    expect(warnings).toEqual([])
  })
})
