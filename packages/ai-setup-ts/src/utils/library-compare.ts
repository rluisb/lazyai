import { createHash } from 'node:crypto'
import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { join, parse } from 'node:path'
import { readStore } from '../store/index.js'
import type { StoreData } from '../store/schema.js'
import { fileExists, fileHash } from './files.js'
import { parseSkillFrontmatter } from './frontmatter.js'

export type SkillStatus = 'current' | 'missing' | 'modified' | 'drifted'

export interface LibrarySkillInfo {
  name: string
  description: string | undefined
  sourcePath: string
  sourceHash: string
}

export interface InstalledSkillInfo {
  sourceName: string
  installedPath: string
  installedHash: string
  storedHash: string
  storedStatus: string
}

export interface SkillCheckResult {
  name: string
  description: string | undefined
  status: SkillStatus
  sourceHash: string
  installedLocations: Array<{
    path: string
    status: SkillStatus
    installedHash: string | undefined
    storedHash: string | undefined
  }>
}

const SKILLS_DIR = join(process.cwd(), 'library', 'skills')

function sha256(content: string): string {
  return createHash('sha256').update(content).digest('hex')
}

/**
 * Discover all library skills and their source hashes.
 */
export function discoverLibrarySkills(): LibrarySkillInfo[] {
  if (!existsSync(SKILLS_DIR)) return []

  return readdirSync(SKILLS_DIR)
    .filter((f) => f.endsWith('.md'))
    .map((file) => {
      const sourcePath = join(SKILLS_DIR, file)
      const content = readFileSync(sourcePath, 'utf-8')
      const { frontmatter } = parseSkillFrontmatter(content)
      return {
        name: frontmatter.name || parse(file).name,
        description: frontmatter.description ?? undefined,
        sourcePath,
        sourceHash: sha256(content),
      }
    })
}

/**
 * Find installed skill paths from the tracked files manifest.
 */
async function findInstalledSkills(targetDir: string): Promise<InstalledSkillInfo[]> {
  const results: InstalledSkillInfo[] = []

  try {
    const store: StoreData = await readStore(targetDir)
    if (!store?.files) return results

    for (const record of store.files) {
      const segments = record.path.replace(/\\/g, '/').split('/')
      const skillsIdx = segments.lastIndexOf('skills')
      if (skillsIdx === -1) continue

      const nameSegment = segments[skillsIdx + 1]
      const fileName = segments[skillsIdx + 2]
      if (!nameSegment || fileName !== 'SKILL.md') continue

      results.push({
        sourceName: nameSegment,
        installedPath: record.path,
        installedHash: fileExists(join(targetDir, record.path)) ? fileHash(join(targetDir, record.path)) : '',
        storedHash: record.hash,
        storedStatus: record.status ?? 'unknown',
      })
    }
  } catch {
    // Manifest not found or unreadable
  }

  return results
}

/**
 * Compare each library skill against all installed locations.
 */
export async function compareLibrarySkills(targetDir: string): Promise<SkillCheckResult[]> {
  const librarySkills = discoverLibrarySkills()
  const installedSkills = await findInstalledSkills(targetDir)

  const results: SkillCheckResult[] = []

  for (const lib of librarySkills) {
    const locations = installedSkills.filter((inst) => inst.sourceName === lib.name)

    const installedLocations = locations.map((inst) => {
      const status = classifySkillStatus(lib.sourceHash, inst)
      return {
        path: inst.installedPath,
        status,
        installedHash: inst.installedHash || undefined,
        storedHash: inst.storedHash || undefined,
      }
    })

    if (installedLocations.length === 0) {
      results.push({
        name: lib.name,
        description: lib.description ?? undefined,
        status: 'missing',
        sourceHash: lib.sourceHash,
        installedLocations: [{ path: `* /skills/${lib.name}/SKILL.md`, status: 'missing', installedHash: undefined, storedHash: undefined }],
      })
      continue
    }

    const statusOrder: SkillStatus[] = ['missing', 'modified', 'drifted', 'current']
    let overall: SkillStatus = 'current'
    for (const s of statusOrder) {
      if (installedLocations.some((l) => l.status === s)) {
        overall = s
      }
    }

    results.push({
      name: lib.name,
      description: lib.description ?? undefined,
      status: overall,
      sourceHash: lib.sourceHash,
      installedLocations,
    })
  }

  return results
}

function classifySkillStatus(sourceHash: string, inst: InstalledSkillInfo): SkillStatus {
  if (!inst.installedHash) return 'missing'
  if (inst.installedHash === sourceHash) return 'current'
  if (inst.installedHash === inst.storedHash) return 'drifted'
  return 'modified'
}
