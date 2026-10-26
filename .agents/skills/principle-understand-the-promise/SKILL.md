---
name: principle-understand-the-promise
description: Apply when starting a feature, interpreting an ambiguous request, or changing a public contract. Establish the caller's intended outcome, required constraints, and consequential assumptions before choosing an implementation.
disable-model-invocation: true
---

# Understand the promise

Establish what the caller should experience and what must remain true. Distinguish
the requested outcome from a suggested implementation. Use the outcome to judge
whether the work is complete.

- Read the current caller, contract, and relevant project guidance before choosing
  an approach. Check remembered claims against current evidence.
- Identify the observable success condition and the existing behavior the change
  must preserve. A caller may be a person, an API client, or another module.
- Resolve factual uncertainty by inspecting or running the system. Ask the user
  when a consequential product choice cannot be settled by evidence or existing
  authorization.
- For a minor, reversible ambiguity, choose a reasonable default within the
  request. State assumptions that materially affect the result.
- When the requested mechanism conflicts with the intended outcome, explain the
  conflict. Do not silently substitute a different product decision.

Scale this work to the uncertainty. A clear request may need only a short trace
through the existing behavior. Avoid inventing a specification or conducting an
architecture exercise when the contract is already clear.
