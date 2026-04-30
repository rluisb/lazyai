import { execFileSync } from 'node:child_process'
import { homedir } from 'node:os'
import path from 'node:path'
import type { PromptId, SkillId } from '../types.js'
import * as files from '../utils/files.js'
import { stripYamlFrontmatter } from '../utils/frontmatter.js'
import {
  copyLibraryDirectory,
  copyWithRecord,
  getOrchestratorPromptContent,
  installRootTemplateIfMissing,
  isOrchestratorEnabled,
  writeContentWithRecord,
} from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

export class CopilotAdapter implements ToolAdapter {
  getToolId(): string {
    return 'copilot'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    const homeDir = ctx.homeDir ?? homedir()
    const githubDir = isGlobal ? path.join(homeDir, '.copilot') : path.join(ctx.targetDir, '.github')

    if (isGlobal && !copilotProbePasses(homeDir)) {
      console.log('🤖  Skipping GitHub Copilot global install (CLI/home config not detected)...')
      return
    }

    files.ensureDir(githubDir)

    const agentsDir = path.join(githubDir, 'agents')
    files.ensureDir(agentsDir)
    const instructionsDir = path.join(githubDir, 'instructions')
    files.ensureDir(instructionsDir)
    const promptsDir = path.join(githubDir, 'prompts')
    files.ensureDir(promptsDir)
    const chatmodesDir = path.join(githubDir, 'chatmodes')
    files.ensureDir(chatmodesDir)

    console.log('🤖  Installing GitHub Copilot tools...')

    const selectedSkills = ctx.selections?.skills ? new Set(ctx.selections.skills) : undefined
    const selectedPrompts = ctx.selections?.prompts ? new Set(ctx.selections.prompts) : undefined

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'copilot/agents',
      toDestPath: file => path.join(agentsDir, path.basename(file)),
      includeFile: file => file.endsWith('.agent.yaml') && (!isOrchestratorEnabled(ctx) || file !== 'orchestrator.agent.yaml'),
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'copilot/instructions',
      toDestPath: file => path.join(instructionsDir, path.basename(file)),
      includeFile: file => file.endsWith('.instructions.md'),
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'chatmodes',
      toDestPath: file => path.join(chatmodesDir, path.basename(file)),
      includeFile: file => file.endsWith('.chatmode.md'),
    })

    // Prompts - Copilot prompt files use the .prompt.md suffix
    const promptTemplatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(promptTemplatesDir)) {
      const fileId = path.parse(file).name as PromptId
      if (selectedPrompts && !selectedPrompts.has(fileId)) continue
      const srcPath = path.join(promptTemplatesDir, file)
      if (files.isDirectory(srcPath)) continue
      const destFile = `${path.parse(file).name}.prompt.md`
      await copyWithRecord({
        src: srcPath,
        dest: path.join(promptsDir, destFile),
        ctx,
      })
    }

    // Skills - transformed into Copilot agent files
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      const src = path.join(skillsDir, file)
      if (files.isDirectory(src)) continue
      const fileId = path.parse(file).name as SkillId
      if (selectedSkills && !selectedSkills.has(fileId)) continue
      const parsed = path.parse(file)
      const destFile = `${parsed.name}.agent.yaml`
      const dest = path.join(agentsDir, destFile)
      await copyWithRecord({
        src,
        dest,
        ctx,
        transform: content => skillToAgentYaml(fileId, stripYamlFrontmatter(content)),
      })
    }

    if (isOrchestratorEnabled(ctx)) {
      await writeContentWithRecord({
        dest: path.join(agentsDir, 'orchestrator.agent.yaml'),
        content: skillToAgentYaml('orchestrator', stripYamlFrontmatter(getOrchestratorPromptContent(ctx))),
        ctx,
        source: 'generated:orchestrator-agent',
      })
    }

    if (!isGlobal) {
      await installRootTemplateIfMissing({
        ctx,
        recordPath: 'AGENTS.md',
        destPath: path.join(ctx.targetDir, 'AGENTS.md'),
        templateSource: 'root/AGENTS.template.md',
      })

      await installRootTemplateIfMissing({
        ctx,
        recordPath: '.github/copilot-instructions.md',
        destPath: path.join(githubDir, 'copilot-instructions.md'),
        templateSource: 'root/copilot-instructions.template.md',
      })
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const _githubDir = path.join(ctx.targetDir, '.github')
    console.log('🗑️  Removing GitHub Copilot tools...')
    // Basic remove implementation
  }
}

export function skillToAgentYaml(skillId: string, skillBody: string): string {
  const body = skillBody.trim()
  if (body.length === 0) {
    throw new Error(`skill ${skillId} has no content`)
  }

  return `name: ${skillId}
displayName: ${toDisplayName(skillId)}
description: >
  ${skillId} skill for the ai-setup orchestrator.
model: claude-sonnet-4.5
tools:
  - "*"
promptParts:
  includeAISafety: true
  includeToolInstructions: true
  includeParallelToolCalling: true
  includeCustomAgentInstructions: false
prompt: |
${indentLines(body, '  ')}
`
}

function toDisplayName(skillId: string): string {
  return skillId
    .split('-')
    .map(part => (part.length > 0 ? `${part[0]?.toUpperCase()}${part.slice(1)}` : part))
    .join(' ')
}

function indentLines(text: string, indent: string): string {
  return text
    .split('\n')
    .map((line, index, lines) => (line !== '' || index < lines.length - 1 ? `${indent}${line}` : line))
    .join('\n')
}

function copilotProbePasses(homeDir: string): boolean {
  if (files.fileExists(path.join(homeDir, '.copilot'))) {
    return true
  }

  try {
    execFileSync('copilot', ['--version'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
}
