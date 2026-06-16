# Engineering Principles for Agents

> Extract the useful parts of engineering doctrine without importing the ceremony.

## First Rule

Respect the project's existing architecture, paradigm, and naming before applying a generic practice.

## Simplicity and Scope

- Implement the smallest change that satisfies current callers, tests, and constraints.
- Prefer a simple concrete solution first; generalize only after repeated pressure.
- Prefer duplication over an abstraction that hides intent or couples unrelated cases.
- Every new layer, dependency, or configuration path must pay for itself immediately.


## Data, State, and Effects

- Prefer pure transformations when they make the code easier to test and read.
- Keep side effects at boundaries: IO, network, database, filesystem, time, and randomness.
- Make state ownership explicit; avoid hidden mutation across module or object boundaries.
- Use named intermediate values when a clever pipeline would hide intent.

## Objects and Cohesion

- Put behavior near the data or invariant it owns.
- Constructors, factories, and parsers should prevent invalid states when practical.
- Avoid empty pass-through classes or modules named `Manager`, `Service`, `Helper`, or `Util` unless the project already uses that convention.
- Do not force classes into a functional codebase or pure functions into an object-oriented codebase.

## Contracts and Boundaries

- Preserve caller expectations when changing implementations.
- Expose the smallest interface the caller needs.
- Depend on stable contracts at external boundaries, not concrete churn inside a subsystem.
- Introduce an abstraction only when there is a real boundary, second implementation, or test seam.

## Patterns and Domain Modeling

- Use existing project patterns before introducing a new named pattern.
- Name a design pattern only when the simple implementation already has that shape.
- A pattern must remove more complexity than it adds.
- Use domain language for names at boundaries and in business rules.
- Keep domain rules near the domain or use-case boundary instead of scattering them through UI, controllers, or scripts.
- Use DDD ceremony only when the project's current domain complexity justifies it.

## Tools and Composition

- For scripts and small tools, prefer one clear purpose over a grab-bag command surface.
- Prefer simple composition through files, arguments, environment, or standard streams when it keeps the tool easy to replace.
- Do not force Unix-style text pipelines onto code that needs stronger typed or domain-specific boundaries.


## Rejections

- Do not add standalone paradigm rewrites unless explicitly requested.
- Do not pattern-shop.
- Do not add generic layers because a practice says they are common.
- Do not treat SOLID, OOP, FP, Design Patterns, or DDD as mandatory architecture.
