---
name: principle-minimise-reader-load
description: Apply when shaping or reviewing code that is hard to trace. Make control flow, dependencies, and state understandable with less mental reconstruction while preserving useful domain boundaries.
disable-model-invocation: true
---

# Minimise reader load

Make behavior understandable without reconstructing it across unnecessary layers,
hidden state, or distant files. Optimize for the next reader's reasoning effort.

- Keep the main flow visible. Flatten nested branches or use early returns when
  they make the alternatives clearer.
- Keep related logic together and mutable state close to its use. Pass the data
  an operation needs so its dependencies and effects are visible.
- Remove forwarding layers that only rename or relay work. Keep helpers that
  name a meaningful operation, isolate complexity, or enforce a domain boundary.
- Use names and data shapes that reveal intent. Prefer explicit code over terse
  expressions that make the reader infer evaluation order or hidden behavior.
- Keep a comment for a non-obvious reason or constraint the code cannot express.
  Avoid narrating the implementation or preserving misleading explanations.

Fewer lines do not necessarily mean less reader effort. A longer explicit flow
may be clearer than a compressed expression. Keep useful abstractions even when
they add a call, and avoid coupling unrelated behavior merely to put it in one
place. Limit changes to the code the task requires.
