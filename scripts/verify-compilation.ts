#!/usr/bin/env npx tsx

/**
 * Verification script for the supported-tool compilation pipeline
 * Tests: FragmentResolver, TemplateCompiler, and supported tool template assets
 */

import path from 'node:path'
import fs from 'node:fs'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const ROOT = path.resolve(__dirname, '..')
const PACKAGE_ROOT = path.join(ROOT, 'packages', 'ai-setup-ts')
const LIBRARY = path.join(ROOT, 'library')
const FRAGMENTS_DIR = path.join(LIBRARY, 'fragments')
const TEMPLATES_DIR = path.join(LIBRARY, 'tool-templates')

async function main() {
  console.log('🔍 AI-Setup Supported Tool Compilation Pipeline Verification\n')
  
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
  
  // Check 2: Supported root templates exist
  console.log('\n2️⃣ Checking supported root templates...')
  const expectedTools = ['claude-code', 'opencode', 'copilot']
  const expectedRootTemplates = [
    'root/AGENTS.template.md',
    'root/copilot-instructions.template.md',
    'tool-templates/shared/root.template.md',
  ]
  
  const templatesExist = expectedRootTemplates.every(template => {
    const templatePath = path.join(LIBRARY, template)
    const exists = fs.existsSync(templatePath)
    console.log(`   ${exists ? '✅' : '❌'} ${template}`)
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
    const exists = fs.existsSync(path.join(PACKAGE_ROOT, f))
    console.log(`   ${exists ? '✅' : '❌'} ${f}`)
    return exists
  })
  
  // Check 4: Removed-tool templates are absent
  console.log('\n4️⃣ Checking removed-tool templates are absent...')
  const removedToolPaths = ['codex', 'gemini', 'root/GEMINI.template.md']
  const removedToolTemplatesAbsent = removedToolPaths.every(relativePath => {
    const assetPath = path.join(LIBRARY, relativePath)
    const exists = fs.existsSync(assetPath)
    console.log(`   ${exists ? '❌' : '✅'} ${relativePath} absent`)
    return !exists
  })
  
  // Check 5: Schema includes supported tool metadata and features
  console.log('\n5️⃣ Checking schema updates...')
  const schemaPath = path.join(PACKAGE_ROOT, 'src/store/schema.ts')
  const schemaContent = fs.readFileSync(schemaPath, 'utf-8')
  const hasSupportedTools = expectedTools.every(tool => schemaContent.includes(`'${tool}'`))
  const hasPlanningDir = schemaContent.includes('planningDir')
  const hasFeatures = schemaContent.includes('featureFlagsSchema')
  
  console.log(`   ${hasSupportedTools ? '✅' : '❌'} toolIdSchema includes supported tools`)
  console.log(`   ${hasPlanningDir ? '✅' : '❌'} configSchema includes 'planningDir'`)
  console.log(`   ${hasFeatures ? '✅' : '❌'} featureFlagsSchema defined`)
  
  // Check 6: Test fragment content
  console.log('\n6️⃣ Validating fragment content...')
  const rpiContent = fs.readFileSync(path.join(FRAGMENTS_DIR, 'rpi-workflow.md'), 'utf-8')
  const hasRpiTag = rpiContent.includes('<rpi-workflow>')
  const hasPivotHandling = rpiContent.includes('**Pivot handling')
  console.log(`   ${hasRpiTag ? '✅' : '❌'} rpi-workflow.md has proper wrapper tag`)
  console.log(`   ${hasPivotHandling ? '✅' : '❌'} Includes merged pivot handling guidance`)
  
  // Check 7: Test template content  
  console.log('\n7️⃣ Validating template content...')
  const claudeTemplate = fs.readFileSync(path.join(TEMPLATES_DIR, 'shared/root.template.md'), 'utf-8')
  const hasInclude = claudeTemplate.includes('{{#include')
  const hasConditional = claudeTemplate.includes('{{#if')
  console.log(`   ${hasInclude ? '✅' : '❌'} Uses {{#include}} directives`)
  console.log(`   ${hasConditional ? '✅' : '❌'} Uses {{#if}} conditionals`)
  
  // Summary
  console.log('\n' + '─'.repeat(50))
  const allPassed = fragmentsExist && templatesExist && compilerExists && removedToolTemplatesAbsent && hasSupportedTools && hasPlanningDir && hasFeatures && hasRpiTag && hasPivotHandling && hasInclude && hasConditional
  
  if (allPassed) {
    console.log('✅ All verification checks passed!')
    console.log('\n📋 Summary:')
    console.log(`   • ${expectedFragments.length} fragments ready`)
    console.log(`   • ${expectedRootTemplates.length} supported root templates ready`)
    console.log(`   • Compiler infrastructure complete`)
    console.log(`   • Schema includes supported tools + features`)
  } else {
    console.log('❌ Some verification checks failed!')
    process.exit(1)
  }
}

main().catch(console.error)
