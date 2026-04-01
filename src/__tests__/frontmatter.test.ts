import { describe, expect, it } from 'vitest'
import { ensureModeAgentFrontmatter, stripYamlFrontmatter } from '../utils/frontmatter.js'

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
