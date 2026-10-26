---
name: principle-fix-root-causes
description: Apply when debugging defects or when successive fixes fail to resolve a symptom. Trace the first incorrect assumption, decision, or state transition and correct it at its owner.
disable-model-invocation: true
---

# Fix root causes

Explain how the wrong result arose before choosing a correction. Trace the
symptom back to the first place required behavior or an invariant was violated.

- Reproduce the reported failure through the relevant caller when practical.
  Preserve the input and conditions that expose it. If reproduction is blocked,
  separate observed facts from the explanation you still need to test.
- Follow the data and decisions that produced the failure. Compare the actual
  state with the contract at each relevant step.
- Fix the incorrect producer or decision at its owner. Avoid concealing invalid
  state with defaults, broad exception handling, or checks in every consumer.
- Treat legitimate failure conditions explicitly. A guard, retry, or fallback is
  appropriate when the condition belongs to the supported behavior and its
  handling preserves the caller's contract.
- When successive fixes fail the same check, write down their shared premise and
  test it before making another change that assumes it.

Verify that the correction addresses the original failure and preserves nearby
required behavior. Add regression protection where it has concrete value. Keep
the correction within the necessary scope. Finding an architectural weakness is
not permission to redesign every related component.
