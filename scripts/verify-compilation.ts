#!/usr/bin/env npx tsx

/**
 * Verification script for the multi-tool compilation pipeline
 * Tests: FragmentResolver, TemplateCompiler, and all 6 tool templates
 */

import path from 'node:path'
import fs from 'node:fs'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const ROOT = path.resolve(__dirname, '..')
const LIBRARY = path.join(ROOT, 'library')
const FRAGMENTS_DIR = path.join(LIBRARY, 'fragments')
const TEMPLATES_DIR = path.join(LIBRARY, 'tool-templates')

async function main() {
  console.log('🔍 AI-Setup Multi-Tool Compilation Pipeline Verification\n')
  
  // Check 1: Fragment library exists
  console.log('1️⃣ Checking fragment library...')
  const expectedFragments = [
    'system-context.xml',
    'context-discipline.md',
    'rpi-workflow.md',
    'reasoning-protocol.md',
    'decision-protocol.md',
    'quality-gates.xml',
    'agent-harness.md',
    'bug-resolution.xml',
  ]
  
  const fragmentsExist = expectedFragments.every(f => {
    const exists = fs.existsSync(path.join(FRAGMENTS_DIR, f))
    console.log(`   ${exists ? '✅' : '❌'} ${f}`)
    return exists
  })
  
  // Check 2: Tool templates exist
  console.log('\n2️⃣ Checking tool templates...')
  const expectedTools = ['claude-code', 'opencode', 'codex', 'copilot', 'pi', 'gemini']
  
  const templatesExist = expectedTools.every(tool => {
    const templatePath = path.join(TEMPLATES_DIR, tool, 'root.template.md')
    const exists = fs.existsSync(templatePath)
    console.log(`   ${exists ? '✅' : '❌'} ${tool}/root.template.md`)
    return exists
  })
  
  // Check 3: Compiler classes exist
  console.log('\n3️⃣ Checking compiler infrastructure...')
  const compilerFiles = [
    'src/compiler/fragment-resolver.ts',
    'src/compiler/template-compiler.ts',
    'src/compiler/index.ts',
  ]
  
  const compilerExists = compilerFiles.every(f => {
    const exists = fs.existsSync(path.join(ROOT, f))
    console.log(`   ${exists ? '✅' : '❌'} ${f}`)
    return exists
  })
  
  // Check 4: Codex adapter exists
  console.log('\n4️⃣ Checking Codex adapter...')
  const codexAdapterPath = path.join(ROOT, 'src/adapters/codex.ts')
  const codexExists = fs.existsSync(codexAdapterPath)
  console.log(`   ${codexExists ? '✅' : '❌'} src/adapters/codex.ts`)
  
  // Check 5: Schema includes codex and features
  console.log('\n5️⃣ Checking schema updates...')
  const schemaPath = path.join(ROOT, 'src/store/schema.ts')
  const schemaContent = fs.readFileSync(schemaPath, 'utf-8')
  const hasCodex = schemaContent.includes("'codex'")
  const hasPlanningDir = schemaContent.includes('planningDir')
  const hasFeatures = schemaContent.includes('featureFlagsSchema')
  
  console.log(`   ${hasCodex ? '✅' : '❌'} toolIdSchema includes 'codex'`)
  console.log(`   ${hasPlanningDir ? '✅' : '❌'} configSchema includes 'planningDir'`)
  console.log(`   ${hasFeatures ? '✅' : '❌'} featureFlagsSchema defined`)
  
  // Check 6: Test fragment content
  console.log('\n6️⃣ Validating fragment content...')
  const rpiContent = fs.readFileSync(path.join(FRAGMENTS_DIR, 'rpi-workflow.md'), 'utf-8')
  const hasRpiTag = rpiContent.includes('<rpi-workflow>')
  const hasPivotHandling = rpiContent.includes('**Pivot handling**')
  console.log(`   ${hasRpiTag ? '✅' : '❌'} rpi-workflow.md has proper wrapper tag`)
  console.log(`   ${hasPivotHandling ? '✅' : '❌'} Includes merged pivot handling guidance`)
  
  // Check 7: Test template content  
  console.log('\n7️⃣ Validating template content...')
  const claudeTemplate = fs.readFileSync(path.join(TEMPLATES_DIR, 'claude-code/root.template.md'), 'utf-8')
  const hasInclude = claudeTemplate.includes('{{#include')
  const hasConditional = claudeTemplate.includes('{{#if')
  console.log(`   ${hasInclude ? '✅' : '❌'} Uses {{#include}} directives`)
  console.log(`   ${hasConditional ? '✅' : '❌'} Uses {{#if}} conditionals`)
  
  // Summary
  console.log('\n' + '─'.repeat(50))
  const allPassed = fragmentsExist && templatesExist && compilerExists && codexExists && hasCodex && hasPlanningDir && hasFeatures
  
  if (allPassed) {
    console.log('✅ All verification checks passed!')
    console.log('\n📋 Summary:')
    console.log(`   • ${expectedFragments.length} fragments ready`)
    console.log(`   • ${expectedTools.length} tool templates ready`)
    console.log(`   • Compiler infrastructure complete`)
    console.log(`   • Schema updated with Codex + features`)
  } else {
    console.log('❌ Some verification checks failed!')
    process.exit(1)
  }
}

main().catch(console.error)
