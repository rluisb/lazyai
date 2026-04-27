# [YOUR_PROJECT_NAME] Uncertainty Markers

## Confidence Levels
- **HIGH** (90%+): Well-understood, proven approach
- **MEDIUM** (60-89%): Reasonable approach, some unknowns
- **LOW** (<60%): Experimental, needs validation

## When to Use Markers
- Architectural decisions with trade-offs
- Performance assumptions without benchmarks
- Third-party integration behavior
- Edge cases not covered by tests

## Marker Format
Use inline markers in specifications and code comments:
- `[CONFIDENCE: HIGH]` — proceed with implementation
- `[CONFIDENCE: MEDIUM]` — implement but plan validation
- `[CONFIDENCE: LOW]` — spike/prototype before committing

<!-- Example usage in a specification:
"The cache invalidation strategy uses TTL-based expiry. [CONFIDENCE: HIGH]
We assume Redis cluster failover takes <5 seconds. [CONFIDENCE: MEDIUM]
The WebSocket reconnection will handle 10K concurrent users. [CONFIDENCE: LOW]"
-->

## Resolution
Low-confidence items must be resolved (upgraded or abandoned) before merging to main.
