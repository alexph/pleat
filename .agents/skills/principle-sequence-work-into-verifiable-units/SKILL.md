---
name: principle-sequence-work-into-verifiable-units
description: Apply when planning or executing multi-step work, migrations, or a sequence of commits or PRs. Order coherent units by dependency, verify each before building on it, and preserve understandable delivery checkpoints.
disable-model-invocation: true
---

# Sequence work into verifiable units

Divide work into coherent units with observable outcomes and meaningful checks.
Order them by dependency. Verify each unit before building further on it so a
fault is found while its cause is still local.

- Before starting a unit, identify its outcome, prerequisites, and the check that
  will establish it. Choose a behavior, contract, or migration step that can be
  verified. A file boundary or arbitrary line count is not enough.
- Prioritize shared prerequisites when they demonstrably enable dependent units
  or remove repeated obstacles. Establish needed types, checks, or tooling before
  the work that relies on them. Tie that setup to the actual plan rather than
  building infrastructure merely because later work might benefit.
- Build on verified prerequisites. If a check fails, repair the unit or reconsider
  its premise before adding dependent work. Continue an independent unit only
  when its own prerequisites hold.
- Keep tightly coupled changes together when separating them would leave an
  invalid intermediate state. Temporary failing states during development,
  including a regression test before its fix, are not delivery checkpoints.
- After integrating units, check their combined behavior. When an earlier unit
  changes, recheck affected dependents rather than relying on their old results.

Use [Prove your work](../principle-prove-your-work/SKILL.md) to establish what each
checkpoint demonstrates. Do not treat a plan checkbox or completed edit as proof.

When delivering through commits, a draft PR, or stacked PRs, preserve coherent
reviewable checkpoints and their dependency order. Keep the relevant checks
passing at each delivered checkpoint. A draft label does not establish a working
base. Record verification limits and distinguish units checked against their
parent from units ready to land independently.

Do not require a separate commit or PR for every local unit. Choose delivery
boundaries that help review and recovery. Keep the unit small enough to diagnose
and repair without fragmenting a meaningful change into unfinished pieces.
