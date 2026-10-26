---
name: bench-architect
description: Design changes to contracts, ownership, state, or data flow across function or module boundaries before implementation. Use for architecture requests or when Bench routes such changes here. Ground the problem, sketch caller usage and types, explore meaningful alternatives in parallel, and reconcile the design during implementation. Routine calls that preserve existing contracts do not need this workflow.
---

# Bench architect

Produce a design concrete enough to expose consequential mistakes before
implementation. Keep the sketch provisional as evidence improves.

## Establish the scope

Use this workflow when a change needs a decision about interfaces, ownership,
representation, or coordination across boundaries. Calling an existing function
does not by itself require architecture work. For a routine change that preserves
those contracts, keep the direct implementation and relevant checks.

Respect the requested deliverable. For design-only work, return the design without
editing application code. For an implementation task, design and then continue
without adding an approval checkpoint. Honor an explicit request to stop at the
design. Ask only about consequential choices that evidence and authorization
cannot settle.

## Ground the decision

Read [Understand the promise](../principle-understand-the-promise/SKILL.md) and
[Blast radius](../principle-blast-radius/SKILL.md) when establishing the scope.
Inspect the real caller, current path, affected consumers, and relevant project
constraints. Explain the part of the existing system that constrains this change.
Name sources and distinguish observed behavior from assumptions.

For a new subsystem, inspect the surrounding integration contracts. Stop expanding
the investigation once the relevant owners, invariants, and consumers are clear.

## Sketch from the caller

Show a realistic caller using the proposed interface before deriving its internal
shape. Trace its inputs through the proposed operations to the observable result.
Include a meaningful failure path where it changes the design.

Apply the relevant principles to the sketch.

- [Model the domain](../principle-model-the-domain/SKILL.md) for types, valid states,
  and transitions.
- [Give responsibilities a home](../principle-give-responsibilities-a-home/SKILL.md)
  for ownership of decisions, state, and effects.
- [Boundary discipline](../principle-boundary-discipline/SKILL.md) for validation,
  adapters, and guarantees that can change.

Use types, signatures, and a module map only where they clarify a decision. Keep
pseudocode and unimplemented bodies in the design note or an isolated scratch
area. Do not scatter placeholders through application code.

## Explore and challenge

When the design has meaningful alternatives, explore them in parallel before
implementation. Start with two independent candidate designers when delegation is
available. Include a direct change to the existing design when it is viable.

Give each designer the same task, constraints, source paths, and output needs.
Do not give them a preferred answer or another candidate's conclusions. Ask for
caller usage, the proposed shape, rationale, and a consequential risk or unresolved
assumption. Keep candidates read-only, or give them separate scratch locations.

Tell delegated designers to return one candidate only. They must not implement,
coordinate other designers, or repeat this fan-out. The coordinating agent owns
comparison, selection, and any subsequent implementation.

Compare genuinely different choices about ownership, representation, or data flow.
Do not manufacture competitors when the constraints settle the shape. If
candidates converge, record that and challenge their shared assumptions. Agreement
does not establish correctness. When delegation is unavailable, compare viable
alternatives sequentially and report that limit.

Read [the design checks](references/design-checks.md) before selecting a shape.
Use [Laziness](../principle-laziness/SKILL.md) and
[Minimise reader load](../principle-minimise-reader-load/SKILL.md) to weigh total
complexity and caller effort. Prefer one coherent design over a mixture that
inherits incompatible assumptions from several candidates.

Use [Seek evidence that could prove you wrong](../principle-seek-evidence-that-could-prove-you-wrong/SKILL.md)
to resolve consequential factual uncertainty. A bounded experiment may settle an
unknown runtime behavior or access pattern. Do not treat a design vote as that
evidence. Stop exploration when the design can be implemented and remaining
uncertainty has an explicit verification path.

## Record and reconcile

Use [the design note](references/design-note.md) to capture the chosen shape and
its rationale. A short answer can contain a small design. For a larger task or a
handoff, update the task's existing plan or design document with the sketch and
decisions. Avoid a second document describing a competing current design.

For authorized implementation, build against the chosen contracts and verify the
caller behavior. Update the note when evidence changes a consequential decision.
A local signature refinement need not restart exploration. Revisit the design
when repeated workarounds, escaping types, or conflicting owners expose a shared
false assumption. Remove only task-scoped obsolete structure and preserve project
retention rules. Re-explore the affected decision rather than restarting every
phase by default.

Return the recommendation or implemented result, the important tradeoffs, and the
evidence and remaining limits. Link a saved design note when one exists. Describe
proposed behavior as proposed until implementation and verification demonstrate it.
