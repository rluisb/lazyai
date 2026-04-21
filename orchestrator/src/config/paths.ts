import os from 'node:os'
import path from 'node:path'

const APP_NAME = 'ai-setup-orchestrator'

function xdgPath(envVar: string, fallbackUnderHome: string): string {
  const xdg = process.env[envVar]
  if (xdg && xdg.length > 0) return path.join(xdg, APP_NAME)
  return path.join(os.homedir(), fallbackUnderHome, APP_NAME)
}

export function getDataDir(): string {
  return xdgPath('XDG_DATA_HOME', path.join('.local', 'share'))
}

export function getStateDir(): string {
  return xdgPath('XDG_STATE_HOME', path.join('.local', 'state'))
}

export function getDatabasePath(): string {
  return path.join(getDataDir(), 'orchestrator.db')
}

export function getSocketPath(): string {
  return path.join(getDataDir(), 'orchestrator.sock')
}

export function getPort(): number {
  const env = process.env.AI_SETUP_ORCHESTRATOR_PORT
  if (env) {
    const parsed = parseInt(env, 10)
    if (!isNaN(parsed) && parsed > 0) return parsed
  }
  return 57372
}

export function getDiscoveryFilePath(): string {
  return path.join(getDataDir(), 'daemon.json')
}

export function getLockFilePath(): string {
  return path.join(getDataDir(), 'daemon.lock')
}

export function getLogDir(): string {
  return path.join(getStateDir(), 'logs')
}
