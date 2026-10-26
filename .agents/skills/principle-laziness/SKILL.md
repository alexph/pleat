---
name: principle-laziness
description: Apply when refactoring, evaluating diff size, or when considering adding abstractions, layers or signal threading. Bias toward deletion and the smallest change that solves the problem.
disable-model-invocation: true
---

# Laziness

Choose the smallest complete implementation that satisfies the request and is easy to understand, verify, and change. Minimize the work the code creates for its next reader. Preserve required behavior, correctness, and applicable project constraints.

## Principles

- **Smallest sufficient change.** Start with the existing implementation. Change only what the requested behavior needs; avoid adjacent cleanup, renaming, formatting churn, and speculative flexibility. A small diff must still solve the whole problem.
- **Deletion first.** Look for unnecessary branches, state, wrappers, and indirection that can be removed within scope. Respect project preservation rules; dormant features are not automatically candidates for deletion.
- **Flat call hierarchy.** Keep the main flow visible in one place with shallow, purposeful helpers. Avoid chains of forwarding functions and layers that merely rename or relay work. Keep useful domain boundaries; flattening should make the code easier to follow.
- **Consolidate decisions.** Resolve a policy, configuration choice, or derived value once at the boundary that owns it, then pass the result to its consumers. Do not repeat the same choice across several functions or modules. Keep the source of truth close to its owner rather than creating a global registry.
- **Directness.** Prefer explicit calls, ordinary data, and visible control flow. Use a helper when its name explains a meaningful operation or it contains complexity worth separating. Avoid helpers that only wrap a simple expression without adding meaning.
- **Earn abstractions.** Reuse existing primitives first. Add an abstraction when current callers share a stable concept, it enforces a real invariant, or it clearly reduces complexity. Avoid factories, generic frameworks, strategy layers, and options for hypothetical future callers. Similar syntax alone does not establish a shared concept.
- **Keep dependencies narrow.** Pass the value a consumer needs instead of an entire context object. Before threading a flag, signal, or callback through several layers, check whether the decision or effect belongs at an existing owner. Preserve cancellation, authorization, and lifecycle semantics where they require propagation.
- **Keep state minimal.** Derive values from the source of truth when practical. Avoid mirrored state, caches, and synchronization machinery unless an actual performance or lifecycle requirement justifies them.
- **Prefer local reasoning.** Keep code that changes together close together. A little duplication can be cheaper than coupling unrelated paths; consolidate shared decisions without forcing unrelated behavior into one abstraction.
- **Stop when done.** Once the request is satisfied and relevant checks pass, stop. Do not expand the patch to improve the architecture in general.

## Apply to the change

Identify the required observable behavior and trace the existing path that owns it. Consider a direct edit or removal before introducing a new layer. Choose the approach with the least total complexity, considering branches, state, dependencies, and reader effort as well as line count.

Before finishing, inspect the diff for unrelated edits, repeated decisions, pass-through helpers, and machinery that no current requirement needs. Remove those where doing so preserves the intended behavior. Run the relevant checks and behavior tests required by the project; keep verification proportional to the change.

If retaining an abstraction or a larger diff is necessary, briefly explain the concrete requirement it serves. Do not compress clear code into clever expressions just to reduce lines.
