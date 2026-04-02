import { describe, expect, it } from 'vitest'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { detectRepoInfo, detectRepoType, scanWorkspaceRepos } from '../utils/repo-detection.js'

function makeTempDir(prefix: string): string {
  return mkdtempSync(path.join(tmpdir(), prefix))
}

describe('repo detection utilities', () => {
  it('detectRepoType detects known marker combinations', () => {
    const root = makeTempDir('ai-setup-repo-types-')

    const railsRepo = path.join(root, 'rails')
    mkdirSync(path.join(railsRepo, 'config'), { recursive: true })
    writeFileSync(path.join(railsRepo, 'Gemfile'), 'source "https://rubygems.org"')
    writeFileSync(path.join(railsRepo, 'config', 'routes.rb'), 'Rails.application.routes.draw do end')

    const nextRepo = path.join(root, 'next')
    mkdirSync(nextRepo, { recursive: true })
    writeFileSync(path.join(nextRepo, 'package.json'), JSON.stringify({ dependencies: { next: '14.0.0' } }))

    const reactRepo = path.join(root, 'react')
    mkdirSync(reactRepo, { recursive: true })
    writeFileSync(path.join(reactRepo, 'package.json'), JSON.stringify({ dependencies: { react: '18.0.0' } }))

    const goRepo = path.join(root, 'go')
    mkdirSync(goRepo, { recursive: true })
    writeFileSync(path.join(goRepo, 'go.mod'), 'module example.com/test')

    const rustRepo = path.join(root, 'rust')
    mkdirSync(rustRepo, { recursive: true })
    writeFileSync(path.join(rustRepo, 'Cargo.toml'), '[package]\nname = "demo"')

    const pythonRepo = path.join(root, 'python')
    mkdirSync(pythonRepo, { recursive: true })
    writeFileSync(path.join(pythonRepo, 'pyproject.toml'), '[project]\nname = "demo"')

    const unknownRepo = path.join(root, 'unknown')
    mkdirSync(unknownRepo, { recursive: true })

    expect(detectRepoType(railsRepo)).toBe('ruby-rails')
    expect(detectRepoType(nextRepo)).toBe('nextjs-typescript')
    expect(detectRepoType(reactRepo)).toBe('react-typescript')
    expect(detectRepoType(goRepo)).toBe('go')
    expect(detectRepoType(rustRepo)).toBe('rust')
    expect(detectRepoType(pythonRepo)).toBe('python')
    expect(detectRepoType(unknownRepo)).toBe('unknown')

    rmSync(root, { recursive: true, force: true })
  })

  it('scanWorkspaceRepos filters non-git and planning repo, sorted by name', () => {
    const workspaceRoot = makeTempDir('ai-setup-repo-scan-')
    const planningRepo = path.join(workspaceRoot, 'planning-repo')
    const bRepo = path.join(workspaceRoot, 'b-repo')
    const aRepo = path.join(workspaceRoot, 'a-repo')
    const nonGit = path.join(workspaceRoot, 'non-git')

    mkdirSync(path.join(planningRepo, '.git'), { recursive: true })
    mkdirSync(path.join(bRepo, '.git'), { recursive: true })
    mkdirSync(path.join(aRepo, '.git'), { recursive: true })
    mkdirSync(nonGit, { recursive: true })

    writeFileSync(path.join(aRepo, 'go.mod'), 'module example.com/a')
    writeFileSync(path.join(bRepo, 'Cargo.toml'), '[package]\nname = "b"')

    const scanned = scanWorkspaceRepos(workspaceRoot, planningRepo)
    expect(scanned.map((repo) => repo.name)).toEqual(['a-repo', 'b-repo'])
    expect(scanned).toEqual([
      {
        name: 'a-repo',
        path: '../a-repo',
        type: 'go',
        isGitRepo: true,
      },
      {
        name: 'b-repo',
        path: '../b-repo',
        type: 'rust',
        isGitRepo: true,
      },
    ])

    rmSync(workspaceRoot, { recursive: true, force: true })
  })

  it('detectRepoInfo reports git status correctly', () => {
    const root = makeTempDir('ai-setup-repo-info-')
    const gitRepo = path.join(root, 'git-repo')
    const plainDir = path.join(root, 'plain-dir')

    mkdirSync(path.join(gitRepo, '.git'), { recursive: true })
    mkdirSync(plainDir, { recursive: true })

    const gitInfo = detectRepoInfo(gitRepo, root)
    const plainInfo = detectRepoInfo(plainDir, root)

    expect(gitInfo.isGitRepo).toBe(true)
    expect(plainInfo.isGitRepo).toBe(false)

    rmSync(root, { recursive: true, force: true })
  })
})
