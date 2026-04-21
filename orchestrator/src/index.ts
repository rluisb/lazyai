#!/usr/bin/env node

import { pathToFileURL } from 'node:url'
import { realpathSync } from 'node:fs'
import { parseTailArgs, runTail, TAIL_HELP } from './cli/tail.js'
import { runCatalog } from './cli/catalog.js'
import { parseServeArgs, runServe, SERVE_HELP } from './cli/serve.js'
import { parseConnectArgs, runConnect, CONNECT_HELP } from './cli/connect.js'
import { runInvoke } from './cli/invoke.js'
import { getPersistenceDb } from './persistence.js'
import { getLogDir, getPort } from './config/paths.js'
import { createLogger, type Logger } from './logging/logger.js'
import { createFileSink } from './logging/sink.js'
import { findRunningServer } from './daemon/discovery.js'

export * from './budget-tracker.js'
export * from './bootstrap.js'
export * from './chain-machine.js'
export * from './compiler.js'
export { composeAgent } from './composer.js'
export type { ComposeAgentInput as ComposePromptInput } from './composer.js'
export * from './error-journal.js'
export * from './housekeeping.js'
export * from './loader.js'
export * from './logging/logger.js'
export * from './logging/sink.js'
export * from './persistence.js'
export * from './server.js'
export * from './tool-handlers.js'
export * from './events/bus.js'
export * from './queue/queue.js'
export * from './queue/worker.js'
export * from './types.js'
export * from './daemon/discovery.js'
export * from './daemon/lock.js'
export * from './daemon/client-tracker.js'

function buildDefaultLogger(): Logger {
  const sink = createFileSink({ dir: getLogDir() })
  const minLevel = (process.env.AI_SETUP_LOG_LEVEL as 'trace' | 'debug' | 'info' | 'warn' | 'error' | undefined) ?? 'info'
  return createLogger({ sink, minLevel, bindings: { pid: process.pid } })
}

const START_HELP = `Usage: ai-setup-orchestrator start

Discover a running orchestrator instance and print its URL, or start one
if none is found. Exits 0 in both cases.

This is the recommended way to ensure a single shared server is running
before pointing MCP clients at http://127.0.0.1:${getPort()}/mcp.
`

async function main(): Promise<void> {
  const [, , subcommand, ...rest] = process.argv

  if (subcommand === 'catalog') {
    await runCatalog(getPersistenceDb(), rest)
    return
  }

  if (subcommand === 'invoke') {
    await runInvoke(getPersistenceDb(), rest)
    return
  }

  if (subcommand === 'connect') {
    const parsed = parseConnectArgs(rest)
    if (parsed.help) { process.stdout.write(CONNECT_HELP); return }
    await runConnect(parsed)
    return
  }

  if (subcommand === 'serve') {
    const parsed = parseServeArgs(rest)
    if (parsed.help) { process.stdout.write(SERVE_HELP); return }
    const logger = buildDefaultLogger()
    logger.info('orchestrator.serve', { transport: 'http', port: parsed.port ?? getPort() })
    await runServe({ ...parsed, logger })
    return
  }

  if (subcommand === 'start') {
    if (rest.includes('-h') || rest.includes('--help')) {
      process.stdout.write(START_HELP)
      return
    }
    const existing = await findRunningServer()
    if (existing) {
      process.stdout.write(`Orchestrator already running at http://127.0.0.1:${existing.port}/mcp (pid ${existing.pid})\n`)
      return
    }
    const logger = buildDefaultLogger()
    const port = getPort()
    logger.info('orchestrator.start', { transport: 'http', port })
    process.stdout.write(`Starting orchestrator at http://127.0.0.1:${port}/mcp\n`)
    await runServe({ port, logger })
    return
  }

  if (subcommand === 'tail') {
    const parsed = parseTailArgs(rest)
    if (parsed.help) {
      process.stdout.write(TAIL_HELP)
      return
    }
    await runTail({
      filters: parsed.filters,
      follow: parsed.follow,
      fromStart: parsed.fromStart,
      ...(parsed.date ? { date: parsed.date } : {}),
    })
    return
  }

  // Default: print status / usage hint
  const existing = await findRunningServer()
  if (existing) {
    process.stdout.write(
      `Orchestrator is running at http://127.0.0.1:${existing.port}/mcp (pid ${existing.pid})\n` +
      `Point your MCP client at that URL or run \`ai-setup-orchestrator tail\` to stream logs.\n`,
    )
  } else {
    process.stdout.write(
      `Orchestrator is not running.\n` +
      `Run \`ai-setup-orchestrator serve\` or \`ai-setup-orchestrator start\` to start it.\n` +
      `Then configure your MCP client:\n` +
      `  { "type": "http", "url": "http://127.0.0.1:${getPort()}/mcp" }\n`,
    )
  }
}

function isEntrypoint(): boolean {
  const entryArg = process.argv[1]
  if (!entryArg) return false
  try {
    return import.meta.url === pathToFileURL(realpathSync(entryArg)).href
  } catch {
    return import.meta.url === pathToFileURL(entryArg).href
  }
}

if (isEntrypoint()) {
  main().catch((error: unknown) => {
    const message = error instanceof Error ? (error.stack ?? error.message) : String(error)
    console.error(message)
    process.exit(1)
  })
}
