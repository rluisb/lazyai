import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@clack/prompts', () => ({
  select: vi.fn(),
  multiselect: vi.fn(),
  text: vi.fn(),
  confirm: vi.fn(),
  note: vi.fn(),
  cancel: vi.fn(),
  intro: vi.fn(),
  outro: vi.fn(),
  spinner: vi.fn(() => ({ start: vi.fn(), stop: vi.fn() })),
  isCancel: vi.fn(() => false),
}))

import * as p from '@clack/prompts'
import { runPhase2Features } from '../wizard/phase2-features.js'

const DEFAULT_FEATURES = {
  contextEngineering: true,
  rpiWorkflow: true,
  chainOfThought: true,
  treeOfThoughts: true,
  adrEnforcement: true,
  qualityGates: true,
  agentHarness: true,
  bugResolution: true,
  pivotHandling: true,
}

describe('phase2 features merge behavior', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(p.isCancel).mockReturnValue(false)
  })

  it('uses defaults when no prior values or CLI overrides are provided', async () => {
    const result = await runPhase2Features({ interactive: false })

    expect(result.planningDir).toBe('.planning')
    expect(result.features).toEqual(DEFAULT_FEATURES)
  })

  it('uses prior planningDir in non-interactive mode', async () => {
    const result = await runPhase2Features({
      interactive: false,
      prior: { planningDir: '.ai-planning' },
    })

    expect(result.planningDir).toBe('.ai-planning')
  })

  it('uses CLI planningDir over prior planningDir', async () => {
    const result = await runPhase2Features({
      interactive: false,
      prior: { planningDir: '.old-planning' },
      cliOverrides: { planningDir: '.new-planning' },
    })

    expect(result.planningDir).toBe('.new-planning')
  })

  it('applies prior feature values over defaults', async () => {
    const result = await runPhase2Features({
      interactive: false,
      prior: {
        features: {
          chainOfThought: false,
          treeOfThoughts: false,
        },
      },
    })

    expect(result.features.chainOfThought).toBe(false)
    expect(result.features.treeOfThoughts).toBe(false)
    expect(result.features.contextEngineering).toBe(true)
  })

  it('CLI --features enables a feature disabled by prior values', async () => {
    const result = await runPhase2Features({
      interactive: false,
      prior: { features: { cursor: false } as never },
      cliOverrides: { features: ['cursor'] },
    })

    expect((result.features as Record<string, boolean>).cursor).toBe(true)
  })

  it('CLI --disable-features disables a specific feature', async () => {
    const result = await runPhase2Features({
      interactive: false,
      cliOverrides: { disableFeatures: ['agentHarness'] },
    })

    expect(result.features.agentHarness).toBe(false)
    expect(result.features.contextEngineering).toBe(true)
  })

  it('CLI --disable-features all disables all default features', async () => {
    const result = await runPhase2Features({
      interactive: false,
      cliOverrides: { disableFeatures: ['all'] },
    })

    expect(Object.values(result.features).every(value => value === false)).toBe(true)
  })

  it('CLI --disable-features all does not crash when some features are already disabled', async () => {
    await expect(
      runPhase2Features({
        interactive: false,
        prior: { features: { qualityGates: false, agentHarness: false } },
        cliOverrides: { disableFeatures: ['all'] },
      }),
    ).resolves.toMatchObject({
      features: {
        contextEngineering: false,
        rpiWorkflow: false,
        chainOfThought: false,
        treeOfThoughts: false,
        adrEnforcement: false,
        qualityGates: false,
        agentHarness: false,
        bugResolution: false,
        pivotHandling: false,
      },
    })
  })

  it('CLI --disable-features all wins over CLI --features in same invocation', async () => {
    const result = await runPhase2Features({
      interactive: false,
      cliOverrides: {
        features: ['contextEngineering', 'qualityGates'],
        disableFeatures: ['all'],
      },
    })

    expect(Object.values(result.features).every(value => value === false)).toBe(true)
  })

  it('accepts unknown feature flags from prior values and keeps them', async () => {
    const result = await runPhase2Features({
      interactive: false,
      prior: { features: { cursor: false } as never },
    })

    expect((result.features as Record<string, boolean>).cursor).toBe(false)
  })

  it('enables unknown feature flags from CLI when they exist in merged features', async () => {
    const result = await runPhase2Features({
      interactive: false,
      prior: { features: { cursor: false } as never },
      cliOverrides: { features: ['cursor'] },
    })

    expect((result.features as Record<string, boolean>).cursor).toBe(true)
  })

  it('disables unknown feature flags from CLI when they exist in merged features', async () => {
    const result = await runPhase2Features({
      interactive: false,
      prior: { features: { cursor: true } as never },
      cliOverrides: { disableFeatures: ['cursor'] },
    })

    expect((result.features as Record<string, boolean>).cursor).toBe(false)
  })

  it('CLI unknown feature flags are ignored if they never existed in merged feature set', async () => {
    const result = await runPhase2Features({
      interactive: false,
      cliOverrides: {
        features: ['cursor'],
        disableFeatures: ['windsurf'],
      },
    })

    expect((result.features as Record<string, boolean>).cursor).toBeUndefined()
    expect((result.features as Record<string, boolean>).windsurf).toBeUndefined()
    expect(result.features.contextEngineering).toBe(true)
  })

  it('disable all also disables unknown feature keys introduced by prior values', async () => {
    const result = await runPhase2Features({
      interactive: false,
      prior: { features: { cursor: true } as never },
      cliOverrides: { disableFeatures: ['all'] },
    })

    expect((result.features as Record<string, boolean>).cursor).toBe(false)
  })

  it('interactive selections override prior values by explicit user selection', async () => {
    vi.mocked(p.text).mockResolvedValueOnce('.planning')
    vi.mocked(p.multiselect).mockResolvedValueOnce(['contextEngineering', 'qualityGates'])
    vi.mocked(p.select)
      .mockResolvedValueOnce('{type}/{ticket}-{description}')
      .mockResolvedValueOnce('{type}({scope}): {description}')
    vi.mocked(p.confirm).mockResolvedValueOnce(false)

    const result = await runPhase2Features({
      interactive: true,
      prior: { features: { contextEngineering: false, chainOfThought: true } },
    })

    expect(result.features.contextEngineering).toBe(true)
    expect(result.features.chainOfThought).toBe(false)
    expect(result.features.qualityGates).toBe(true)
  })

  it('interactive selections can disable defaults when user deselects them', async () => {
    vi.mocked(p.text).mockResolvedValueOnce('.planning')
    vi.mocked(p.multiselect).mockResolvedValueOnce(['contextEngineering'])
    vi.mocked(p.select)
      .mockResolvedValueOnce('{type}/{ticket}-{description}')
      .mockResolvedValueOnce('{type}({scope}): {description}')
    vi.mocked(p.confirm).mockResolvedValueOnce(false)

    const result = await runPhase2Features({ interactive: true })

    expect(result.features.contextEngineering).toBe(true)
    expect(result.features.rpiWorkflow).toBe(false)
    expect(result.features.agentHarness).toBe(false)
  })

  it('interactive flow uses selected features, not CLI overrides', async () => {
    vi.mocked(p.text).mockResolvedValueOnce('.planning')
    vi.mocked(p.multiselect).mockResolvedValueOnce(['qualityGates'])
    vi.mocked(p.select)
      .mockResolvedValueOnce('{type}/{ticket}-{description}')
      .mockResolvedValueOnce('{type}({scope}): {description}')
    vi.mocked(p.confirm).mockResolvedValueOnce(false)

    const result = await runPhase2Features({
      interactive: true,
      cliOverrides: {
        features: ['contextEngineering'],
        disableFeatures: ['all'],
      },
    })

    expect(result.features.qualityGates).toBe(true)
    expect(result.features.contextEngineering).toBe(false)
  })
})
