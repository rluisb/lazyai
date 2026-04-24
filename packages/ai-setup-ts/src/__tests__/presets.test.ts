import { describe, expect, it } from 'vitest'
import { rulesForPreset, specsDirsForPreset, templatesForPreset } from '../presets.js'

describe('preset scaffold mappings', () => {
  it('returns the agreed specs directories for each preset', () => {
    expect(specsDirsForPreset('minimal')).toEqual(['standards', 'memory'])
    expect(specsDirsForPreset('standard')).toEqual([
      'features',
      'bugfixes',
      'rules',
      'adrs',
      'standards',
      'templates',
      'memory',
    ])
    expect(specsDirsForPreset('full')).toEqual([
      'features',
      'bugfixes',
      'refactors',
      'tech-debt',
      'adrs',
      'memory',
      'prompts',
      'standards',
      'templates',
      'rules',
    ])
    expect(specsDirsForPreset('custom')).toEqual(specsDirsForPreset('full'))
  })

  it('returns filtered templates and rules per preset', () => {
    expect(templatesForPreset('minimal')).toEqual([])
    expect(templatesForPreset('standard')).toEqual([
      'plan-template',
      'spec-template',
      'task',
      'adr',
      'bugfix-rca-template',
      'standard',
      'checklist-template',
    ])
    expect(templatesForPreset('full')).toEqual([
      'plan-template',
      'spec-template',
      'task',
      'adr',
      'bugfix-rca-template',
      'standard',
      'checklist-template',
      'code-review-template',
      'postmortem-template',
      'tech-debt-template',
    ])

    expect(rulesForPreset('minimal')).toEqual([])
    expect(rulesForPreset('standard')).toEqual([
      'code-style',
      'testing',
      'security',
      'workflow',
      'access',
    ])
    expect(rulesForPreset('full')).toEqual([
      'access',
      'agent-security',
      'code-style',
      'cost',
      'review',
      'security',
      'testing',
      'tool-use',
      'workflow',
    ])
    expect(rulesForPreset('custom')).toEqual(rulesForPreset('full'))
  })
})
