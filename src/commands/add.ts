import type { Command } from 'commander'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import * as p from '@clack/prompts'
import type { ToolId, FileRecord } from '../types.js'
import { AdapterRegistry } from '../adapters/registry.js'
import { fileExists, resolveLibraryDir } from '../utils/files.js'
import { createStore, writeStore } from '../store/index.js'
import { Errors } from '../errors/index.js'

const libraryDir = resolveLibraryDir(dirname(fileURLToPath(import.meta.url)))

export function registerAdd(program: Command): void {
  program
    .command('add')
    .description('Add a tool to existing setup')
    .argument('<tool>', 'Tool to add (e.g. pi, opencode, claude-code, gemini, copilot)')
    .action(async (tool: string) => {
      const registry = new AdapterRegistry()

      if (!registry.getRegisteredIds().includes(tool)) {
        throw Errors.invalidInput(`unknown tool: ${tool}`, {
          available: registry.getRegisteredIds(),
        })
      }

      const toolId = tool as ToolId

      const targetDir = process.cwd()
      const configPath = join(targetDir, '.ai-setup.json')

      if (!fileExists(configPath)) {
        throw Errors.manifestNotFound(targetDir)
      }

      const db = await createStore(targetDir)
      const data = db.data

      if (data.config.tools.includes(toolId)) {
        p.log.info(`${toolId} is already installed.`)
        return
      }

      p.intro(`Adding ${toolId} to your setup...`)

      const s = p.spinner()
      s.start(`Installing ${toolId} files`)

      const adapter = registry.get(toolId)

      if (!adapter) {
        s.stop(`Failed to load adapter for ${toolId}`, 1)
        throw Errors.missingDependency(`adapter:${toolId}`)
      }

      const newFiles: FileRecord[] = []
      await adapter.install({
        targetDir,
        libraryDir,
        fileRecords: newFiles,
      })

      s.stop(`Installed ${toolId} files`)

      data.config.tools.push(toolId)
      const now = new Date().toISOString()
      data.files = [
        ...data.files,
        ...newFiles.map((file) => ({
          ...file,
          status: 'installed' as const,
          installedAt: now,
          lastCheckedAt: now,
        })),
      ]

      await writeStore(targetDir, data)

      p.outro(`✅ Successfully added ${toolId}!`)
    })
}
