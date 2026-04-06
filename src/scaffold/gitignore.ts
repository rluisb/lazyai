import fs from 'node:fs'
import path from 'node:path'

const GITIGNORE_ENTRIES = [
  '.ai/memory/',
  '.env',
  '.env.local',
  '.env*.local',
]

export function checkGitignoreGuidance(targetDir: string): void {
  const gitignorePath = path.join(targetDir, '.gitignore')

  if (!fs.existsSync(gitignorePath)) {
    console.log('\n💡 Consider creating a .gitignore with:')
    for (const entry of GITIGNORE_ENTRIES) {
      console.log(`   ${entry}`)
    }
    return
  }

  const content = fs.readFileSync(gitignorePath, 'utf-8')
  const missing: string[] = []

  for (const entry of GITIGNORE_ENTRIES) {
    // Check if the entry (or a broader pattern covering it) is already present
    const searchPattern = entry.replace(/\*/g, '').replace(/\./g, '\\.')
    if (!content.includes(entry) && !new RegExp(searchPattern).test(content)) {
      missing.push(entry)
    }
  }

  if (missing.length > 0) {
    console.log('\n💡 Consider adding to .gitignore:')
    for (const entry of missing) {
      console.log(`   ${entry}`)
    }
  }
}
