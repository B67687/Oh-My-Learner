# ADR-001: SM-2 with Session-Level Interleaving

**STATUS:** Accepted

## Context
The scheduler needed to support both spaced repetition (proven algorithm) and interleaving (mixing subjects for long-term retention). Research showed SM-2 is the right choice for a new tool (works from day one, no training data needed) while FSRS requires 1000+ reviews.

## Decision
SM-2 for card-level scheduling (proven, ~80 lines of math) + session-level interleaving layer (randomize card order across subjects at session build time). The SM-2 algorithm is unchanged — interleaving is purely a session-level concern.

## Consequences
**Positive:** SM-2 works from the first review. No cold-start problem.
**Positive:** Interleaving is achieved without modifying the scheduling algorithm.
**Negative:** FSRS would be more efficient long-term (20-30% fewer reviews). SM-2 is a known ceiling. Upgrade path exists.
