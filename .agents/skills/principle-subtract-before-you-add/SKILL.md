---
name: principle-subtract-before-you-add
description: Apply when an addition appears to require another branch, wrapper, or compatibility layer. Check whether unnecessary existing structure creates that need, and simplify it first when that directly helps the task.
disable-model-invocation: true
---

# Subtract before you add

Before adding machinery, check whether an unnecessary branch, duplicated
decision, or obsolete layer is creating the need for it. Simplify that structure
first when doing so makes the requested change clearer.

- Trace why the addition seems necessary. Look for state that can be derived,
  adapters that no longer serve a caller, or policies repeated at several sites.
- Check callers and required behavior before removing anything. Existing code is
  not obsolete merely because the current task does not use it.
- Remove or consolidate within the task's scope, then build on the simpler path.
  Keep each intermediate change verifiable when sequencing matters.
- If removal increases risk without reducing the complexity of this task, make
  the direct addition. Subtraction is a useful option, not a prerequisite ritual.

Preserve project retention rules, supported contracts, and concrete compatibility
needs. Dormant features are not deletion candidates by default. Do not start an
unrelated cleanup campaign, remove unique regression protection, or compress
clear code just to reduce line count.
