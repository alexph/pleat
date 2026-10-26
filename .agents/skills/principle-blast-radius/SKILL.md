---
name: principle-blast-radius
description: Apply when changing shared code, contracts, permissions, stored data, or asynchronous effects. Identify affected consumers and failure consequences, then scale verification and delivery to that reach.
disable-model-invocation: true
---

# Blast radius

Understand who and what can be affected if a change is wrong. Judge reach by
dependencies and consequences, not diff size.

- Inspect consumers of shared behavior and contracts, including callers outside
  the feature being edited. Identify which behavior they rely on.
- Follow effects on stored data, access, external systems, and work already in
  progress when the change touches them. For retried or asynchronous work,
  consider partial completion and repeated effects.
- Prefer containing effects within the intended scope. Use the correct owner
  even when correcting it requires updating several callers.
- Match verification to the plausible consequences. A shared permission change
  needs checks for access gained and lost. An isolated presentation change needs
  different evidence.
- For a broad or persistent change, choose a delivery approach that can expose
  mistakes before they spread and support recovery where practical. Identify
  irreversible effects rather than assuming a code rollback reverses them.

Keep investigation proportional to the reachable behavior. Do not preserve a
faulty shared design merely to keep a patch local. Planning delivery does not
grant permission to deploy, delete data, or mutate external systems.
