# Spec Kit Vibe-Lab Preset

A Spec Kit-native preset that shapes Spec Kit-generated artifacts around vibe-lab's four-point operating principles.

## Install

1. Initialize Spec Kit in your project normally:
   ```bash
   specify init
   ```

2. Add this preset:
   ```bash
   specify preset add --dev <path-to-canonical/speckit-vibe-lab-preset>
   ```

3. Verify resolution:
   ```bash
   specify preset resolve spec-template
   ```

## Remove

```bash
specify preset remove speckit-vibe-lab-preset
```

## What it changes

- **Templates** — overrides default Spec Kit templates with vibe-lab-shaped versions
  that encode four-point framing: WHAT / HOW / DON'T WANT / VALIDATE.
- **Command text** — overrides the core `speckit.constitution`,
  `speckit.specify`, `speckit.plan`, `speckit.tasks`, and `speckit.checklist`
  prompts so installed command files follow anti-speculation, simplest-viable,
  test-first, and verification-truth rules.

## What it does NOT do

- Does not ship an extension, hook layer, workflow runner, ledger, or agent pack.
- Does not take ownership of this repo's `CLAUDE.md`, `AGENTS.md`, `.claude/`,
  `.opencode/`, `.pi/`, `.agents/`, or `canonical/` surfaces.
- Does not claim live OpenCode or Pi command behavior beyond what is directly
  verified in isolated fixtures.

## Validation

This preset is validated only in isolated fixture projects. Current direct
verification covers:

- template resolution through `specify preset resolve`
- command override generation in an isolated Claude fixture after
  `specify preset add --dev`

It does not claim catalog publication, extension support, or wider runtime
behavior without separate proof.
