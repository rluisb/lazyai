import path from 'node:path'
import type { WizardConfig, ToolId } from '../types.js'
import { fileExists } from '../utils/files.js'

export interface PlannedFile {
  destPath: string
  srcPath: string | null
  category: 'docs' | 'docs-agent' | 'template' | 'rule' | 'agent' | 'skill' | 'prompt' | 'infra' | 'root'
  isNew: boolean
}

const ROOT_FILE_BY_TOOL: Record<ToolId, string> = {
  opencode: 'AGENTS.md',
  pi: 'CLAUDE.md',
  'claude-code': 'CLAUDE.md',
  gemini: 'GEMINI.md',
  copilot: '.github/copilot-instructions.md',
}

const INFRA_FILE_MAP: Record<WizardConfig['selections']['infra'][number], { destPath: string; srcPath: string }> = {
  'pre-commit': {
    destPath: '.git/hooks/pre-commit',
    srcPath: 'infra/pre-commit.hook',
  },
  compliance: {
    destPath: 'docs/compliance.md',
    srcPath: 'infra/compliance.md',
  },
  KNOWLEDGE_MAP: {
    destPath: 'KNOWLEDGE_MAP.md',
    srcPath: 'infra/KNOWLEDGE_MAP.template.md',
  },
}

const ADAPTER_PATHS: Record<ToolId, { agentDir: string; skillDir: string; promptDir: string }> = {
  'claude-code': {
    agentDir: '.claude',
    skillDir: '.claude/commands',
    promptDir: '.claude/templates',
  },
  opencode: {
    agentDir: '.opencode/agents',
    skillDir: '.opencode/commands',
    promptDir: '.opencode/templates',
  },
  gemini: {
    agentDir: '.gemini',
    skillDir: '.gemini/skills',
    promptDir: '.gemini/templates',
  },
  copilot: {
    agentDir: '.github',
    skillDir: '.github/prompts',
    promptDir: '.github/templates',
  },
  pi: {
    agentDir: '.pi/agents',
    skillDir: '.pi/skills',
    promptDir: '.pi/templates',
  },
}

function makePlannedFile(
  targetDir: string,
  destPath: string,
  srcPath: string | null,
  category: PlannedFile['category'],
): PlannedFile {
  return {
    destPath,
    srcPath,
    category,
    isNew: !fileExists(path.join(targetDir, destPath)),
  }
}

export function planFiles(opts: {
  targetDir: string
  libraryDir: string
  config: WizardConfig
}): PlannedFile[] {
  const { targetDir, config } = opts
  const { selections } = config
  const planned: PlannedFile[] = []

  // 1) docs directories
  for (const dir of selections.docsDirs) {
    planned.push(makePlannedFile(targetDir, `docs/${dir}/`, null, 'docs'))
  }

  // 2) docs agents
  for (const dir of selections.docsAgents) {
    planned.push(
      makePlannedFile(
        targetDir,
        `docs/${dir}/AGENTS.md`,
        `docs-agents/${dir}.md`,
        'docs-agent',
      ),
    )
  }

  // 3) templates
  for (const templateId of selections.templates) {
    planned.push(
      makePlannedFile(
        targetDir,
        `docs/templates/${templateId}.md`,
        `templates/${templateId}.md`,
        'template',
      ),
    )
  }

  // 4) rules
  for (const ruleId of selections.rules) {
    planned.push(
      makePlannedFile(targetDir, `docs/rules/${ruleId}.md`, `rules/${ruleId}.md`, 'rule'),
    )
  }

  // 5) infra
  for (const infraId of selections.infra) {
    const infra = INFRA_FILE_MAP[infraId]
    planned.push(makePlannedFile(targetDir, infra.destPath, infra.srcPath, 'infra'))
  }

  // 6) root files (deduplicated by destination path)
  const seenRootDestPaths = new Set<string>()
  for (const tool of config.tools) {
    const destPath = ROOT_FILE_BY_TOOL[tool]
    if (seenRootDestPaths.has(destPath)) continue
    seenRootDestPaths.add(destPath)
    planned.push(makePlannedFile(targetDir, destPath, null, 'root'))
  }

  // 7) adapter files (agents, skills, prompts)
  for (const tool of config.tools) {
    const paths = ADAPTER_PATHS[tool]

    for (const agentId of selections.agents) {
      planned.push(
        makePlannedFile(
          targetDir,
          path.posix.join(paths.agentDir, `${agentId}.md`),
          `agents/${agentId}.md`,
          'agent',
        ),
      )
    }

    for (const skillId of selections.skills) {
      const skillDestName = tool === 'copilot' ? `${skillId}.prompt.md` : `${skillId}.md`
      planned.push(
        makePlannedFile(
          targetDir,
          path.posix.join(paths.skillDir, skillDestName),
          `skills/${skillId}.md`,
          'skill',
        ),
      )
    }

    for (const promptId of selections.prompts) {
      planned.push(
        makePlannedFile(
          targetDir,
          path.posix.join(paths.promptDir, `${promptId}.md`),
          `prompts/${promptId}.md`,
          'prompt',
        ),
      )
    }
  }

  return planned
}
