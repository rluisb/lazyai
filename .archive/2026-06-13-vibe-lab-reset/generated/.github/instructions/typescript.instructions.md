---
applyTo: "**/*.ts,**/*.tsx"
description: TypeScript conventions and best practices for this codebase
---

# TypeScript Conventions

## Type discipline

- Use strict null checks. Never use `any` — use `unknown` + type guard if truly dynamic.
- Prefer `type` over `interface` when defining union types or complex shapes.
- Use explicit return types on public functions — omit only for obvious short implementations.
- Avoid optional properties with undefined default; use either required or discriminated unions.

## Imports and organization

- Sort imports: standard library → local absolute → relative. Single import per file from same module.
- Use absolute imports (`import { X } from '@/src/...'`) when available; relative imports for adjacent files.
- Group related imports and separate with blank lines.

## Testing

- Write tests with Vitest or Jest — use `describe` + `it` blocks.
- Test filenames match source: `foo.ts` → `foo.test.ts` or `foo.spec.ts`.
- Aim for >80% coverage on critical paths; ignore utilities tested indirectly.
- Use factories and fixtures over test data singletons to avoid cross-test coupling.

## Error handling

- Throw only `Error` instances or subclasses; never throw primitives.
- Include context in error messages — what was being done, on what data, why it failed.
- Use error boundaries at UI layer; let errors bubble in data/business logic layers.

## Performance

- Avoid premature optimization — profile first.
- Use `const` by default; `let` only when reassignment needed (rarely).
- Lazy-load modules for expensive initialization; document why.
