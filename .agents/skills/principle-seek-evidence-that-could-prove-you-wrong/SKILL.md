---
name: principle-seek-evidence-that-could-prove-you-wrong
description: Apply when choosing verification or assessing the strength of evidence. Challenge consequential assumptions with independent expected results and plausible failure cases. Judge whether a check could expose a meaningful mistake.
disable-model-invocation: true
---

# Seek evidence that could prove you wrong

Choose checks that could expose a meaningful mistake in your explanation or
implementation. Establish the expected result independently of the code under
test.

- Name the consequential assumption and an observable result that would refute
  it. Check a plausible failure case as well as the intended success path where
  those outcomes differ materially.
- Use the real caller or artifact at the nearest practical boundary. Type checks,
  builds, and internal mocks prove narrower claims than successful use of the
  connected behavior.
- Keep expected values independent of production helpers. For a regression test,
  confirm it catches the original fault when practical.
- Select checks by the consequences and likelihood of failure. Do not generate
  exhaustive matrices or duplicate tests without a distinct behavior to protect.
- Treat reviewer or model agreement as another opinion. Resolve important claims
  against code, contracts, or observed results rather than counting agreeing
  agents.

After relevant checks pass, repeat or broaden them only for a new change, a
failure, or an unresolved concern. Use [Prove your work](../principle-prove-your-work/SKILL.md)
to decide what completion claim the results justify and report their limits.
