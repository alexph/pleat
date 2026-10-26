---
name: figure-it-out
description: Design and execute an auditable workflow for a large or cross-cutting change, work reviewed after a handoff, or a task with no suitable Bench playbook. Use for /figure-it-out or 'figure it out'.
disable-model-invocation: true
---

# Figure it out

Design the workflow before implementation. Scale verification to the consequences, test assumptions as you work, and leave enough evidence for a reviewer to assess the result.

## Start

Read the Principles section of [bench-mode](../bench-mode/SKILL.md) and the relevant linked principles. Track phases A through E in the existing task plan or a local run note.

## Phase A: Frame

Inspect the current implementation and affected callers. Record the following before implementation.

- The observable outcome and what evidence would disprove completion.
- The affected areas, rough units of work, consequential assumptions, and blockers.
- The rigor level and the checks it requires. Increase verification and recovery planning for irreversible changes or broad effects. Keep reversible, local work lighter.

Present the framing before a long run, then continue authorized work. Ask about consequential choices that evidence and existing authorization cannot settle. For planning-only requests, stop after phase B.

## Phase B: Design the workflow

Apply [Sequence work into verifiable units](../principle-sequence-work-into-verifiable-units/SKILL.md). Resolve consequential unknowns early, subject to dependencies. Keep tightly coupled changes together when splitting them would leave an invalid intermediate state.

Write an ordered workflow with each unit's outcome, prerequisites, and verification.

- Capture a baseline where comparison matters. Reuse existing checks. Build a harness only when needed.
- Use [bench-architect](../bench-architect/SKILL.md) for changes to contracts, ownership, or data flow across boundaries. Skip it for routine work that preserves those contracts.
- If delegation is authorized and useful, assign independent units with explicit ownership and acceptance checks. Isolate overlapping edits. The coordinator verifies the integrated artifact.

Execute under phase C and keep the phase D trail current.

## Phase C: Run the loop

For each unit, state the hypothesis, make the smallest useful change, and measure against its expected outcome. Keep changes supported by evidence. Revise or remove unsuccessful attempts without discarding unrelated work.

- Inspect the resulting artifact or caller behavior. A worker's completion report is not proof.
- If a check passes without exercising the claim, repair the check. Do not weaken acceptance criteria to fit the implementation.
- After a failure, continue independent units only when their prerequisites hold.
- A verdict is VERIFIED, NOT VERIFIED, or INCONCLUSIVE. Inconclusive is not a pass. Don't hide a negative.

## Phase D: Keep the audit trail

Record consequential decisions and reasons, checks and observed results, failed hypotheses that changed the approach, and unresolved limits. Link evidence rather than copying transcripts.

Use the plan or run note from Start. Include it in the handoff or an authorized PR when useful.

## Phase E: Verify and hand back

Apply [Prove your work](../principle-prove-your-work/SKILL.md). Check the integrated result against phase A through the relevant caller or final artifact. State what remains unverified and why.

For recurring failures, apply [Encode lessons in structure](../principle-encode-lessons-in-structure/SKILL.md) when an enforceable check justifies its maintenance cost.

**Reply:** the designed workflow, rigor level and reason, decision-trail path, verified outcomes, and open limits. For planning-only work, identify the workflow as proposed.
