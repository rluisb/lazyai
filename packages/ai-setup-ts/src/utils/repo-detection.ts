import { readdirSync } from 'node:fs'
import path from 'node:path'
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

export interface ProjectStack {
  language: string
  framework?: string
  testFramework?: string
  packageManager?: string
  commands: {
    test?: string
    lint?: string
    build?: string
    dev?: string
    install?: string
  }
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

function detectPackageManager(absolutePath: string): string | undefined {
  if (fileExists(path.join(absolutePath, 'pnpm-lock.yaml'))) return 'pnpm'
  if (fileExists(path.join(absolutePath, 'yarn.lock'))) return 'yarn'
  if (fileExists(path.join(absolutePath, 'bun.lockb')) || fileExists(path.join(absolutePath, 'bun.lock'))) return 'bun'
  if (fileExists(path.join(absolutePath, 'package-lock.json'))) return 'npm'
  if (fileExists(path.join(absolutePath, 'Gemfile.lock'))) return 'bundle'
  if (fileExists(path.join(absolutePath, 'Cargo.lock'))) return 'cargo'
  if (fileExists(path.join(absolutePath, 'requirements.txt')) || fileExists(path.join(absolutePath, 'pyproject.toml'))) {
    return 'pip'
  }

  return undefined
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

export function detectProjectStack(absolutePath: string): ProjectStack {
  const repoType = detectRepoType(absolutePath)
  const description = getRepoDescription(absolutePath)

  const languageMap: Record<RepoType, string> = {
    'ruby-rails': 'Ruby',
    'nextjs-typescript': 'TypeScript',
    'react-typescript': 'TypeScript',
    go: 'Go',
    rust: 'Rust',
    python: 'Python',
    unknown: 'Unknown',
  }

  const frameworkMap: Partial<Record<RepoType, string>> = {
    'ruby-rails': 'Rails',
    'nextjs-typescript': 'Next.js',
    'react-typescript': 'React',
  }

  const language = languageMap[repoType]
  const framework = frameworkMap[repoType]
  const packageManager = detectPackageManager(absolutePath)
  const commands: ProjectStack['commands'] = {}

  const packageJsonPath = path.join(absolutePath, 'package.json')
  if (fileExists(packageJsonPath)) {
    try {
      const pkg = JSON.parse(readFile(packageJsonPath)) as {
        scripts?: Record<string, string>
      }
      const scripts = pkg.scripts ?? {}
      const packageManagerCommand = packageManager ?? 'npm'

      if (scripts.test) commands.test = `${packageManagerCommand} test`
      if (scripts.lint) commands.lint = `${packageManagerCommand} run lint`
      if (scripts.build) commands.build = `${packageManagerCommand} run build`
      if (scripts.dev) commands.dev = `${packageManagerCommand} run dev`
      commands.install = `${packageManagerCommand} install`
    } catch {
      // ignore invalid package.json
    }
  } else {
    switch (repoType) {
      case 'go':
        commands.test = 'go test ./...'
        commands.build = 'go build ./...'
        commands.install = 'go mod download'
        break
      case 'rust':
        commands.test = 'cargo test'
        commands.build = 'cargo build'
        commands.lint = 'cargo clippy'
        commands.install = 'cargo build'
        break
      case 'python':
        commands.test = 'pytest'
        commands.lint = 'ruff check .'
        commands.install = 'pip install -r requirements.txt'
        break
      case 'ruby-rails':
        commands.test = 'bundle exec rspec'
        commands.build = 'bundle exec rails assets:precompile'
        commands.lint = 'bundle exec rubocop'
        commands.install = 'bundle install'
        commands.dev = 'bundle exec rails server'
        break
      default:
        break
    }
  }

  let testFramework: string | undefined
  if (packageHasDependency(absolutePath, 'vitest')) testFramework = 'Vitest'
  else if (packageHasDependency(absolutePath, 'jest')) testFramework = 'Jest'
  else if (packageHasDependency(absolutePath, 'mocha')) testFramework = 'Mocha'
  else if (repoType === 'go') testFramework = 'go test'
  else if (repoType === 'rust') testFramework = 'cargo test'
  else if (repoType === 'python') testFramework = 'pytest'
  else if (repoType === 'ruby-rails') testFramework = 'RSpec'

  return {
    language,
    ...(framework ? { framework } : {}),
    ...(testFramework ? { testFramework } : {}),
    ...(packageManager ? { packageManager } : {}),
    commands,
    ...(description ? { description } : {}),
  }
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
