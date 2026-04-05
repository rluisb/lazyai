import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { AgentGenerator } from '../generators/agent.js'
import { CommandGenerator } from '../generators/command.js'
import { PromptGenerator } from '../generators/prompt.js'
import { GeneratorRegistry } from '../generators/registry.js'
import { SkillGenerator } from '../generators/skill.js'
import { TemplateGenerator } from '../generators/template.js'
import { WorkflowGenerator } from '../generators/workflow.js'

function makeTempDir(prefix: string): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), prefix))
}

function seedLibrary(targetDir: string): void {
  fs.mkdirSync(path.join(targetDir, 'library/agents'), { recursive: true })
  fs.mkdirSync(path.join(targetDir, 'library/skills'), { recursive: true })
  fs.mkdirSync(path.join(targetDir, 'library/prompts'), { recursive: true })
  fs.mkdirSync(path.join(targetDir, 'library/templates'), { recursive: true })

  fs.writeFileSync(path.join(targetDir, 'library/agents/scout.md'), '# scout', 'utf-8')
  fs.writeFileSync(path.join(targetDir, 'library/skills/research.md'), '# research', 'utf-8')
  fs.writeFileSync(path.join(targetDir, 'library/prompts/plan.md'), '# plan', 'utf-8')
  fs.writeFileSync(path.join(targetDir, 'library/templates/task.md'), '# task', 'utf-8')
}

describe('generators', () => {
  it('all generators expose correct type property', () => {
    expect(new AgentGenerator().type).toBe('agent')
    expect(new SkillGenerator().type).toBe('skill')
    expect(new CommandGenerator().type).toBe('command')
    expect(new PromptGenerator().type).toBe('prompt')
    expect(new TemplateGenerator().type).toBe('template')
    expect(new WorkflowGenerator().type).toBe('workflow')
  })

  it('all generators return prompt questions', () => {
    const generators = [
      new AgentGenerator(),
      new SkillGenerator(),
      new CommandGenerator(),
      new PromptGenerator(),
      new TemplateGenerator(),
      new WorkflowGenerator(),
    ]

    for (const generator of generators) {
      expect(generator.getPromptQuestions().length).toBeGreaterThan(0)
    }
  })

  it('each generator produces expected file path', async () => {
    const targetDir = makeTempDir('ai-setup-generators-paths-')
    seedLibrary(targetDir)

    const cases = [
      { generator: new AgentGenerator(), expected: 'library/agents/my-agent.md' },
      { generator: new SkillGenerator(), expected: 'library/skills/my-skill.md' },
      { generator: new CommandGenerator(), expected: 'src/commands/my-command.ts' },
      { generator: new PromptGenerator(), expected: 'library/prompts/my-prompt.md' },
      { generator: new TemplateGenerator(), expected: 'library/templates/my-template.md' },
      { generator: new WorkflowGenerator(), expected: 'library/workflows/my-workflow.md' },
    ]

    for (const testCase of cases) {
      const files = await testCase.generator.generate({
        name: testCase.expected.split('/').pop()?.replace(/\.(md|ts)$/, '') ?? 'name',
        targetDir,
      })
      expect(files[0]?.path).toBe(testCase.expected)
    }
  })

  it('agent generator emits frontmatter', async () => {
    const targetDir = makeTempDir('ai-setup-agent-generator-')
    const files = await new AgentGenerator().generate({
      name: 'builder',
      targetDir,
      answers: { model: 'gpt-4o', mode: 'interactive' },
    })

    const content = files[0]?.content ?? ''
    expect(content.startsWith('---')).toBe(true)
    expect(content).toContain('name: Builder')
    expect(content).toContain('model: gpt-4o')
    expect(content).toContain('mode: interactive')
    expect(content).toContain('## Identity')
  })

  it('skill generator emits command and goal sections', async () => {
    const targetDir = makeTempDir('ai-setup-skill-generator-')
    const files = await new SkillGenerator().generate({
      name: 'implement',
      targetDir,
      answers: { command: 'implement' },
    })

    const content = files[0]?.content ?? ''
    expect(content).toContain('**Command:** /implement [args]')
    expect(content).toContain('**Goal:**')
    expect(content).toContain('## Workflow')
  })

  it('command generator emits valid TypeScript scaffold', async () => {
    const targetDir = makeTempDir('ai-setup-command-generator-')
    const files = await new CommandGenerator().generate({
      name: 'sync-data',
      targetDir,
      answers: { arguments: '[name]' },
    })

    const content = files[0]?.content ?? ''
    expect(content).toContain("import type { Command } from 'commander'")
    expect(content).toContain('export function registerSyncData')
    expect(content).toContain(".command('sync-data [name]')")
  })

  it('prompt generator emits task and instruction sections', async () => {
    const targetDir = makeTempDir('ai-setup-prompt-generator-')
    const files = await new PromptGenerator().generate({
      name: 'implement',
      targetDir,
    })

    const content = files[0]?.content ?? ''
    expect(content).toContain('**Task:**')
    expect(content).toContain('## Instructions')
    expect(content).toContain('## Output Format')
  })

  it('template generator emits standard sections', async () => {
    const targetDir = makeTempDir('ai-setup-template-generator-')
    const files = await new TemplateGenerator().generate({
      name: 'task',
      targetDir,
      answers: { sections: 'Objective,Subtasks,Done When', fields: 'Phase,Status' },
    })

    const content = files[0]?.content ?? ''
    expect(content).toContain('**Phase:** [value]')
    expect(content).toContain('## Objective')
    expect(content).toContain('## Subtasks')
    expect(content).toContain('## Done When')
  })

  it('workflow generator emits step references', async () => {
    const targetDir = makeTempDir('ai-setup-workflow-generator-')
    seedLibrary(targetDir)

    const files = await new WorkflowGenerator().generate({
      name: 'release-flow',
      targetDir,
      answers: {
        steps: ['Research:agent=scout,skill=research', 'Plan:prompt=plan,template=task'],
      },
    })

    const content = files[0]?.content ?? ''
    expect(content).toContain('1. **Research**')
    expect(content).toContain('../agents/scout.md')
    expect(content).toContain('../skills/research.md')
    expect(content).toContain('../prompts/plan.md')
    expect(content).toContain('../templates/task.md')
  })

  it('generator registry returns all six types', () => {
    const registry = new GeneratorRegistry()
    expect(registry.getTypes().sort()).toEqual(['agent', 'command', 'prompt', 'skill', 'template', 'workflow'])
    expect(registry.get('agent')).toBeTruthy()
    expect(registry.get('workflow')).toBeTruthy()
  })
})
