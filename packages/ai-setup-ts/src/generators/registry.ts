import type { ArtifactType } from '../types.js'
import { AgentGenerator } from './agent.js'
import { CommandGenerator } from './command.js'
import { DomainGenerator } from './domain.js'
import { ModeGenerator } from './mode.js'
import { PromptGenerator } from './prompt.js'
import { SkillGenerator } from './skill.js'
import { TemplateGenerator } from './template.js'
import type { Generator } from './types.js'
import { WorkflowGenerator } from './workflow.js'

export class GeneratorRegistry {
  private generators: Map<ArtifactType, Generator> = new Map()

  constructor() {
    const builtins: Generator[] = [
      new AgentGenerator(),
      new SkillGenerator(),
      new CommandGenerator(),
      new PromptGenerator(),
      new TemplateGenerator(),
      new WorkflowGenerator(),
      new DomainGenerator(),
      new ModeGenerator(),
    ]

    for (const generator of builtins) {
      this.register(generator)
    }
  }

  register(generator: Generator): void {
    this.generators.set(generator.type, generator)
  }

  get(type: ArtifactType): Generator | undefined {
    return this.generators.get(type)
  }

  getTypes(): ArtifactType[] {
    return Array.from(this.generators.keys())
  }
}
