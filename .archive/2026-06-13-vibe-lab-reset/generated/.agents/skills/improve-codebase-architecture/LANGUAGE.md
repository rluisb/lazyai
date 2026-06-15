# Architecture Vocabulary

Glossary of terms used in improve-codebase-architecture. Use these terms exactly — consistent language is the point.

## Module
Anything with an interface and an implementation: function, class, package, slice.

## Interface
Everything a caller must know to use the module: types, invariants, error modes, ordering, config. Not just the type signature.

## Implementation
The code inside a module.

## Depth
Leverage at the interface: a lot of behaviour behind a small interface. Deep = high leverage. Shallow = interface nearly as complex as the implementation.

## Seam
Where an interface lives — a place behaviour can be altered without editing in place. Use this, not "boundary."

## Adapter
A concrete thing satisfying an interface at a seam.

## Leverage
What callers get from depth — the benefit of a deep module.

## Locality
What maintainers get from depth: change, bugs, knowledge concentrated in one place.

## Deletion Test
Imagine deleting the module. If complexity vanishes, it was a pass-through (shallow). If complexity reappears across N callers, it was earning its keep (deep). Ask: "does deleting this concentrate or distribute complexity?"

## Interface Is the Test Surface
The interface is what you test — not the implementation details.

## One Adapter = Hypothetical Seam, Two Adapters = Real Seam
One adapter = you could swap the implementation. Two adapters = there is actually a seam you can exploit.