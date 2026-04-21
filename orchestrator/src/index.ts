#!/usr/bin/env node

import { pathToFileURL } from 'node:url'
import { startStdioServer } from './server.js'

export * from './budget-tracker.js'
export * from './bootstrap.js'
export * from './chain-machine.js'
export * from './compiler.js'
export { composeAgent } from './composer.js'
export type { ComposeAgentInput as ComposePromptInput } from './composer.js'
export * from './error-journal.js'
export * from './housekeeping.js'
export * from './loader.js'
export * from './persistence.js'
export * from './server.js'
export * from './tool-handlers.js'
export * from './types.js'

async function main(): Promise<void> {
  await startStdioServer({
    projectRoot: process.cwd(),
  })
}

function isEntrypoint(): boolean {
  const entryArg = process.argv[1]
  if (!entryArg) return false
  return import.meta.url === pathToFileURL(entryArg).href
}

if (isEntrypoint()) {
  main().catch((error: unknown) => {
    const message = error instanceof Error ? error.stack ?? error.message : String(error)
    console.error(message)
    process.exit(1)
  })
}
