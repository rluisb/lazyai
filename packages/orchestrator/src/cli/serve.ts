import { getPort } from '../config/paths.js'
import { startHttpServer } from '../transport/http-server.js'
import { acquireLock, releaseLock } from '../daemon/lock.js'
import type { Logger } from '../logging/logger.js'
import type { ToolHandlerOptions } from '../tool-handlers.js'
import { getPersistenceDb } from '../persistence.js'
import { startQueueWorker, stopQueueWorker } from '../queue/worker.js'
import { registerBuiltinHandlers } from '../queue/handlers.js'

export const SERVE_HELP = `Usage: ai-setup-orchestrator serve [options]

Start the orchestrator as a shared HTTP server over TCP.
Multiple CLI clients (Claude Code, OpenCode, Codex) can connect simultaneously
via the MCP Streamable HTTP transport at http://127.0.0.1:<port>/mcp.

On startup a discovery file is written to the XDG data directory so other
processes can detect the running instance. The server shuts down automatically
after all SSE clients disconnect (5 s grace period).

Options:
  --port <number>    TCP port to listen on (default: ${getPort()}, env: AI_SETUP_ORCHESTRATOR_PORT)
  --project <path>   Project root to scope runs (default: cwd)
  --detach           Start the server in the background and free the terminal
  -h, --help         Show this help
`

export interface ServeOptions {
  port?: number
  projectRoot?: string
  logger?: Logger
  detach?: boolean
  /** Override the lock file path (useful in tests to avoid global lock contention). */
  lockPath?: string
}

export function parseServeArgs(args: string[]): ServeOptions & { help: boolean } {
  let port: number | undefined
  let projectRoot: string | undefined
  let detach = false
  let help = false

  for (let i = 0; i < args.length; i++) {
    const arg = args[i]
    if (arg === '--port' && args[i + 1]) {
      const parsed = parseInt(args[++i] ?? '', 10)
      if (!isNaN(parsed)) port = parsed
    } else if (arg === '--project' && args[i + 1]) {
      projectRoot = args[++i]
    } else if (arg === '--detach') {
      detach = true
    } else if (arg === '-h' || arg === '--help') {
      help = true
    }
  }

  return {
    ...(port !== undefined ? { port } : {}),
    ...(projectRoot !== undefined ? { projectRoot } : {}),
    detach,
    help,
  }
}

export async function runServe(options: ServeOptions = {}, signal?: AbortSignal): Promise<void> {
  const port = options.port ?? getPort()
  const projectRoot = options.projectRoot ?? process.cwd()
  const log = options.logger

  if (options.detach) {
    const { spawnDetachedServer } = await import('../daemon/detach.js')
    await spawnDetachedServer(port)
    process.stdout.write(`Orchestrator started at http://127.0.0.1:${port}/mcp\n`)
    return
  }

  if (!acquireLock(options.lockPath)) {
    throw new Error(
      `Another orchestrator process already holds the startup lock. ` +
      `If the previous process crashed, delete ${options.lockPath ?? (await import('../config/paths.js')).getLockFilePath()}.`,
    )
  }

  registerBuiltinHandlers(startQueueWorker({ db: getPersistenceDb(), ...(log ? { logger: log } : {}) }))

  const handlerOptions: ToolHandlerOptions = { projectRoot }

  const ctx = await startHttpServer({
    ...handlerOptions,
    port,
    writeDiscoveryFile: true,
    autoShutdown: true,
    ...(log ? { logger: log } : {}),
  })

  const stop = (): void => {
    stopQueueWorker()
    releaseLock(options.lockPath)
    ctx.close().catch((err: unknown) => {
      if (log) log.error('serve.close-error', { error: err instanceof Error ? err.message : String(err) })
    })
  }

  if (signal) signal.addEventListener('abort', stop, { once: true })
  process.once('SIGINT', stop)
  process.once('SIGTERM', stop)

  if (log) log.info('orchestrator.serve.ready', { port: ctx.port, url: `http://127.0.0.1:${ctx.port}/mcp` })

  await new Promise<void>((resolve) => {
    ctx.server.once('close', resolve)
    signal?.addEventListener('abort', () => resolve(), { once: true })
  })

  releaseLock(options.lockPath)
}
