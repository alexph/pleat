---
name: principle-boundary-discipline
description: Apply when wiring validation, error handling, or framework adapters. Establish trust at boundaries, preserve internal guarantees, and keep transport details separate from domain decisions.
disable-model-invocation: true
---

# Boundary discipline

Validate external data where it enters and convert it into meaningful internal
types. Keep transport and framework details close to their adapters. Give
business rules an identifiable owner.

- Identify the trust boundary and the guarantees needed by its consumers. Parse
  raw inputs and reject unsupported values before internal logic relies on them.
- Use domain concepts across internal interfaces. Avoid making callers depend
  on wire formats, request objects, or storage representations without a reason.
- Inside a validated path, rely on guarantees that still hold. Avoid repeating
  shape checks or swallowing errors that indicate a broken internal contract.
- Check again when data crosses another trust boundary or a guarantee can change.
  Authorization, concurrent updates, and historical persisted data may require
  checks at the point of use.
- Keep framework wiring mechanical. Separate business decisions from I/O where
  doing so makes their inputs, effects, and behavior clearer to verify.

Input validation does not replace domain rules or atomic enforcement of mutable
invariants. Keep each rule at the owner that can enforce it correctly. Do not
extract every expression into a pure helper or add layers merely to make a
diagram look separated.
