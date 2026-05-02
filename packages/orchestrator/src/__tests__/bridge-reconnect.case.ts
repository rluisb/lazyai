import http from 'node:http'
import net from 'node:net'
import os from 'node:os'
import path from 'node:path'
import crypto from 'node:crypto'
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

async function waitForHealth(port: number, deadlineMs = 3000): Promise<void> {
  const deadline = Date.now() + deadlineMs
  while (Date.now() < deadline) {
    const ok = await new Promise<boolean>((resolve) => {
      const req = http.get(`http://127.0.0.1:${port}/health`, (res) => {
        res.resume()
        resolve(res.statusCode === 200)
      })
      req.on('error', () => resolve(false))
      req.setTimeout(200, () => { req.destroy(); resolve(false) })
    })
    if (ok) return
    await new Promise((r) => setTimeout(r, 50))
  }
  throw new Error(`server did not become healthy on port ${port}`)
}

describe('bridge session resilience', () => {
  let abort: AbortController

  beforeEach(() => { abort = new AbortController() })
  afterEach(() => { abort.abort() })

  it('re-initializes upstream after daemon restart loses the session', async () => {
    const port = await getFreePort()
    const lockPath = path.join(os.tmpdir(), `orchestrator-bridge-test-${port}.lock`)

    // First daemon
    const serve1 = runServe({ port, lockPath, projectRoot: process.cwd() }, abort.signal)
    await waitForHealth(port)

    // Use the SDK Client directly with the same transport the bridge uses, so
    // we exercise the same session-id flow without spawning a stdio bridge.
    const { Client: ClientCtor } = await import('@modelcontextprotocol/sdk/client/index.js')
    const { StreamableHTTPClientTransport } = await import('@modelcontextprotocol/sdk/client/streamableHttp.js')

    const client = new ClientCtor({ name: 'test-bridge', version: '0.0.0' })
    const transport = new StreamableHTTPClientTransport(new URL(`http://127.0.0.1:${port}/mcp`))
    await client.connect(transport as never)

    // Sanity: tools/list works against the live session
    const first = await client.listTools()
    expect(Array.isArray(first.tools)).toBe(true)

    // Kill daemon — abort cleans up the listener; sessions vanish with the process
    abort.abort()
    await serve1.catch(() => { /* expected on abort */ })

    // Wait for port to free
    const freed = await new Promise<boolean>((resolve) => {
      const deadline = Date.now() + 3000
      const tick = (): void => {
        const probe = net.createConnection({ host: '127.0.0.1', port })
        probe.once('error', () => resolve(true))
        probe.once('connect', () => {
          probe.destroy()
          if (Date.now() > deadline) resolve(false)
          else setTimeout(tick, 50)
        })
      }
      tick()
    })
    expect(freed).toBe(true)

    // New daemon on same port
    const abort2 = new AbortController()
    const serve2 = runServe({ port, lockPath, projectRoot: process.cwd() }, abort2.signal)
    await waitForHealth(port)

    // The pre-existing client now has a stale mcp-session-id. Calling tools/list
    // against the new daemon should 404. The bridge reconnect logic lives in
    // startBridge → withReconnect; here we mirror it: catch the 404 and retry
    // with a fresh client. This proves the design: once reconnected, a new
    // session is minted and calls succeed.
    let sawSessionLost = false
    try {
      await client.listTools()
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      sawSessionLost = msg.includes('Session not found') || msg.includes('404') || msg.includes('Not connected')
    }
    expect(sawSessionLost).toBe(true)

    // Fresh client (what withReconnect does internally) must succeed
    const recovered = new ClientCtor({ name: 'test-bridge', version: '0.0.0' })
    const transport2 = new StreamableHTTPClientTransport(new URL(`http://127.0.0.1:${port}/mcp`))
    await recovered.connect(transport2 as never)
    const second = await recovered.listTools()
    expect(Array.isArray(second.tools)).toBe(true)

    await recovered.close().catch(() => { /* ignore */ })
    await client.close().catch(() => { /* ignore */ })
    abort2.abort()
    await serve2.catch(() => { /* expected on abort */ })
  }, 15000)

  it('returns 404 for unknown session-id (regression: prevents silent success)', async () => {
    const port = await getFreePort()
    const lockPath = path.join(os.tmpdir(), `orchestrator-404-test-${port}.lock`)
    const serve = runServe({ port, lockPath, projectRoot: process.cwd() }, abort.signal)
    await waitForHealth(port)

    const status = await new Promise<number>((resolve, reject) => {
      const req = http.request(
        {
          hostname: '127.0.0.1',
          port,
          path: '/mcp',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json, text/event-stream',
            'mcp-session-id': crypto.randomUUID(),
          },
        },
        (res) => { res.resume(); resolve(res.statusCode ?? 0) },
      )
      req.on('error', reject)
      req.end(JSON.stringify({ jsonrpc: '2.0', id: 1, method: 'tools/list', params: {} }))
    })
    expect(status).toBe(404)

    abort.abort()
    await serve.catch(() => { /* expected */ })
  }, 10000)
})
