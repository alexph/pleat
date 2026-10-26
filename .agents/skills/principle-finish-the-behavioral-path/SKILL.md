---
name: principle-finish-the-behavioral-path
description: Apply when a change crosses producers, consumers, persistence, or delivery paths. Complete and verify the connected work needed for the caller to experience the requested result.
disable-model-invocation: true
---

# Finish the behavioral path

Follow the requested behavior from its entry point to the result its caller can
observe. Complete the connected changes needed for that result.

- Trace the actual path before editing. Inspect producers and consumers of any
  contract you change, including indirect callers such as jobs or event handlers
  when they use that contract.
- Include persistence, serialization, permissions, generated artifacts, and
  failure handling where the changed behavior depends on them. Follow the path
  that exists rather than treating this list as mandatory work for every task.
- Coordinate transitions between old and new representations. Account for
  existing data or in-flight work when relevant. Keep compatibility only for a
  concrete delivery or consumer requirement.
- Verify the caller's result through the connected path. A passing check on one
  component does not establish that its consumers can use the change.
- If an external dependency prevents completion, finish the independent work and
  identify the exact remaining gap. Distinguish implemented pieces from a
  demonstrated outcome.

Follow dependencies only as far as the requested behavior requires. This
principle does not expand a local task into a platform rewrite or authorize
deployment and other external actions.
