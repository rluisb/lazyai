import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { type FragmentContext, FragmentResolver, TemplateCompiler, type ToolId } from '../compiler/index.js'

const libraryDir = path.resolve(process.cwd(), 'library')

describe('FragmentResolver', () => {
  let tempDir: string
  let fragmentsDir: string
  let resolver: FragmentResolver

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'fragment-resolver-'))
    fragmentsDir = path.join(tempDir, 'fragments')
    mkdirSync(fragmentsDir, { recursive: true })
    resolver = new FragmentResolver(tempDir)
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  describe('include directive resolution', () => {
    it('resolves {{#include fragments/X.xml}} directives correctly', () => {
      const fragmentContent = '<test>Fragment content</test>'
      writeFileSync(path.join(fragmentsDir, 'test.xml'), fragmentContent)

      const template = 'Header\n{{#include fragments/test.xml}}\nFooter'
      const context: FragmentContext = {
        projectName: 'Test Project',
        planningDir: 'planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('<test>Fragment content</test>')
      expect(result).toContain('Header')
      expect(result).toContain('Footer')
    })

    it('handles multiple include directives in same template', () => {
      writeFileSync(path.join(fragmentsDir, 'frag1.xml'), '<frag1>Content 1</frag1>')
      writeFileSync(path.join(fragmentsDir, 'frag2.xml'), '<frag2>Content 2</frag2>')

      const template = '{{#include fragments/frag1.xml}}\n{{#include fragments/frag2.xml}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('<frag1>Content 1</frag1>')
      expect(result).toContain('<frag2>Content 2</frag2>')
    })

    it('caches fragment content for performance', () => {
      const fragmentContent = '<cached>Cached content</cached>'
      writeFileSync(path.join(fragmentsDir, 'cached.xml'), fragmentContent)

      const template = '{{#include fragments/cached.xml}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
      }

      // First resolution
      const result1 = resolver.resolve(template, context)
      // Second resolution should use cache
      const result2 = resolver.resolve(template, context)

      expect(result1).toContain('<cached>Cached content</cached>')
      expect(result2).toContain('<cached>Cached content</cached>')
    })

    it('handles missing fragment file gracefully', () => {
      const template = '{{#include fragments/nonexistent.xml}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('<!-- Fragment not found: fragments/nonexistent.xml -->')
    })

    it('preserves content around includes', () => {
      writeFileSync(path.join(fragmentsDir, 'middle.xml'), 'MIDDLE')

      const template = 'START\n{{#include fragments/middle.xml}}\nEND'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).toBe('START\nMIDDLE\nEND')
    })
  })

  describe('conditional resolution', () => {
    it('supports camelCase condition names against snake_case feature context', () => {
      const template = '{{#if features.treeOfThoughts}}CAMELCASE_MATCH{{/if}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          tree_of_thoughts: true,
        },
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('CAMELCASE_MATCH')
    })

    it('supports snake_case condition names against camelCase feature context', () => {
      const template = '{{#if features.agent_harness}}SNAKECASE_MATCH{{/if}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          agentHarness: true,
        },
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('SNAKECASE_MATCH')
    })

    it('includes block when feature flag is true', () => {
      const template = '{{#if features.tree_of_thoughts}}<tree-of-thoughts>Enabled</tree-of-thoughts>{{/if}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          tree_of_thoughts: true,
          agent_harness: false,
          bug_resolution: false,
        },
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('<tree-of-thoughts>Enabled</tree-of-thoughts>')
    })

    it('excludes block when feature flag is false', () => {
      const template = '{{#if features.tree_of_thoughts}}<tree-of-thoughts>Enabled</tree-of-thoughts>{{/if}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          tree_of_thoughts: false,
          agent_harness: false,
          bug_resolution: false,
        },
      }

      const result = resolver.resolve(template, context)
      expect(result).not.toContain('<tree-of-thoughts>')
      expect(result).toBe('')
    })

    it('handles multiple conditional blocks', () => {
      const template = `
{{#if features.tree_of_thoughts}}TREE_OF_THOUGHTS{{/if}}
{{#if features.agent_harness}}AGENT_HARNESS{{/if}}
{{#if features.bug_resolution}}BUG_RESOLUTION{{/if}}
`.trim()

      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          tree_of_thoughts: true,
          agent_harness: false,
          bug_resolution: true,
        },
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('TREE_OF_THOUGHTS')
      expect(result).not.toContain('AGENT_HARNESS')
      expect(result).toContain('BUG_RESOLUTION')
    })

    it('handles nested content in conditionals', () => {
      const template = `
Start
{{#if features.agent_harness}}
Agent Configuration:
- Setting 1
- Setting 2
{{/if}}
End
`.trim()

      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          agent_harness: true,
        },
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('Agent Configuration:')
      expect(result).toContain('- Setting 1')
      expect(result).toContain('- Setting 2')
    })

    it('handles missing feature flag as false', () => {
      const template = '{{#if features.nonexistent}}SHOULD_NOT_APPEAR{{/if}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).not.toContain('SHOULD_NOT_APPEAR')
      expect(result).toBe('')
    })
  })

  describe('variable resolution', () => {
    it('resolves {{PROJECT_NAME}} substitution', () => {
      const template = 'Project: {{PROJECT_NAME}}'
      const context: FragmentContext = {
        projectName: 'My Awesome Project',
        planningDir: 'planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).toBe('Project: My Awesome Project')
    })

    it('resolves {{PLANNING_DIR}} substitution', () => {
      const template = 'Planning directory: {{PLANNING_DIR}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: '/path/to/planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).toBe('Planning directory: /path/to/planning')
    })

    it('resolves {{PRIMARY_LANGUAGE}} with default', () => {
      const template = 'Language: {{PRIMARY_LANGUAGE}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).toBe('Language: TypeScript')
    })

    it('resolves {{PRIMARY_LANGUAGE}} with custom value', () => {
      const template = 'Language: {{PRIMARY_LANGUAGE}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        primaryLanguage: 'Go',
      }

      const result = resolver.resolve(template, context)
      expect(result).toBe('Language: Go')
    })

    it('resolves {{FRAMEWORK}} substitution', () => {
      const template = 'Framework: {{FRAMEWORK}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        framework: 'React',
      }

      const result = resolver.resolve(template, context)
      expect(result).toBe('Framework: React')
    })

    it('resolves {{WORKSPACE_TYPE}} substitution', () => {
      const template = 'Workspace: {{WORKSPACE_TYPE}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        workspaceType: 'monorepo',
      }

      const result = resolver.resolve(template, context)
      expect(result).toBe('Workspace: monorepo')
    })

    it('resolves {{PROJECT_INSTRUCTIONS}} substitution', () => {
      const instructions = 'Use Python style guidelines.\nRun tests before committing.'
      const template = 'Instructions:\n{{PROJECT_INSTRUCTIONS}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        projectInstructions: instructions,
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain(instructions)
    })

    it('resolves tool-specific variables', () => {
      const template = '{{TOOL_DESCRIPTION}}\n{{TOOL_NOTES}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
        toolDescription: 'Uses Claude Code.',
        toolNotes: '## Notes\n\n- Agents live here',
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('Uses Claude Code.')
      expect(result).toContain('## Notes')
      expect(result).toContain('- Agents live here')
    })

    it('handles multiple variable substitutions', () => {
      const template = '# {{PROJECT_NAME}}\nLanguage: {{PRIMARY_LANGUAGE}}\nDir: {{PLANNING_DIR}}'
      const context: FragmentContext = {
        projectName: 'TestProj',
        planningDir: 'plans',
        primaryLanguage: 'Rust',
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('# TestProj')
      expect(result).toContain('Language: Rust')
      expect(result).toContain('Dir: plans')
    })

    it('leaves unresolved variables as-is', () => {
      const template = 'Unknown: {{UNKNOWN_VAR}}'
      const context: FragmentContext = {
        projectName: 'Test',
        planningDir: 'planning',
      }

      const result = resolver.resolve(template, context)
      expect(result).toBe('Unknown: {{UNKNOWN_VAR}}')
    })
  })

  describe('combined resolution', () => {
    it('resolves includes, conditionals, and variables together', () => {
      writeFileSync(path.join(fragmentsDir, 'feature.xml'), '<feature>TOT Content</feature>')

      const template = `# {{PROJECT_NAME}}
Planning: {{PLANNING_DIR}}
{{#if features.tree_of_thoughts}}
{{#include fragments/feature.xml}}
{{/if}}`

      const context: FragmentContext = {
        projectName: 'Combined Test',
        planningDir: 'docs/planning',
        features: {
          tree_of_thoughts: true,
        },
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('# Combined Test')
      expect(result).toContain('Planning: docs/planning')
      expect(result).toContain('<feature>TOT Content</feature>')
    })

    it('respects resolution order: conditionals, includes, variables', () => {
      writeFileSync(path.join(fragmentsDir, 'test.xml'), 'Project: {{PROJECT_NAME}}')

      const template = '{{#if features.agent_harness}}{{#include fragments/test.xml}}{{/if}}'
      const context: FragmentContext = {
        projectName: 'OrderTest',
        planningDir: 'planning',
        features: {
          agent_harness: true,
        },
      }

      const result = resolver.resolve(template, context)
      expect(result).toContain('Project: OrderTest')
    })
  })
})

