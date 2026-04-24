import { describe, expect, it } from 'vitest'
import {
  ensureModeAgentFrontmatter,
  extractTools,
  normalizeToolsFrontmatter,
  parseAgentFrontmatter,
  stripFrontmatterAndInjectModel,
  stripYamlFrontmatter,
} from '../utils/frontmatter.js'

describe('stripYamlFrontmatter', () => {
  it('removes leading YAML frontmatter when present', () => {
    const content = `---
title: Example
mode: user
---

# Hello\n`

    expect(stripYamlFrontmatter(content)).toBe('# Hello\n')
  })

  it('returns original content when no frontmatter exists', () => {
    const content = '# Hello\nmode: user\n'
    expect(stripYamlFrontmatter(content)).toBe(content)
  })

  it('only strips frontmatter at start of file', () => {
    const content = `# Intro

---
mode: user
---
`

    expect(stripYamlFrontmatter(content)).toBe(content)
  })
})

describe('ensureModeAgentFrontmatter', () => {
  it('adds frontmatter when missing', () => {
    const content = '# Skill\nDo thing.\n'

    expect(ensureModeAgentFrontmatter(content)).toBe(`---
mode: agent
---

# Skill
Do thing.
`)
  })

  it('replaces an existing top-level mode key', () => {
    const content = `---
title: Skill
mode: user
---

Body\n`

    expect(ensureModeAgentFrontmatter(content)).toBe(`---
title: Skill
mode: agent
---

Body
`)
  })

  it('appends mode key when frontmatter exists but mode is missing', () => {
    const content = `---
title: Skill
owner: team
---

Body\n`

    expect(ensureModeAgentFrontmatter(content)).toBe(`---
title: Skill
owner: team
mode: agent
---

Body
`)
  })

  it('does not rewrite mode-like text in YAML block scalar strings', () => {
    const content = `---
title: Skill
description: |
  preserve this line
  mode: keep-as-text
---

Body\n`

    expect(ensureModeAgentFrontmatter(content)).toBe(`---
title: Skill
description: |
  preserve this line
  mode: keep-as-text
mode: agent
---

Body
`)
  })
})

describe('extractTools', () => {
  it('parses a space-separated tools line', () => {
    const content = `---
name: Orchestrator
tools: list_catalog compose_agent start_chain
---

body`
    expect(extractTools(content)).toEqual(['list_catalog', 'compose_agent', 'start_chain'])
  })

  it('parses a comma-separated tools line', () => {
    const content = `---
name: Orchestrator
tools: list_catalog, compose_agent, start_chain
---

body`
    expect(extractTools(content)).toEqual(['list_catalog', 'compose_agent', 'start_chain'])
  })

  it('parses a YAML list form', () => {
    const content = `---
name: Orchestrator
tools:
  - list_catalog
  - compose_agent
  - start_chain
---

body`
    expect(extractTools(content)).toEqual(['list_catalog', 'compose_agent', 'start_chain'])
  })

  it('returns an empty array when the tools key is missing', () => {
    const content = `---
name: Orchestrator
model: opus
---

body`
    expect(extractTools(content)).toEqual([])
  })

  it('returns an empty array when the tools value is empty', () => {
    const content = `---
name: Orchestrator
tools:
---

body`
    expect(extractTools(content)).toEqual([])
  })

  it('returns an empty array when no frontmatter exists', () => {
    expect(extractTools('# Plain markdown\n')).toEqual([])
  })

  it('ignores extra whitespace between tool names', () => {
    const content = `---
tools:    a   b    c
---
body`
    expect(extractTools(content)).toEqual(['a', 'b', 'c'])
  })
})

describe('parseAgentFrontmatter', () => {
  it('extracts name, model, description, and tools', () => {
    const content = `---
name: Orchestrator
model: opus
description: Coordinates chains
tools: start_chain advance_chain
---

body`
    expect(parseAgentFrontmatter(content)).toEqual({
      name: 'Orchestrator',
      model: 'opus',
      description: 'Coordinates chains',
      tools: ['start_chain', 'advance_chain'],
    })
  })

  it('returns null when no frontmatter exists', () => {
    expect(parseAgentFrontmatter('body only')).toBeNull()
  })

  it('omits tools when the tools key is absent', () => {
    const content = `---
name: Agent
model: opus
---

body`
    const parsed = parseAgentFrontmatter(content)
    expect(parsed?.tools).toBeUndefined()
  })
})

describe('normalizeToolsFrontmatter', () => {
  it('rewrites a space-separated tools line to comma form', () => {
    const content = `---
name: Orchestrator
model: opus
tools: start_chain advance_chain get_status
---

body`
    const out = normalizeToolsFrontmatter(content, 'comma')
    expect(out).toContain('tools: start_chain, advance_chain, get_status')
    expect(out).toContain('name: Orchestrator')
    expect(out).toContain('model: opus')
    expect(out).toMatch(/\nbody$/)
  })

  it('rewrites a YAML list form to a single space-separated line', () => {
    const content = `---
name: Orchestrator
tools:
  - a
  - b
  - c
---

body`
    const out = normalizeToolsFrontmatter(content, 'space')
    expect(out).toContain('tools: a b c')
    expect(out).not.toContain('- a')
  })

  it('is a no-op when tools key is missing', () => {
    const content = `---
name: Orchestrator
---

body`
    expect(normalizeToolsFrontmatter(content, 'comma')).toBe(content)
  })

  it('is a no-op when no frontmatter exists', () => {
    expect(normalizeToolsFrontmatter('body', 'comma')).toBe('body')
  })
})

describe('stripFrontmatterAndInjectModel with tools', () => {
  it('emits an allowed-tools comment when tools are declared', () => {
    const content = `---
name: Orchestrator
model: opus
tools: start_chain advance_chain
---

# Body`
    const out = stripFrontmatterAndInjectModel(content)
    expect(out).toContain('<!-- Recommended model: opus -->')
    expect(out).toContain('<!-- allowed-tools: start_chain, advance_chain -->')
    expect(out).toContain('# Body')
  })

  it('omits the allowed-tools comment when tools are absent', () => {
    const content = `---
name: Agent
model: opus
---

# Body`
    const out = stripFrontmatterAndInjectModel(content)
    expect(out).not.toContain('allowed-tools')
    expect(out).toContain('<!-- Recommended model: opus -->')
  })

  it('emits only the allowed-tools comment when model is absent', () => {
    const content = `---
tools: a b
---

body`
    const out = stripFrontmatterAndInjectModel(content)
    expect(out).toContain('<!-- allowed-tools: a, b -->')
    expect(out).not.toContain('Recommended model')
  })
})
