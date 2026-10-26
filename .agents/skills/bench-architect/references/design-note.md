# Record the design decision

Capture the parts below that matter to the task. Keep a small decision brief.
Expand only where another engineer needs the detail to implement, challenge, or
maintain it. Link to source files and existing contracts instead of copying them.

## Outcome and constraints

State the caller's intended result and the constraints that make the design
non-obvious. Identify observed facts, consequential assumptions, and source paths.

## Caller usage

Show realistic inputs, calls, and outputs. Include a failure result where it
affects the contract. Derive the internal sketch from this usage.

## Proposed shape

Show the relevant types, signatures, and module responsibilities. Trace state
transitions, ownership, validation, and effects where they matter. Identify the
invariants and how the design enforces them. Keep sketches outside live code.

## Decision and alternatives

Name the chosen shape and why it fits the constraints. Compare meaningful
alternatives by their consequences for callers and maintainers. If candidates
were delegated, identify the selected design and any adaptations. Reconcile
their assumptions before combining ideas.

Record when the existing design was sufficient or the constraints left only one
viable shape. Do not invent alternatives to fill this section. Name tradeoffs a
future reader could mistake for oversights.

## Verification and delivery

Name the observable outcomes that will verify the design and expose its main
failure risks. Include affected consumers, data transitions, and partial delivery
where relevant. Separate planned checks from completed checks and their results.

## Remaining decisions and next action

State unresolved choices and the evidence or product judgment needed to settle
them. Give the next implementation or investigation step. For design-only work,
label implementation as future work. Do not imply authorization to proceed.
