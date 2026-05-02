import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { Client } from '@modelcontextprotocol/sdk/client/index.js'
import { StreamableHTTPClientTransport } from '@modelcontextprotocol/sdk/client/streamableHttp.js'
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  type CallToolRequest,
} from '@modelcontextprotocol/sdk/types.js'
import type { Transport } from '@modelcontextprotocol/sdk/shared/transport.js'
// Server is the correct low-level class for a proxy — McpServer only declares
// the tools capability after registerTool() is called, which a transparent
// proxy cannot do ahead of time.
// eslint-disable-next-line @typescript-eslint/no-deprecated
import { Server } from '@modelcontextprotocol/sdk/server/index.js'

// Errors that mean "your session is gone, re-handshake": HTTP 404 from the
// daemon's session router (`Session not found`), or transport-level closure.
// Matching by message string because the SDK doesn't surface a typed code for
// these — keep the patterns narrow so we don't retry real protocol errors.
function isSessionLostError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err)
  return (
    msg.includes('Session not found') ||
    msg.includes('HTTP 404') ||
    /\b404\b/.test(msg) ||
    msg.includes('Not connected') ||
    msg.includes('Transport is closed')
  )
}

export async function startBridge(mcpUrl: string): Promise<void> {
  const url = new URL(mcpUrl)
  let upstream: Client | null = null

  async function connectUpstream(): Promise<Client> {
    const client = new Client({ name: 'orchestrator-bridge', version: '0.1.0' })
    const httpTransport = new StreamableHTTPClientTransport(url)
    await client.connect(httpTransport as unknown as Transport)
    return client
  }

  async function getUpstream(): Promise<Client> {
    if (upstream) return upstream
    upstream = await connectUpstream()
    return upstream
  }

  async function reconnectUpstream(): Promise<Client> {
    const old = upstream
    upstream = null
    if (old) await old.close().catch(() => { /* best-effort */ })
    upstream = await connectUpstream()
    return upstream
  }

  // Run an upstream call, transparently re-initializing once if the daemon
  // reports our session is gone. One retry only — repeated failures are real
  // and should bubble to the caller as MCP errors.
  async function withReconnect<T>(fn: (c: Client) => Promise<T>): Promise<T> {
    try {
      return await fn(await getUpstream())
    } catch (err) {
      if (!isSessionLostError(err)) throw err
      const fresh = await reconnectUpstream()
      return await fn(fresh)
    }
  }

  upstream = await connectUpstream()

  const proxy =
    // eslint-disable-next-line @typescript-eslint/no-deprecated
    new Server(
      { name: 'ai-setup-orchestrator', version: '0.1.0' },
      { capabilities: { tools: {} } },
    )

  proxy.setRequestHandler(ListToolsRequestSchema, () => withReconnect((c) => c.listTools()))
  proxy.setRequestHandler(CallToolRequestSchema, (req: CallToolRequest) =>
    withReconnect((c) => c.callTool(req.params)),
  )

  await proxy.connect(new StdioServerTransport())

  await new Promise<void>((resolve) => {
    process.stdin.once('close', resolve)
    proxy.onclose = resolve
  })

  if (upstream) await upstream.close().catch(() => { /* best-effort */ })
}
