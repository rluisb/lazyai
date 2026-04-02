import fs from 'node:fs'
import path from 'node:path'

const GITIGNORE_ENTRY = '.ai/memory/'

export function checkGitignoreGuidance(targetDir: string): void {
  const gitignorePath = path.join(targetDir, '.gitignore')

  if (!fs.existsSync(gitignorePath)) {
    console.log('\n💡 Consider creating a .gitignore with:')
    console.log(`   ${GITIGNORE_ENTRY}`)
    return
  }

  const content = fs.readFileSync(gitignorePath, 'utf-8')
  if (!content.includes('.ai/memory')) {
    console.log('\n💡 Consider adding to .gitignore:')
    console.log(`   ${GITIGNORE_ENTRY}`)
  }
}
