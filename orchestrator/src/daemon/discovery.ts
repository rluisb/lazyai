import fs from 'node:fs'
import http from 'node:http'
import path from 'node:path'
import { getDiscoveryFilePath } from '../config/paths.js'

export interface DaemonInfo {
  port: number
  pid: number
  startedAt: string
}

export function writeDiscovery(port: number, pid: number): void {
  const filePath = getDiscoveryFilePath()
  fs.mkdirSync(path.dirname(filePath), { recursive: true })
  const info: DaemonInfo = { port, pid, startedAt: new Date().toISOString() }
  fs.writeFileSync(filePath, JSON.stringify(info, null, 2), 'utf8')
}

export function readDiscovery(): DaemonInfo | null {
  try {
    const raw = fs.readFileSync(getDiscoveryFilePath(), 'utf8')
    return JSON.parse(raw) as DaemonInfo
  } catch {
    return null
  }
}

export function clearDiscovery(): void {
  try { fs.unlinkSync(getDiscoveryFilePath()) } catch { /* already gone */ }
}

export function pingServer(port: number, timeoutMs = 2000): Promise<boolean> {
  return new Promise((resolve) => {
    const req = http.get(`http://127.0.0.1:${port}/health`, (res) => {
      res.resume()
      resolve(res.statusCode === 200)
    })
    req.setTimeout(timeoutMs, () => { req.destroy(); resolve(false) })
    req.on('error', () => resolve(false))
  })
}

export async function findRunningServer(): Promise<DaemonInfo | null> {
  const info = readDiscovery()
  if (!info) return null
  const alive = await pingServer(info.port)
  return alive ? info : null
}
