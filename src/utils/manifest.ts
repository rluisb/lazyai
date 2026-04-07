import { join } from 'node:path'
import { readStore } from '../store/index.js'
import type { FeatureFlags, GitConventions } from '../store/schema.js'
import type { AiSetupConfig, WizardSelections } from '../types.js'
import {
  ALL_AGENTS,
  ALL_INFRA,
  ALL_PROMPTS,
  ALL_RULES,
  ALL_SKILLS,
  ALL_TEMPLATES,
} from '../types.js'
import { fileExists } from './files.js'

const MANIFEST_FILE = '.ai-setup.json'

/**
 * Extended manifest type with new fields
 */
export interface ManifestWithFeatures extends AiSetupConfig {
  planningDir?: string
  features?: FeatureFlags
  gitConventions?: GitConventions
}

/**
 * @deprecated Use readStore() from src/store/index.ts directly.
 * This wrapper remains only for backward compatibility and will be removed in a future version.
 */
export async function readManifest(targetDir: string): Promise<ManifestWithFeatures | null> {
  const manifestPath = join(targetDir, MANIFEST_FILE)
  if (!fileExists(manifestPath)) return null

  try {
    const data = await readStore(targetDir)
    return {
      version: data.meta.cliVersion,
      setupScope: data.config.setupScope,
      ...(data.config.setupType ? { setupType: data.config.setupType } : {}),
      tools: data.config.tools,
      projectName: data.config.projectName,
      installedAt: data.meta.installedAt,
      files: data.files.map((file) => ({
        path: file.path,
        hash: file.hash,
        source: file.source,
        owner: file.owner ?? 'library',
      })),
      selections: data.selections,
      ...(data.config.planningDir != null ? { planningDir: data.config.planningDir } : {}),
      ...(data.selections.features != null ? { features: data.selections.features } : {}),
      ...(data.selections.gitConventions != null ? { gitConventions: data.selections.gitConventions } : {}),
    }
  } catch {
    return null
  }
}

/**
 * Extract wizard selections from an existing manifest by mapping
 * file paths back to selection IDs.
 */
export function extractSelections(manifest: ManifestWithFeatures): Partial<WizardSelections> {
  // If manifest already has selections field (written by wizard), return it
  if (manifest.selections) return manifest.selections

  // Otherwise, infer selections from file paths
  const selections: Partial<WizardSelections> = {}
  const files = manifest.files.map(f => f.path)

  // Templates: look for specs/templates/<name>.md
  const templates = ALL_TEMPLATES.filter(t => files.some(f => f === `specs/templates/${t}.md`))
  if (templates.length > 0) selections.templates = templates

  // Rules: look for specs/rules/<name>.md
  const rules = ALL_RULES.filter(r => files.some(f => f === `specs/rules/${r}.md`))
  if (rules.length > 0) selections.rules = rules

  // Agents: look for agent files in any adapter dir pattern
  const agents = ALL_AGENTS.filter(a =>
    files.some(
      f =>
        f.endsWith(`/${a}.md`) &&
        (f.includes('.claude/') ||
          f.includes('.opencode/') ||
          f.includes('.gemini/') ||
          f.includes('.pi/') ||
          f.includes('.agents/') ||
          f.includes('.github/')),
    ),
  )
  if (agents.length > 0) selections.agents = agents

  // Skills
  const skills = ALL_SKILLS.filter(s =>
    files.some(f =>
      f.endsWith(`/${s}.md`) || f.endsWith(`/skills/${s}/SKILL.md`) || f.endsWith(`/prompts/${s}.prompt.md`),
    ),
  )
  if (skills.length > 0) selections.skills = skills

  // Prompts
  const prompts = ALL_PROMPTS.filter(pr => files.some(f => f.endsWith(`/${pr}.md`) && f.includes('templates/')))
  if (prompts.length > 0) selections.prompts = prompts

  // Infra
  const infra = ALL_INFRA.filter(i =>
    files.some(f => {
      const name = f.split('/').pop() || ''
      if (i === 'codeowners') return name === 'CODEOWNERS'
      return name.startsWith(i)
    }),
  )
  if (infra.length > 0) selections.infra = infra

  return selections
}
