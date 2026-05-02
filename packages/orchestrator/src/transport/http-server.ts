import crypto from 'node:crypto'
import http from 'node:http'
import { StreamableHTTPServerTransport } from '@modelcontextprotocol/sdk/server/streamableHttp.js'
import { createOrchestratorServer } from '../server.js'
import type { ToolHandlerOptions } from '../tool-handlers.js'
import type { Logger } from '../logging/logger.js'
import { createNoopLogger } from '../logging/logger.js'
import { ClientTracker } from '../daemon/client-tracker.js'
import { writeDiscovery, clearDiscovery } from '../daemon/discovery.js'
import { handoffActiveRuns } from '../persistence.js'

export interface HttpServerOptions extends ToolHandlerOptions {
  port: number
  logger?: Logger
  /** Write daemon.json on startup and clear it on shutdown. Default: false. */
  writeDiscoveryFile?: boolean
  /** Auto-shutdown when the last SSE client disconnects (after grace period). Default: false. */
  autoShutdown?: boolean
  /** Grace period in ms before auto-shutdown. Default: 5000. */
  gracePeriodMs?: number
  /** Idle TTL for an MCP session before it's GC'd. Sessions are touched on every
   *  request; bridges killed with SIGKILL never close gracefully, and without
   *  this their server-side transports leak until the daemon shuts down.
   *  Default: 600_000 (10 min). Set to 0 to disable. */
  sessionIdleTtlMs?: number
  /** How often to scan for idle sessions. Default: 60_000 (1 min). */
  sessionGcIntervalMs?: number
}

export interface HttpServerContext {
  server: http.Server
  port: number
  close(): Promise<void>
}

async function readBody(req: http.IncomingMessage): Promise<string> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = []
    req.on('data', (chunk: Buffer) => chunks.push(chunk))
    req.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')))
    req.on('error', reject)
  })
}

function isSseRequest(req: http.IncomingMessage): boolean {
  return req.method === 'GET' && (req.headers.accept ?? '').includes('text/event-stream')
}

export async function startHttpServer(options: HttpServerOptions): Promise<HttpServerContext> {
  const log = (options.logger ?? createNoopLogger()).child({ component: 'http-server' })

  // One transport+server pair per client session. Each initialize POST creates a
  // new entry; subsequent requests (GET SSE or POST with mcp-session-id) route to
  // the existing session. This is required because StreamableHTTPServerTransport
  // is not designed to handle concurrent clients on a single instance.
  interface SessionEntry {
    transport: StreamableHTTPServerTransport
    lastActivity: number
  }
  const sessions = new Map<string, SessionEntry>()

  async function routeRequest(req: http.IncomingMessage, res: http.ServerResponse, parsedBody: unknown): Promise<void> {
    const sessionId = req.headers['mcp-session-id'] as string | undefined

    if (sessionId) {
      const existing = sessions.get(sessionId)
      if (!existing) {
        res.writeHead(404, { 'Content-Type': 'application/json' })
        res.end(JSON.stringify({ error: 'Session not found' }))
        return
      }
      existing.lastActivity = Date.now()
      await existing.transport.handleRequest(req, res, parsedBody)
      return
    }

    // No session ID → new initialize request; create a dedicated transport+server pair
    const transport = new StreamableHTTPServerTransport({
      sessionIdGenerator: () => crypto.randomUUID(),
      onsessioninitialized: (id) => { sessions.set(id, { transport, lastActivity: Date.now() }) },
      onsessionclosed: (id) => { sessions.delete(id) },
    })

    const { server: mcpServer } = createOrchestratorServer(options)
    await mcpServer.connect(transport)
    await transport.handleRequest(req, res, parsedBody)
  }

  let tracker: ClientTracker | undefined
  let shutdownSignaled = false
  let gcTimer: ReturnType<typeof setInterval> | null = null

  const httpServer = http.createServer(async (req, res) => {
    const url = req.url ?? '/'

    if (url === '/health') {
      res.writeHead(200, { 'Content-Type': 'application/json' })
      res.end(JSON.stringify({ status: 'ok', pid: process.pid, clients: tracker?.count ?? 0 }))
      return
    }

    if (url !== '/mcp') {
      res.writeHead(404, { 'Content-Type': 'text/plain' })
      res.end('Not found — POST or GET /mcp')
      return
    }

    // Track SSE connections for auto-shutdown
    if (tracker && isSseRequest(req)) {
      tracker.track(res)
      log.info('http-server.client-connected', { clients: tracker.count })
      res.on('close', () => {
        log.info('http-server.client-disconnected', { clients: tracker?.count ?? 0 })
      })
    }

    try {
      let parsedBody: unknown
      if (req.method === 'POST') {
        const raw = await readBody(req)
        parsedBody = raw.length > 0 ? (JSON.parse(raw) as unknown) : undefined
      }
      await routeRequest(req, res, parsedBody)
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      log.error('request error', { error: msg })
      if (!res.headersSent) {
        res.writeHead(500, { 'Content-Type': 'application/json' })
        res.end(JSON.stringify({ error: msg }))
      }
    }
  })

  const boundPort = await new Promise<number>((resolve, reject) => {
    httpServer.listen(options.port, '127.0.0.1', () => {
      const addr = httpServer.address()
      const port = typeof addr === 'object' && addr !== null ? addr.port : options.port
      log.info('http-server.listening', { port })
      resolve(port)
    })
    httpServer.once('error', reject)
  })

  if (options.writeDiscoveryFile) {
    writeDiscovery(boundPort, process.pid)
  }

  const doClose = async (): Promise<void> => {
    if (shutdownSignaled) return
    shutdownSignaled = true
    tracker?.clear()
    if (options.writeDiscoveryFile) clearDiscovery()
    try {
      await handoffActiveRuns(options.projectRoot, log)
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      log.error('http-server.auto-handoff-failed', { error: msg })
    }
    if (gcTimer) { clearInterval(gcTimer); gcTimer = null }
    await Promise.all([...sessions.values()].map((s) => s.transport.close().catch(() => { /* best-effort */ })))
    sessions.clear()
    await new Promise<void>((resolve, reject) => {
      httpServer.close((err) => (err ? reject(err) : resolve()))
    })
    log.info('http-server.stopped')
  }

  const idleTtl = options.sessionIdleTtlMs ?? 600_000
  const gcInterval = options.sessionGcIntervalMs ?? 60_000
  if (idleTtl > 0) {
    gcTimer = setInterval(() => {
      const now = Date.now()
      for (const [id, entry] of sessions) {
        if (now - entry.lastActivity > idleTtl) {
          log.info('http-server.session-idle-gc', { sessionId: id, idleMs: now - entry.lastActivity })
          sessions.delete(id)
          entry.transport.close().catch(() => { /* best-effort */ })
        }
      }
    }, gcInterval)
    gcTimer.unref?.()
  }

  if (options.autoShutdown) {
    tracker = new ClientTracker({
      gracePeriodMs: options.gracePeriodMs ?? 5000,
      onShutdown: () => {
        log.info('http-server.auto-shutdown', { reason: 'no clients' })
        doClose().catch(() => { /* best-effort */ })
        // Give the close a moment then exit
        setTimeout(() => process.exit(0), 500)
      },
    })
  }

  return {
    server: httpServer,
    port: boundPort,
    close: doClose,
  }
}
