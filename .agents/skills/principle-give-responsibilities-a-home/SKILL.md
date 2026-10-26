---
name: principle-give-responsibilities-a-home
description: Apply when policy decisions, mutable state, or invariants are duplicated or passed through several layers. Identify the owner and keep enforcement and updates with the responsibility they serve.
disable-model-invocation: true
---

# Give responsibilities a home

Give each shared decision, piece of state, and invariant an identifiable owner.
Keep the code that enforces or changes it close to that owner.

- Trace who produces, changes, and consumes the value. Reuse an existing owner
  before adding a service, registry, or context object.
- Resolve a shared policy once where the necessary information and authority
  meet. Pass its result to consumers instead of repeating the decision.
- Keep a source of truth for mutable state. Derive other views when practical.
  Give necessary copies an explicit update and consistency contract.
- Put effects with the lifecycle that owns them. Before threading a flag or
  callback through layers, check whether the operation belongs elsewhere.
- Separate policies that can change independently. Similar code does not prove
  that two callers share a responsibility.

Ownership does not require a new class or module. A local function or existing
boundary may be enough. Keep authorization and state checks at each place where
their guarantees must hold. Sharing a policy does not make an earlier result
valid forever.
