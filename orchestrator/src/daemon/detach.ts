import { spawn } from 'node:child_process'
import { pingServer } from './discovery.js'

export async function spawnDetachedServer(port: number): Promise<void> {
  const child = spawn(
    process.execPath,
    [process.argv[1]!, 'serve', '--port', String(port)],
    { detached: true, stdio: 'ignore', env: { ...process.env } },
  )
  child.unref()

  const deadline = Date.now() + 5000
  while (Date.now() < deadline) {
    await new Promise<void>((r) => setTimeout(r, 150))
    if (await pingServer(port)) return
  }
  throw new Error(`Orchestrator did not become ready on port ${port} within 5s`)
}