describe('TemplateCompiler', () => {
  let tempDir: string
  let libraryTempDir: string
  let templateDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'compiler-'))
    libraryTempDir = mkdtempSync(path.join(tmpdir(), 'compiler-lib-'))
    templateDir = path.join(libraryTempDir, 'tool-templates', 'test-tool')
    mkdirSync(templateDir, { recursive: true })
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
    rmSync(libraryTempDir, { recursive: true, force: true })
  })

  it('compiles a tool template and returns files array', () => {
    writeFileSync(path.join(templateDir, 'root.template.md'), '# Test\nContent')

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'TestProj',
        planningDir: 'planning',
      },
    })

    const output = compiler.compile()

    expect(output.tool).toBe('test-tool')
    expect(output.files).toHaveLength(1)
    expect(output.files[0]).toEqual({
      relativePath: 'root.md',
      content: '# Test\nContent',
    })
  })

  it('includes root.md in the files list', () => {
    writeFileSync(path.join(templateDir, 'root.template.md'), 'Root content')

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
      },
    })

    const output = compiler.compile()
    const paths = output.files.map(f => f.relativePath)
    expect(paths).toContain('root.md')
  })

  it('handles nested template directories', () => {
    const nestedDir = path.join(templateDir, 'subfolder')
    mkdirSync(nestedDir, { recursive: true })
    writeFileSync(path.join(templateDir, 'root.template.md'), 'Root')
    writeFileSync(path.join(nestedDir, 'nested.template.md'), 'Nested')

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
      },
    })

    const output = compiler.compile()
    const paths = output.files.map(f => f.relativePath)
    expect(paths).toContain('root.md')
    expect(paths).toContain('subfolder/nested.md')
  })

  it('throws on missing tool directory', () => {
    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'nonexistent-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
      },
    })

    expect(() => compiler.compile()).toThrow(
      'Tool template directory not found'
    )
  })

  it('applies feature flag conditionals to output', () => {
    writeFileSync(
      path.join(templateDir, 'root.template.md'),
      'Start\n{{#if features.tree_of_thoughts}}\nTOT\n{{/if}}\nEnd'
    )

    // Test with feature enabled
    const compilerEnabled = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          tree_of_thoughts: true,
        },
      },
    })

    const outputEnabled = compilerEnabled.compile()
    expect(outputEnabled.files.length).toBeGreaterThan(0)
    expect(outputEnabled.files[0]?.content).toContain('TOT')

    // Test with feature disabled
    const compilerDisabled = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          tree_of_thoughts: false,
        },
      },
    })

    const outputDisabled = compilerDisabled.compile()
    expect(outputDisabled.files.length).toBeGreaterThan(0)
    expect(outputDisabled.files[0]?.content).not.toContain('TOT')
  })

  it('prefers the shared root template and injects tool overrides', () => {
    const sharedDir = path.join(libraryTempDir, 'tool-templates', 'shared')
    mkdirSync(sharedDir, { recursive: true })
    writeFileSync(
      path.join(sharedDir, 'root.template.md'),
      '{{TOOL_DESCRIPTION}}\n{{TOOL_NOTES}}\n{{PROJECT_NAME}}'
    )
    writeFileSync(path.join(templateDir, 'root.template.md'), 'Legacy per-tool root should be ignored')

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'copilot' as ToolId,
      context: {
        projectName: 'SharedTemplateTest',
        planningDir: 'planning',
      },
    })

    const output = compiler.compile()
    expect(output.files).toHaveLength(1)
    expect(output.files[0]?.relativePath).toBe('root.md')
    expect(output.files[0]?.content).toContain('This project uses GitHub Copilot with ai-setup integration.')
    expect(output.files[0]?.content).toContain('## Copilot-Specific Notes')
    expect(output.files[0]?.content).toContain('SharedTemplateTest')
    expect(output.files[0]?.content).not.toContain('Legacy per-tool root should be ignored')
  })

  it('interpolates variables in compiled output', () => {
    writeFileSync(
      path.join(templateDir, 'root.template.md'),
      '# {{PROJECT_NAME}}\nDir: {{PLANNING_DIR}}\nLang: {{PRIMARY_LANGUAGE}}'
    )

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'MyProject',
        planningDir: 'docs/planning',
        primaryLanguage: 'Python',
      },
    })

    const output = compiler.compile()
    expect(output.files.length).toBeGreaterThan(0)
    const content = output.files[0]?.content ?? ''
    expect(content).toContain('# MyProject')
    expect(content).toContain('Dir: docs/planning')
    expect(content).toContain('Lang: Python')
  })

  it('converts .template.md to .md in output', () => {
    writeFileSync(path.join(templateDir, 'config.template.md'), 'Config content')

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
      },
    })

    const output = compiler.compile()
    const paths = output.files.map(f => f.relativePath)
    expect(paths).toContain('config.md')
    expect(paths).not.toContain('config.template.md')
  })

  it('handles multiple file extensions correctly', () => {
    writeFileSync(path.join(templateDir, 'readme.template.md'), 'Readme')
    writeFileSync(path.join(templateDir, 'config.template.yaml'), 'yaml: config')
    writeFileSync(path.join(templateDir, 'script.template.sh'), 'script')

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
      },
    })

    const output = compiler.compile()
    const paths = output.files.map(f => f.relativePath)
    expect(paths).toContain('readme.md')
    expect(paths).toContain('config.yaml')
    expect(paths).toContain('script.sh')
  })

  it('writes compiled files to disk', () => {
    writeFileSync(path.join(templateDir, 'root.template.md'), '# Test')

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
      },
    })

    compiler.compileAndWrite()

    const outputFile = path.join(tempDir, 'root.md')
    expect(readFileSync(outputFile, 'utf-8')).toBe('# Test')
  })

  it('creates nested directories when writing files', () => {
    const nestedDir = path.join(templateDir, 'config', 'nested')
    mkdirSync(nestedDir, { recursive: true })
    writeFileSync(path.join(nestedDir, 'deep.template.md'), 'Deep content')

    const compiler = new TemplateCompiler({
      libraryDir: libraryTempDir,
      outputDir: tempDir,
      tool: 'test-tool' as ToolId,
      context: {
        projectName: 'Test',
        planningDir: 'planning',
      },
    })

    compiler.compileAndWrite()

    const outputFile = path.join(tempDir, 'config', 'nested', 'deep.md')
    expect(readFileSync(outputFile, 'utf-8')).toBe('Deep content')
  })
})

