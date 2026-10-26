---
name: principle-model-the-domain
description: Apply when designing stateful logic, related flags, or repeated shape-dependent branches. Represent domain concepts, valid relationships, and transitions with types and data structures that fit the required behavior and access patterns.
disable-model-invocation: true
---

# Model the domain

Represent the concepts, states, and transitions the system actually works with.
Choose types and data structures that make valid relationships clear and invalid
combinations difficult to express.

- Identify the domain concepts and invariants before adding logic. Distinguish
  stored facts, derived values, and states that can change.
- Trace the dominant reads, writes, lookups, and transitions through the proposed
  shape. Choose structures that support those actual paths. Use measurements
  when performance costs affect the decision. Avoid speculative indexes, caches,
  or alternative representations for access patterns no current caller needs.
- When flags allow contradictory combinations, consider named states or a
  discriminated union. Make required transition rules explicit where they matter.
- When several branches reconstruct the same meaning, consider representing that
  meaning directly. A typed record, collection, table, or reducer may be enough.
- Use domain terms in the internal model. Keep transport and storage details in
  adapters unless those representations already express the domain clearly.
- Validate runtime inputs where static types cannot establish the invariant.
  Use constructors or constraints when they provide a real guarantee.

Keep the model no more elaborate than the behavior requires. Do not introduce a
class hierarchy, generic engine, or state-machine library when an ordinary type
and direct control flow express the rules. Prefer evidence from current callers
over assumptions about hypothetical future domains.
