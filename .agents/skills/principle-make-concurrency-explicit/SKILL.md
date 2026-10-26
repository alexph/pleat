---
name: principle-make-concurrency-explicit
description: Apply when requests, jobs, callbacks, processes, or parallel agents can overlap while using mutable state or effects. Trace interleavings, prefer independent state where valid, and enforce shared invariants at the appropriate atomic boundary.
disable-model-invocation: true
---

# Make concurrency explicit

Identify actors that can overlap and the state or effects they can change. Trace
what happens when their operations interleave. One owner does not imply one
execution at a time.

- Identify which actors can read or write each relevant value and which invariant
  must hold across their operations. Include separate processes or machines when
  they participate in the actual path.
- Prefer immutable inputs or independent state when the domain permits it.
  Separate scratch files and per-task state before coordinating writers. Do not
  split a shared invariant into copies that cannot enforce it together.
- For shared authority, identify the atomic boundary that protects the invariant.
  Consider transactions, conditional updates, constraints, or explicit
  coordination according to the storage and execution model. A process-local
  lock cannot coordinate writers in other processes.
- Trace the interval between a read and the decision or write that relies on it.
  Check for stale assumptions, lost updates, repeated effects, and partial
  completion where those failures are plausible.
- Account for retries, cancellation, and cleanup when they can overlap with
  another actor's work. Establish which effects may repeat and who may release or
  remove shared resources. A database transaction does not automatically cover
  an external side effect.
- Verify the important interleavings using the relevant execution boundary.
  Use separate connections or processes when the claim depends on them. A
  sequential success or mocked call order does not demonstrate concurrency safety.

Keep coordination at the invariant's owner and match it to the real actors.
Do not introduce locks, queues, or distributed coordination for state that is not
shared. Isolation is a preference where valid, not a substitute for enforcing
invariants that genuinely require shared authority.
