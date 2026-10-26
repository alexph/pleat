---
name: principle-encode-lessons-in-structure
description: Apply when a failure recurs or the same corrective instruction is repeatedly needed. Prefer a type, interface, constraint, or focused check that prevents or exposes the failure when it justifies its maintenance cost.
disable-model-invocation: true
---

# Encode lessons in structure

When a failure recurs, look for a structural change that prevents or exposes it.
Reduce reliance on remembering the same instruction at every call site.

- Identify the repeated failure and the invariant it violates. Confirm that the
  examples share a cause before generalizing from them.
- Prefer an existing enforcement point. A domain type, constructor, database
  constraint, or supported interface may express the guarantee directly.
- Use a focused test, lint rule, generator, or script when it catches a concrete
  recurring mistake and is worth maintaining. Do not build tooling for every
  one-off correction.
- Keep guidance near the structure that owns the decision. Link to the source of
  truth instead of copying its schema or policy into several instruction files.
- Retire redundant instructions only after checking that the replacement covers
  their useful scope. Explain judgment and exceptions that the mechanism cannot
  express.

Structural enforcement is useful only when it matches the actual requirement.
Do not turn a contextual preference into a rigid constraint or expand the task
to introduce a framework. Keep prose guidance when the decision genuinely needs
case-by-case judgment.
