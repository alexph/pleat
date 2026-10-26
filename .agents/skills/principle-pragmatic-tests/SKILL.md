---
name: principle-pragmatic-tests
description: Apply when writing, changing, reviewing, or deciding whether to keep tests. Call code through the boundaries its users use, assert observable behavior, and keep only tests that add concrete regression value.
disable-model-invocation: true
---

# Pragmatic Tests

Ask: “Does this test add any real value?” Prefer the smallest test that would catch a meaningful failure through the way the code is actually used. Judge value by the behavior protected and the maintenance cost, rather than test count or coverage percentage.

## Write through the caller's boundary

- Call the code the way its users do. A user may be a person using a UI, an API client, or code consuming a public function. Use the relevant supported entry point: a UI action, HTTP request, command, or public API.
- Prefer a small test near the top of the relevant call chain that exercises the real internal path. This does not require a full browser or system test for every behavior; choose the nearest caller boundary that covers the contract and runs deterministically.
- Assert the result that caller can observe against a literal expected value: returned data, a response, visible content, persisted state read through its supported interface, or an externally visible effect. Keep expectations independent of the implementation. Do not calculate the expected answer with the production helper being tested or duplicate its algorithm in the test.
- Trust internal code to collaborate. Exercise it rather than mocking it into a prescribed call sequence. Avoid assertions about private helpers, internal method calls, incidental query counts, or object layout unless that detail is an explicit contract or a demonstrated requirement.
- Replace external boundaries when needed for deterministic tests, such as network services, clocks, and randomness. Keep internal behavior real. When an external effect is the contract, assert the meaningful outbound payload or effect rather than incidental calls used to construct it.
- Use only the setup and assertions needed to expose the behavior. Prefer existing fixtures and one representative case; add cases when they protect distinct outcomes or plausible regressions. Avoid test frameworks, fixture hierarchies, and parameter matrices that exceed the problem.

## Decide whether a test earns its place

Before adding or retaining a test, name the concrete regression it would catch. Check existing tests at higher caller boundaries for the same scenario and observable assertion. Add or strengthen a test only where that protection is missing or the existing test cannot expose the failure reliably.

For a lower-level test, ask: “Could this test fail for a meaningful behavior regression while the retained tests further up the chain still pass?” If yes, identify that gap. If no, it may be redundant. Passing today, executing the same code, or sharing a coverage line does not establish redundancy; compare inputs, branches, failure conditions, and assertions.

Keep a focused lower-level test when it protects a distinct public contract, an important edge case, or a requirement that is impractical to exercise higher up. Do not preserve tests solely because they exist, and do not delete them solely because they are unit tests. Useful tests should survive an internal refactor that preserves their contract.

When pruning within the requested scope:

- Inspect each candidate and the retained tests before removing it. Preserve unique regression cases, moving them to a caller boundary when practical.
- Prefer improving one existing behavior test over adding several overlapping tests.
- Run the relevant retained tests after changes. If a replacement regression test can reasonably be checked against the previous faulty behavior, confirm that it catches that fault.
- Keep deletion within the task's scope. Creating or invoking this skill does not authorize a repository-wide test purge.

## Report pruned tests

Include this table in the final response whenever tests are removed or replaced. Give each removed test its own row; identify parameter cases if only some are pruned. Use actual paths and test names, and name the retained test rather than claiming vaguely that coverage exists.

| Pruned test       | Behavior protected                     | Why pruned                                     | Retained or replacement test                                     | Verification                                          |
| ----------------- | -------------------------------------- | ---------------------------------------------- | ---------------------------------------------------------------- | ----------------------------------------------------- |
| `path::test_name` | Concrete scenario and expected outcome | Specific redundancy or implementation coupling | `path::test_name`, or a specific reason no replacement is needed | Command and result, or explicitly not run with reason |

For replacement tests, explain how they preserve the removed test's useful protection. Report verification limits plainly. If no tests were pruned, omit the table.
