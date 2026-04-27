import { getPort } from '../config/paths.js'
import { findRunningServer } from '../daemon/discovery.js'
import { spawnDetachedServer } from '../daemon/detach.js'
import { startBridge } from '../transport/bridge.js'

export const CONNECT_HELP = `Usage: ai-setup-orchestrator connect [options]

Ensure the shared orchestrator HTTP server is running (starting it detached
if needed), then start a stdio MCP bridge to it.

This is the command to use in .mcp.json / opencode.jsonc so multiple
clients share a single server without blocking a terminal.

Options:
  --port <number>   Port to use when starting a new server (default: ${getPort()})
  -h, --help        Show this help
`

export interface ConnectOptions {
  port?: number
}

export function parseConnectArgs(args: string[]): ConnectOptions & { help: boolean } {
  let port: number | undefined
  let help = false

  for (let i = 0; i < args.length; i++) {
    const arg = args[i]
    if (arg === '--port' && args[i + 1]) {
      const parsed = parseInt(args[++i] ?? '', 10)
      if (!isNaN(parsed)) port = parsed
    } else if (arg === '-h' || arg === '--help') {
      help = true
    }
  }

  return { ...(port !== undefined ? { port } : {}), help }
}

export async function runConnect(options: ConnectOptions = {}): Promise<void> {
  const preferredPort = options.port ?? getPort()

  let info = await findRunningServer()
  if (!info) {
    await spawnDetachedServer(preferredPort)
    info = await findRunningServer()
    if (!info) throw new Error(`Failed to start orchestrator server on port ${preferredPort}`)
  }

  await startBridge(`http://127.0.0.1:${info.port}/mcp`)
}
