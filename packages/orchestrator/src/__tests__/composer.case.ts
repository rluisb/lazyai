import { describe, expect, it } from 'vitest'
import { composeAgent } from '../composer.js'

describe('composeAgent', () => {
  it('intersects tools and merges prompt layers deterministically', () => {
    const result = composeAgent({
      root: {
        source: 'root',
        name: 'root',
        prompt: 'Root prompt',
        allowedTools: ['Read', 'Write', 'Bash'],
        constraints: ['stay scoped'],
      },
      base: {
        source: 'base',
        name: 'builder',
        prompt: 'Base prompt',
        allowedTools: ['Read', 'Write', 'Bash'],
        constraints: ['stay scoped', 'run tests'],
        modelHint: 'sonnet',
      },
      domain: {
        source: 'domain',
        name: 'typescript',
        prompt: 'Domain prompt',
        allowedTools: ['Read', 'Bash'],
        constraints: ['use types'],
        approvalPolicy: 'normal',
      },
      step: {
        source: 'step',
        name: 'repair',
        prompt: 'Step prompt',
        allowedTools: ['Read', 'Bash', 'Edit'],
        constraints: ['run tests'],
        approvalPolicy: 'strict',
      },
    })

    expect(result.tools).toEqual(['Bash', 'Read'])
    expect(result.constraints).toEqual(['stay scoped', 'run tests', 'use types'])
    expect(result.approvalPolicy).toBe('strict')
    expect(result.prompt).toBe('Root prompt\n\nBase prompt\n\nDomain prompt\n\nStep prompt')
  })
})
