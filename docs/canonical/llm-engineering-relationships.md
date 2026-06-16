# LLM Engineering Relationships

Raw model, CLI harness, project adapter, generated context, skill, agent, hook, and ai-memory are distinct layers.

- **Raw model** — runtime-enforced only by provider/system message mechanics.
- **CLI harness** — runtime-enforced execution boundary and tool surface.
- **Project adapter** — generated bridge for a specific CLI/runtime.
- **Generated context** — generated instructions; do not edit directly.
- **Skill** — advisory procedure loaded when relevant.
- **Agent** — advisory role prompt plus delegated scope.
- **Hook** — runtime-enforced policy or automation at configured events.
- **ai-memory** — advisory retrieved data unless explicitly system-authored.
- **Human gate** — human-gated approval or rejection of high-impact actions.
