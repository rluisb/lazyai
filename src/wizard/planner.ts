import path from 'node:path'
import { ROOT_FILE_BY_TOOL, ROOT_TEMPLATE_BY_FILE } from '../scaffold/root-file-map.js'
import type { ToolId, WizardConfig } from '../types.js'
import { fileExists } from '../utils/files.js'

export interface PlannedFile {
  destPath: string
  srcPath: string | null
  category: 'template' | 'rule' | 'agent' | 'skill' | 'prompt' | 'infra' | 'root' | 'constitution'
  isNew: boolean
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
  codeowners: {
    destPath: 'CODEOWNERS',
    srcPath: 'infra/CODEOWNERS.template',
  },
}

const ADAPTER_PATHS: Record<ToolId, { agentDir?: string; skillDir: string; promptDir?: string }> = {
  'claude-code': {
    agentDir: '.claude/agents',
    skillDir: '.claude/skills',
  },
  opencode: {
    agentDir: '.opencode/agents',
    skillDir: '.opencode/skills',
  },
  gemini: {
    skillDir: '.gemini/skills',
  },
  copilot: {
    agentDir: '.github/agents',
    skillDir: '.github/prompts',
    promptDir: '.github/prompts',
  },
  pi: {
    agentDir: '.pi/agents',
    skillDir: '.pi/skills',
    promptDir: '.pi/templates',
  },
  codex: {
    // Codex agents are inline in AGENTS.md, no separate directory
    skillDir: '.codex/skills',
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

export async function computePlan(
  config: WizardConfig,
  targetDir: string,
  selections: WizardConfig['selections'],
): Promise<PlannedFile[]> {
  const planned: PlannedFile[] = []

  // 1) constitution files
  for (const fileName of selections.constitution) {
    planned.push(
      makePlannedFile(targetDir, fileName, `constitution/${fileName}`, 'constitution'),
    )
  }

  // 2) root constitution (e.g., AGENTS.md or CLAUDE.md)
  const rootFiles = new Set<string>()
  for (const tool of config.tools) {
    const rootFile = ROOT_FILE_BY_TOOL[tool]
    if (rootFile) {
      if (!rootFiles.has(rootFile)) {
        rootFiles.add(rootFile)
        const srcPath = ROOT_TEMPLATE_BY_FILE[rootFile] ?? null
        planned.push(
          makePlannedFile(targetDir, rootFile, srcPath, 'root'),
        )
      }
    }
  }

  // 3) rule files
  for (const fileName of selections.rules) {
    planned.push(
      makePlannedFile(targetDir, fileName, `rules/${fileName}.md`, 'rule'),
    )
  }

  // 4) infra files
  for (const key of selections.infra) {
    const mapped = INFRA_FILE_MAP[key]
    if (mapped) {
      planned.push(
        makePlannedFile(targetDir, mapped.destPath, mapped.srcPath, 'infra'),
      )
    }
  }

  // 5) adapter files (agents, skills, prompts)
  for (const tool of config.tools) {
    const paths = ADAPTER_PATHS[tool]

    // Only add agents if tool supports them (Codex has inline agents)
    if (paths.agentDir) {
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
    }

    for (const skillId of selections.skills) {
      const skillDestPath = tool === 'copilot'
        ? `${skillId}.prompt.md`
        : tool === 'claude-code' || tool === 'opencode' || tool === 'codex' || tool === 'gemini'
          ? `${skillId}/SKILL.md`
          : `${skillId}.md`
      planned.push(
        makePlannedFile(
          targetDir,
          path.posix.join(paths.skillDir, skillDestPath),
          `skills/${skillId}.md`,
          'skill',
        ),
      )
    }

    if (tool === 'claude-code' || tool === 'opencode' || tool === 'codex') {
      continue
    }

    // Only add prompts if tool supports them
    if (paths.promptDir) {
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
  }

  return planned
}
