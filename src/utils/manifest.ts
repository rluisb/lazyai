import { join } from 'node:path'
import type {
  AiSetupConfig,
  WizardSelections,
  DocsDirId,
  AgentId,
  SkillId,
  PromptId,
  TemplateId,
  RuleId,
  InfraId,
} from '../types.js'
import { fileExists, readFile } from './files.js'

const MANIFEST_FILE = '.ai-setup.json'

/**
 * Read existing .ai-setup.json manifest from a target directory.
 * Returns null if file is absent or contains invalid JSON.
 */
export async function readManifest(targetDir: string): Promise<AiSetupConfig | null> {
  const manifestPath = join(targetDir, MANIFEST_FILE)
  if (!(await fileExists(manifestPath))) return null
  try {
    const raw = await readFile(manifestPath)
    return JSON.parse(raw) as AiSetupConfig
  } catch {
    return null
  }
}

/**
 * Extract wizard selections from an existing manifest by mapping
 * file paths back to selection IDs.
 */
export function extractSelections(manifest: AiSetupConfig): Partial<WizardSelections> {
  // If manifest already has selections field (written by wizard), return it
  if (manifest.selections) return manifest.selections

  // Otherwise, infer selections from file paths
  const selections: Partial<WizardSelections> = {}
  const files = manifest.files.map(f => f.path)

  // Docs dirs: look for docs/<dirname>/ paths
  const ALL_DOCS_DIRS: DocsDirId[] = [
    'features',
    'bugfixes',
    'refactors',
    'tech-debt',
    'adrs',
    'memory',
    'prompts',
    'standards',
    'templates',
    'rules',
  ]
  const docsDirs = ALL_DOCS_DIRS.filter(dir => files.some(f => f.startsWith(`docs/${dir}/`) || f === `docs/${dir}`))
  if (docsDirs.length > 0) selections.docsDirs = docsDirs

  // Docs agents: look for docs/<dirname>/AGENTS.md
  const docsAgents = ALL_DOCS_DIRS.filter(dir => files.some(f => f === `docs/${dir}/AGENTS.md`))
  if (docsAgents.length > 0) selections.docsAgents = docsAgents

  // Templates: look for docs/templates/<name>.md
  const ALL_TEMPLATES: TemplateId[] = ['adr', 'bugfix-rca-template', 'code-review-template', 'postmortem-template', 'prd-template', 'progress', 'standard', 'task', 'tasks-template', 'tech-debt-template', 'techspec-template']
  const templates = ALL_TEMPLATES.filter(t => files.some(f => f === `docs/templates/${t}.md`))
  if (templates.length > 0) selections.templates = templates

  // Rules: look for docs/rules/<name>.md
  const ALL_RULES: RuleId[] = ['cost', 'review', 'security', 'workflow']
  const rules = ALL_RULES.filter(r => files.some(f => f === `docs/rules/${r}.md`))
  if (rules.length > 0) selections.rules = rules

  // Agents: look for agent files in any adapter dir pattern
  const ALL_AGENTS: AgentId[] = ['builder', 'documenter', 'planner', 'red-team', 'reviewer', 'scout']
  const agents = ALL_AGENTS.filter(a =>
    files.some(
      f =>
        f.endsWith(`/${a}.md`) &&
        (f.includes('.claude/') ||
          f.includes('.opencode/') ||
          f.includes('.gemini/') ||
          f.includes('.pi/') ||
          f.includes('.github/')),
    ),
  )
  if (agents.length > 0) selections.agents = agents

  // Skills
  const ALL_SKILLS: SkillId[] = ['implement', 'iterate', 'plan', 'research']
  const skills = ALL_SKILLS.filter(s =>
    files.some(f => f.endsWith(`/${s}.md`) && (f.includes('commands/') || f.includes('skills/') || f.includes('prompts/'))),
  )
  if (skills.length > 0) selections.skills = skills

  // Prompts
  const ALL_PROMPTS: PromptId[] = ['compact', 'implement', 'local-example', 'plan', 'research']
  const prompts = ALL_PROMPTS.filter(p => files.some(f => f.endsWith(`/${p}.md`) && f.includes('templates/')))
  if (prompts.length > 0) selections.prompts = prompts

  // Infra
  const ALL_INFRA: InfraId[] = ['pre-commit', 'compliance', 'KNOWLEDGE_MAP']
  const infra = ALL_INFRA.filter(i =>
    files.some(f => {
      const name = f.split('/').pop() || ''
      return name.startsWith(i)
    }),
  )
  if (infra.length > 0) selections.infra = infra

  return selections
}
