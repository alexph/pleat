---
name: principle-keep-uncertain-decisions-reversible
description: Apply when an uncertain structural choice could be expensive to undo, such as an integration boundary, persisted representation, or public contract. Localize the uncertain decision and preserve a practical replacement path without speculative flexibility.
disable-model-invocation: true
---

# Keep uncertain decisions reversible

Identify decisions that are expensive to undo. When requirements or assumptions
are uncertain, prefer a structure that keeps the decision local and permits
replacement at reasonable cost. Accept commitment when evidence or simplicity
justifies it.

- Name the uncertain assumption and the concrete cost of changing course. Give
  more attention to decisions embedded in persisted data, public contracts, or
  many consumers than to local code that is cheap to replace.
- Keep the uncertain choice at its existing owner or boundary. Avoid spreading
  vendor details, representations, or policies through callers that need only a
  stable domain operation.
- When the next useful unit can proceed without settling the choice, defer it.
  Otherwise choose a sufficient implementation within the task's scope. Resolve
  consequential factual uncertainty with evidence before cementing it into a
  wider contract.
- Compare the cost of preserving a replacement path with the plausible cost of
  change. A direct implementation may be the better commitment when another layer
  adds more complexity than it saves.
- For a consequential commitment, record the assumption and why the tradeoff is
  justified. Distinguish replaceable code from persistent or external effects.
  Swapping an adapter does not undo stored data or actions already taken.

Preserve options through local decisions and clear contracts. Do not add plugin
systems, configuration switches, generic interfaces, or support for hypothetical
providers solely because requirements might change. Reversibility is a decision
criterion, not a demand to keep every possible design available.
