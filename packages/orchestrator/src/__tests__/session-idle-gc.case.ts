import http from 'node:http'
import net from 'node:net'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { startHttpServer } from '../transport/http-server.js'

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

function postInit(port: number): Promise<{ status: number; sessionId: string | undefined }> {
  return new Promise((resolve, reject) => {
    const body = JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2024-11-05',
        capabilities: {},
        clientInfo: { name: 'idle-gc-test', version: '0.0.0' },
      },
    })
    const req = http.request(
      {
        hostname: '127.0.0.1', port, path: '/mcp', method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(body),
          'Accept': 'application/json, text/event-stream',
        },
      },
      (res) => {
        res.resume()
        const sid = res.headers['mcp-session-id']
        resolve({
          status: res.statusCode ?? 0,
          sessionId: typeof sid === 'string' ? sid : undefined,
        })
      },
    )
    req.on('error', reject)
    req.end(body)
  })
}

function postWithSession(port: number, sessionId: string): Promise<number> {
  return new Promise((resolve, reject) => {
    const body = JSON.stringify({ jsonrpc: '2.0', id: 2, method: 'tools/list', params: {} })
    const req = http.request(
      {
        hostname: '127.0.0.1', port, path: '/mcp', method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(body),
          'Accept': 'application/json, text/event-stream',
          'mcp-session-id': sessionId,
        },
      },
      (res) => { res.resume(); resolve(res.statusCode ?? 0) },
    )
    req.on('error', reject)
    req.end(body)
  })
}

describe('session idle GC', () => {
  let ctx: Awaited<ReturnType<typeof startHttpServer>> | null = null

  beforeEach(() => { ctx = null })
  afterEach(async () => { if (ctx) await ctx.close() })

  it('GCs sessions whose last activity exceeds the idle TTL', async () => {
    const port = await getFreePort()
    ctx = await startHttpServer({
      projectRoot: process.cwd(),
      port,
      sessionIdleTtlMs: 50,
      sessionGcIntervalMs: 25,
    })

    const init = await postInit(port)
    expect(init.status).toBe(200)
    expect(init.sessionId).toBeDefined()

    // Immediate reuse works
    const ok = await postWithSession(port, init.sessionId!)
    expect(ok).toBe(200)

    // Wait for idle TTL + a GC tick
    await new Promise((r) => setTimeout(r, 150))

    // Session should be gone — server returns 404
    const status = await postWithSession(port, init.sessionId!)
    expect(status).toBe(404)
  }, 5000)

  it('does not GC active sessions', async () => {
    const port = await getFreePort()
    ctx = await startHttpServer({
      projectRoot: process.cwd(),
      port,
      sessionIdleTtlMs: 200,
      sessionGcIntervalMs: 25,
    })

    const init = await postInit(port)
    expect(init.sessionId).toBeDefined()

    // Activity every 50ms keeps the session alive past the 200ms TTL
    for (let i = 0; i < 8; i++) {
      const status = await postWithSession(port, init.sessionId!)
      expect(status).toBe(200)
      await new Promise((r) => setTimeout(r, 50))
    }
  }, 5000)
})
