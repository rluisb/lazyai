import { join } from 'node:path'
import type { AiSetupConfig, WizardSelections } from '../types.js'
import type { FeatureFlags, GitConventions } from '../store/schema.js'
import {
  ALL_AGENTS,
  ALL_INFRA,
  ALL_PROMPTS,
  ALL_RULES,
  ALL_SKILLS,
  ALL_TEMPLATES,
} from '../types.js'
import { fileExists } from './files.js'
import { readStore } from '../store/index.js'

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
  const files = manifest.files || []
  const paths = files.map((f) => f.path)

  const templates = ALL_TEMPLATES.filter((t) =>
    paths.some((p) => p.includes(`templates/${t}`) || p.includes(`.ai/templates/${t}`)),
  )
  const rules = ALL_RULES.filter((r) =>
    paths.some((p) => p.includes(`rules/${r}`) || p.includes(`.ai/rules/${r}`)),
  )
  const agents = ALL_AGENTS.filter((a) =>
    paths.some(
      (p) =>
        p.includes(`agents/${a}`) ||
        p.includes(`.ai/agents/${a}`) ||
        p.includes(`.claude/agents/${a}`) ||
        p.includes(`.opencode/agents/${a}`) ||
        p.includes(`.pi/agents/${a}`) ||
        p.includes(`.codex/agents/${a}`) ||
        p.includes(`.github/agents/${a}`),
    ),
  )
  const skills = ALL_SKILLS.filter((s) =>
    paths.some(
      (p) =>
        p.includes(`skills/${s}`) ||
        p.includes(`.ai/skills/${s}`) ||
        p.includes(`.claude/skills/${s}`) ||
        p.includes(`.opencode/skills/${s}`) ||
        p.includes(`.pi/skills/${s}`) ||
        p.includes(`.codex/skills/${s}`) ||
        p.includes(`.gemini/skills/${s}`),
    ),
  )
  const prompts = ALL_PROMPTS.filter((pr) =>
    paths.some(
      (p) =>
        p.includes(`prompts/${pr}`) ||
        p.includes(`.ai/prompts/${pr}`) ||
        p.includes(`.github/prompts/${pr}`),
    ),
  )
  const infra = ALL_INFRA.filter((i) =>
    paths.some((p) => p.includes(`infra/${i}`) || p.includes(`.ai/infra/${i}`)),
  )

  return {
    ...(templates.length > 0 ? { templates } : {}),
    ...(rules.length > 0 ? { rules } : {}),
    ...(agents.length > 0 ? { agents } : {}),
    ...(skills.length > 0 ? { skills } : {}),
    ...(prompts.length > 0 ? { prompts } : {}),
    ...(infra.length > 0 ? { infra } : {}),
    ...(manifest.selections?.constitution != null ? { constitution: manifest.selections.constitution } : {}),
    ...(manifest.features != null ? { features: manifest.features } : {}),
    ...(manifest.gitConventions != null ? { gitConventions: manifest.gitConventions } : {}),
  }
}
