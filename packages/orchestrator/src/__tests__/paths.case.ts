import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import {
  getDataDir,
  getDatabasePath,
  getDiscoveryFilePath,
  getLockFilePath,
  getLogDir,
  getPort,
  getSocketPath,
  getStateDir,
} from '../config/paths.js'

describe('config paths', () => {
  let originalData: string | undefined
  let originalState: string | undefined
  let originalPort: string | undefined

  beforeEach(() => {
    originalData = process.env.XDG_DATA_HOME
    originalState = process.env.XDG_STATE_HOME
    originalPort = process.env.AI_SETUP_ORCHESTRATOR_PORT
  })

  afterEach(() => {
    if (originalData === undefined) delete process.env.XDG_DATA_HOME
    else process.env.XDG_DATA_HOME = originalData
    if (originalState === undefined) delete process.env.XDG_STATE_HOME
    else process.env.XDG_STATE_HOME = originalState
    if (originalPort === undefined) delete process.env.AI_SETUP_ORCHESTRATOR_PORT
    else process.env.AI_SETUP_ORCHESTRATOR_PORT = originalPort
  })

  it('uses XDG_DATA_HOME when set', () => {
    process.env.XDG_DATA_HOME = '/tmp/xdg-data'
    expect(getDataDir()).toBe(path.join('/tmp/xdg-data', 'ai-setup-orchestrator'))
    expect(getDatabasePath()).toBe(
      path.join('/tmp/xdg-data', 'ai-setup-orchestrator', 'orchestrator.db'),
    )
    expect(getSocketPath()).toBe(
      path.join('/tmp/xdg-data', 'ai-setup-orchestrator', 'orchestrator.sock'),
    )
    expect(getDiscoveryFilePath()).toBe(
      path.join('/tmp/xdg-data', 'ai-setup-orchestrator', 'daemon.json'),
    )
    expect(getLockFilePath()).toBe(
      path.join('/tmp/xdg-data', 'ai-setup-orchestrator', 'daemon.lock'),
    )
  })

  it('uses XDG_STATE_HOME when set', () => {
    process.env.XDG_STATE_HOME = '/tmp/xdg-state'
    expect(getStateDir()).toBe(path.join('/tmp/xdg-state', 'ai-setup-orchestrator'))
    expect(getLogDir()).toBe(path.join('/tmp/xdg-state', 'ai-setup-orchestrator', 'logs'))
  })

  it('falls back to ~/.local/share and ~/.local/state when XDG vars are unset', () => {
    delete process.env.XDG_DATA_HOME
    delete process.env.XDG_STATE_HOME
    expect(getDataDir()).toMatch(/[/\\]\.local[/\\]share[/\\]ai-setup-orchestrator$/)
    expect(getStateDir()).toMatch(/[/\\]\.local[/\\]state[/\\]ai-setup-orchestrator$/)
  })

  it('getPort returns default 57372 when env is unset', () => {
    delete process.env.AI_SETUP_ORCHESTRATOR_PORT
    expect(getPort()).toBe(57372)
  })

  it('getPort respects AI_SETUP_ORCHESTRATOR_PORT env var', () => {
    process.env.AI_SETUP_ORCHESTRATOR_PORT = '9999'
    expect(getPort()).toBe(9999)
  })
})
