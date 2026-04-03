import path from 'node:path'
import { readdirSync } from 'node:fs'
import { fileExists, isDirectory, readFile } from './files.js'

export type RepoType =
  | 'ruby-rails'
  | 'nextjs-typescript'
  | 'react-typescript'
  | 'go'
  | 'rust'
  | 'python'
  | 'unknown'

export interface RepoInfo {
  name: string
  path: string
  type: RepoType
  isGitRepo: boolean
  description?: string
}

function getRepoDescription(absolutePath: string): string | undefined {
  const packageJsonPath = path.join(absolutePath, 'package.json')
  if (fileExists(packageJsonPath)) {
    try {
      const parsed = JSON.parse(readFile(packageJsonPath)) as { description?: string }
      if (parsed.description) return parsed.description
    } catch {
      // ignore invalid package.json
    }
  }

  const cargoPath = path.join(absolutePath, 'Cargo.toml')
  if (fileExists(cargoPath)) {
    try {
      const content = readFile(cargoPath)
      const match = content.match(/^description\s*=\s*"([^"]+)"/m)
      if (match?.[1]) return match[1]
    } catch {
      // ignore read/parse errors
    }
  }

  const pyprojectPath = path.join(absolutePath, 'pyproject.toml')
  if (fileExists(pyprojectPath)) {
    try {
      const content = readFile(pyprojectPath)
      const match = content.match(/^description\s*=\s*"([^"]+)"/m)
      if (match?.[1]) return match[1]
    } catch {
      // ignore read/parse errors
    }
  }

  return undefined
}

function packageHasDependency(absolutePath: string, dependency: string): boolean {
  const packageJsonPath = path.join(absolutePath, 'package.json')
  if (!fileExists(packageJsonPath)) return false

  try {
    const parsed = JSON.parse(readFile(packageJsonPath)) as {
      dependencies?: Record<string, string>
      devDependencies?: Record<string, string>
      peerDependencies?: Record<string, string>
      optionalDependencies?: Record<string, string>
    }

    return Boolean(
      parsed.dependencies?.[dependency] ||
        parsed.devDependencies?.[dependency] ||
        parsed.peerDependencies?.[dependency] ||
        parsed.optionalDependencies?.[dependency],
    )
  } catch {
    return false
  }
}

export function detectRepoType(absolutePath: string): RepoType {
  if (fileExists(path.join(absolutePath, 'Gemfile')) && fileExists(path.join(absolutePath, 'config', 'routes.rb'))) {
    return 'ruby-rails'
  }

  if (packageHasDependency(absolutePath, 'next')) {
    return 'nextjs-typescript'
  }

  if (packageHasDependency(absolutePath, 'react')) {
    return 'react-typescript'
  }

  if (fileExists(path.join(absolutePath, 'go.mod'))) {
    return 'go'
  }

  if (fileExists(path.join(absolutePath, 'Cargo.toml'))) {
    return 'rust'
  }

  if (fileExists(path.join(absolutePath, 'requirements.txt')) || fileExists(path.join(absolutePath, 'pyproject.toml'))) {
    return 'python'
  }

  return 'unknown'
}

export function detectRepoInfo(absolutePath: string, relativeTo: string): RepoInfo {
  const description = getRepoDescription(absolutePath)

  return {
    name: path.basename(absolutePath),
    path: path.relative(relativeTo, absolutePath).replaceAll('\\', '/'),
    type: detectRepoType(absolutePath),
    isGitRepo: fileExists(path.join(absolutePath, '.git')),
    ...(description ? { description } : {}),
  }
}

export function scanWorkspaceRepos(parentDir: string, planningRepoPath: string): RepoInfo[] {
  const entries = readdirSync(parentDir, { withFileTypes: true })

  return entries
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(parentDir, entry.name))
    .filter((entryPath) => isDirectory(entryPath))
    .filter((entryPath) => path.resolve(entryPath) !== path.resolve(planningRepoPath))
    .map((entryPath) => detectRepoInfo(entryPath, planningRepoPath))
    .filter((repo) => repo.isGitRepo)
    .sort((a, b) => a.name.localeCompare(b.name))
}
