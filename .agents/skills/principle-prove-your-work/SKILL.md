---
name: principle-prove-your-work
description: Apply before marking a work unit verified, declaring an outcome complete, or handing work over. Demonstrate the claimed result through the relevant caller or artifact, check the final state, and disclose verification limits.
disable-model-invocation: true
---

# Prove your work

Demonstrate the outcome you claim through the relevant caller or resulting
artifact. Match the evidence to the claim and verify the state being handed over.
Distinguish demonstrated behavior from assumptions, plans, and blocked checks.

- Establish the expected outcome from the requirement or contract before checking
  the implementation. Use [Seek evidence that could prove you wrong](../principle-seek-evidence-that-could-prove-you-wrong/SKILL.md)
  to choose checks that could expose a meaningful mistake.
- Exercise the nearest practical caller boundary that supports the claim. Inspect
  persistence, generated output, or external effects when they are part of that
  contract. Check the actual artifact rather than relying on an intermediate
  representation.
- Keep evidence within its scope. A build establishes a narrower claim than a
  working interaction. One component's tests do not establish that its consumers
  can use the connected feature.
- After a correction or final edit, rerun the affected checks. For connected
  units, verify their combined result before claiming the whole outcome. Repeat
  broader checks only when the change or an unresolved concern justifies them.
- Report the relevant check or action, observed result, and material limits.
  Identify the checked revision or artifact when its state could be ambiguous.
  Link useful evidence without dumping every log or requiring a permanent record
  for a minor change.

When access, dependencies, or an existing failure block verification, complete the
independent work and state which claim remains unsupported. Distinguish a failure
introduced by the change from an observed baseline problem where evidence allows.
Do not hide a relevant failure behind unrelated passing checks.

Keep proof proportional to the outcome. A label change needs evidence of that
label, not a demonstration of the whole application. Structural validation of a
skill establishes its format and references, not its effect on agent judgment.
