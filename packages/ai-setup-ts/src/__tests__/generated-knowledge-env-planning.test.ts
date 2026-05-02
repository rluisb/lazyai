import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { findMonorepoLibraryDir } from './test-helpers.js'

describe('generated knowledge and environment-aware planning guidance', () => {
  const libraryDir = findMonorepoLibraryDir()
  const reasoningProtocol = readFileSync(
    path.join(libraryDir, 'fragments', 'reasoning-protocol.md'),
    'utf8',
  )
  const plannerSkill = readFileSync(path.join(libraryDir, 'skills', 'plan.md'), 'utf8')
  const combinedPlanningGuidance = `${reasoningProtocol}\n${plannerSkill}`

  it('requires a concise bounded Knowledge Surface for non-trivial planning', () => {
    expect(reasoningProtocol).toContain('## Knowledge Surface')
    expect(reasoningProtocol).toMatch(/non-trivial/i)
    expect(reasoningProtocol).toMatch(/concise|bounded/i)
    expect(reasoningProtocol).toMatch(/Facts/i)
    expect(reasoningProtocol).toMatch(/Constraints/i)
    expect(reasoningProtocol).toMatch(/Assumptions/i)
    expect(reasoningProtocol).toMatch(/Unknowns/i)
    expect(reasoningProtocol).toMatch(/Evidence sources/i)
    expect(reasoningProtocol).toMatch(/cite|evidence/i)
  })

  it('requires an Environment Snapshot with verified and unverified assumption labels', () => {
    expect(plannerSkill).toContain('## Environment Snapshot')
    expect(plannerSkill).toMatch(/toolchain/i)
    expect(plannerSkill).toMatch(/package manager/i)
    expect(plannerSkill).toMatch(/CI\/check latency/i)
    expect(plannerSkill).toMatch(/platform/i)
    expect(plannerSkill).toMatch(/budget\/token constraints/i)
    expect(plannerSkill).toMatch(/network\/secrets constraints/i)
    expect(plannerSkill).toMatch(/verified assumptions/i)
    expect(plannerSkill).toMatch(/unverified assumptions/i)
  })

  it('does not promise deferred Wave 3 automation in planning content', () => {
    expect(combinedPlanningGuidance).not.toMatch(/\bRAG\b/i)
    expect(combinedPlanningGuidance).not.toMatch(/model routing/i)
    expect(combinedPlanningGuidance).not.toMatch(/provider billing/i)
    expect(combinedPlanningGuidance).not.toMatch(/automatic retrieval/i)
  })
})
