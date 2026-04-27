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

export async function startBridge(mcpUrl: string): Promise<void> {
  const upstream = new Client({ name: 'orchestrator-bridge', version: '0.1.0' })
  const httpTransport = new StreamableHTTPClientTransport(new URL(mcpUrl))
  await upstream.connect(httpTransport as unknown as Transport)

  // eslint-disable-next-line @typescript-eslint/no-deprecated
  const proxy = new Server(
    { name: 'ai-setup-orchestrator', version: '0.1.0' },
    { capabilities: { tools: {} } },
  )

  proxy.setRequestHandler(ListToolsRequestSchema, () => upstream.listTools())
  proxy.setRequestHandler(CallToolRequestSchema, (req: CallToolRequest) => upstream.callTool(req.params))

  await proxy.connect(new StdioServerTransport())

  await new Promise<void>((resolve) => {
    process.stdin.once('close', resolve)
    proxy.onclose = resolve
  })

  await upstream.close().catch(() => { /* best-effort */ })
}