describe('Real library integration', () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'compiler-integration-'))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('compiles all 5 supported tools without errors', () => {
    const tools: ToolId[] = ['claude-code', 'opencode', 'codex', 'copilot', 'gemini']
    const context: FragmentContext = {
      projectName: 'Integration Test',
      planningDir: 'planning',
      primaryLanguage: 'TypeScript',
      framework: 'React',
      features: {
        tree_of_thoughts: true,
        agent_harness: true,
        bug_resolution: true,
      },
    }

    for (const tool of tools) {
      const compiler = new TemplateCompiler({
        libraryDir,
        outputDir: tempDir,
        tool,
        context,
      })

      expect(() => {
        const output = compiler.compile()
        expect(output.tool).toBe(tool)
        expect(output.files.length).toBeGreaterThan(0)
      }).not.toThrow()
    }
  })

  it('compiles claude-code template with real fragments', () => {
    const compiler = new TemplateCompiler({
      libraryDir,
      outputDir: tempDir,
      tool: 'claude-code',
      context: {
        projectName: 'Integration Test Project',
        planningDir: 'docs/planning',
        primaryLanguage: 'TypeScript',
        framework: 'React',
        projectInstructions: 'Follow TypeScript best practices.',
        features: {
          tree_of_thoughts: true,
          agent_harness: true,
          bug_resolution: true,
        },
      },
    })

    const output = compiler.compile()

    expect(output.tool).toBe('claude-code')
    expect(output.files.length).toBeGreaterThan(0)

    const rootFile = output.files.find(f => f.relativePath === 'root.md')
    expect(rootFile).toBeDefined()

    if (rootFile !== undefined) {
      expect(rootFile.content).toContain('Integration Test Project')
      expect(rootFile.content).toContain('docs/planning')
      expect(rootFile.content).toContain('Follow TypeScript best practices.')
      // Check that features were included
      expect(rootFile.content).toContain('decision-protocol')
      expect(rootFile.content).toContain('agent-harness')
    }
  })

  it('filters features correctly in real template', () => {
    // Test with features disabled
    const compiler = new TemplateCompiler({
      libraryDir,
      outputDir: tempDir,
      tool: 'claude-code',
      context: {
        projectName: 'Test',
        planningDir: 'planning',
        features: {
          tree_of_thoughts: false,
          agent_harness: false,
          bug_resolution: false,
        },
      },
    })

    const output = compiler.compile()
    const rootFile = output.files.find(f => f.relativePath === 'root.md')

    expect(rootFile).toBeDefined()
    if (rootFile) {
      // These fragments should not appear when conditionals are false
      expect(rootFile.content).not.toContain('decision-protocol')
      expect(rootFile.content).not.toContain('agent-harness')
    }
  })

  it('writes real compiled output to disk', () => {
    const compiler = new TemplateCompiler({
      libraryDir,
      outputDir: tempDir,
      tool: 'claude-code',
      context: {
        projectName: 'WriteTest',
        planningDir: 'planning',
        features: {
          tree_of_thoughts: true,
        },
      },
    })

    compiler.compileAndWrite()

    const rootFile = path.join(tempDir, 'root.md')
    const content = readFileSync(rootFile, 'utf-8')
    expect(content).toContain('WriteTest')
  })
})
