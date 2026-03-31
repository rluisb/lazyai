import type { Command } from 'commander'
import { readFileSync, writeFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import * as p from '@clack/prompts'
import type { ToolId, AiSetupConfig, FileRecord } from '../types.js'
import { AdapterRegistry } from '../adapters/registry.js'
import { fileExists, resolveLibraryDir } from '../utils/files.js'

const libraryDir = resolveLibraryDir(dirname(fileURLToPath(import.meta.url)))

export function registerAdd(program: Command): void {
  program
    .command('add')
    .description('Add a tool to existing setup')
    .argument('<tool>', 'Tool to add (e.g. pi, opencode, claude-code, gemini, copilot)')
    .action(async (tool: string) => {
      const registry = new AdapterRegistry()

      if (!registry.getRegisteredIds().includes(tool)) {
        p.log.error(`Unknown tool: ${tool}. Available tools: ${registry.getRegisteredIds().join(', ')}`)
        process.exit(1)
      }

      const toolId = tool as ToolId

      const targetDir = process.cwd()
      const configPath = join(targetDir, '.ai-setup.json')

      if (!fileExists(configPath)) {
        p.log.error('No .ai-setup.json found. Please run init first.')
        process.exit(1)
      }

      const configStr = readFileSync(configPath, 'utf-8')
      const config = JSON.parse(configStr) as AiSetupConfig

      if (config.tools.includes(toolId)) {
        p.log.info(`${toolId} is already installed.`)
        return
      }

      p.intro(`Adding ${toolId} to your setup...`)

      const s = p.spinner()
      s.start(`Installing ${toolId} files`)

      const adapter = registry.get(toolId)

      if (!adapter) {
        s.stop(`Failed to load adapter for ${toolId}`, 1)
        process.exit(1)
      }

      const newFiles: FileRecord[] = []
      await adapter.install({
        targetDir,
        libraryDir,
        fileRecords: newFiles,
      })

      s.stop(`Installed ${toolId} files`)

      config.tools.push(toolId)
      config.files = [...config.files, ...newFiles]

      writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8')

      p.outro(`✅ Successfully added ${toolId}!`)
    })
}
