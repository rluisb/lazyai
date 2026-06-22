# Skill Quality Guidelines

A high-quality skill should be clearly defined and validated. Skills are flat Markdown files with YAML frontmatter.

## Semantic Validation

During `lazyai-cli validate --all`, skills are checked for both structural and semantic quality.

### Structural Requirements (Errors)
- Must be a readable file.
- Must contain valid YAML frontmatter.
- Must declare a `name` and `description` in frontmatter.
- The body must not be empty.

### Quality Guidelines (Warnings)
Skills must contain key guidance sections in their markdown body:
- **Trigger guidance:** Explain when the skill should be activated (e.g. `When to use`).
- **Misuse guidance:** Explain when the skill should *not* be used (e.g. `When not to use`).
- **Evidence requirements:** State what proves the skill's work is done (e.g. `Required evidence`, `Done criteria`).
- **Output format:** State the expected deliverable format (e.g. `Expected output`).

Adhering to these guidelines ensures agents reliably trigger and fulfill the skill's purpose without unintended side effects.