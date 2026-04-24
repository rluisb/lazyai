import http from 'node:http'
import net from 'node:net'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { runServe } from '../cli/serve.js'

function getFreePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const server = net.createServer()
    server.listen(0, '127.0.0.1', () => {
      const addr = server.address()
      const port = typeof addr === 'object' && addr !== null ? addr.port : 0
      server.close((err) => (err ? reject(err) : resolve(port)))
    })
    server.once('error', reject)
  })
}

function postMcp(port: number, body: unknown): Promise<{ status: number; data: unknown }> {
  return new Promise((resolve, reject) => {
    const payload = JSON.stringify(body)
    const req = http.request(
      {
        hostname: '127.0.0.1',
        port,
        path: '/mcp',
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(payload),
          'Accept': 'application/json, text/event-stream',
        },
      },
      (res) => {
        const chunks: Buffer[] = []
        res.on('data', (c: Buffer) => chunks.push(c))
        res.on('end', () => {
          const raw = Buffer.concat(chunks).toString('utf8')
          try {
            resolve({ status: res.statusCode ?? 0, data: JSON.parse(raw) as unknown })
          } catch {
            const match = raw.match(/^data:\s*(.+)$/m)
            if (match) {
              try {
                resolve({ status: res.statusCode ?? 0, data: JSON.parse(match[1] ?? '') as unknown })
                return
              } catch { /* fall through */ }
            }
            resolve({ status: res.statusCode ?? 0, data: null })
          }
        })
      },
    )
    req.on('error', reject)
    req.write(payload)
    req.end()
  })
}

describe('HTTP/SSE transport', () => {
  let port: number
  let lockPath: string
  let abort: AbortController

  beforeEach(async () => {
    port = await getFreePort()
    lockPath = path.join(os.tmpdir(), `orchestrator-test-${port}.lock`)
    abort = new AbortController()
  })

  afterEach(() => {
    abort.abort()
  })

  it('responds 200 on /health', async () => {
    const serverReady = runServe({ port, lockPath, projectRoot: process.cwd() }, abort.signal)
    await new Promise((r) => setTimeout(r, 150))

    const result = await new Promise<{ status: number }>((resolve, reject) => {
      const req = http.request(
        { hostname: '127.0.0.1', port, path: '/health', method: 'GET' },
        (res) => { res.resume(); resolve({ status: res.statusCode ?? 0 }) },
      )
      req.on('error', reject)
      req.end()
    })
    expect(result.status).toBe(200)
    abort.abort()
    await serverReady
  })

  it('responds 404 on unknown path', async () => {
    const serverReady = runServe({ port, lockPath, projectRoot: process.cwd() }, abort.signal)
    await new Promise((r) => setTimeout(r, 150))

    const result = await new Promise<{ status: number }>((resolve, reject) => {
      const req = http.request(
        { hostname: '127.0.0.1', port, path: '/unknown', method: 'GET' },
        (res) => { res.resume(); resolve({ status: res.statusCode ?? 0 }) },
      )
      req.on('error', reject)
      req.end()
    })
    expect(result.status).toBe(404)
    abort.abort()
    await serverReady
  })

  it('handles MCP initialize request', async () => {
    const serverReady = runServe({ port, lockPath, projectRoot: process.cwd() }, abort.signal)
    await new Promise((r) => setTimeout(r, 150))

    const { status, data } = await postMcp(port, {
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2024-11-05',
        capabilities: {},
        clientInfo: { name: 'test-client', version: '0.0.1' },
      },
    })

    expect(status).toBe(200)
    const resp = data as { result?: { protocolVersion?: string } }
    expect(resp.result?.protocolVersion).toBeDefined()

    abort.abort()
    await serverReady
  })
})
