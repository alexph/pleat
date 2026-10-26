# Challenge the design

Trace realistic usage through the candidate before selecting it. Use the relevant
checks below. A warning sign calls for examination, not automatic rejection.
The linked principle skills own the engineering guidance.

## Caller and interface

Check whether the caller can complete a meaningful operation without knowing
private sequencing or coordinating several internal methods. Does the interface
hide useful complexity, or merely rename another call?

If representation details cross the interface, establish whether they are part of
the intended contract. A protocol client may deliberately expose protocol types.
Domain logic usually needs a representation independent of the transport.

## State and ownership

Trace who can change each invariant and where enforcement occurs. Look for
multiple writers with incompatible rules, mirrored derived state, and lists that
must be edited together. Separate policies that only happen to look alike today.

For mutable invariants, inspect the read-to-write interval. Can another actor
invalidate the earlier check? Identify the required transaction, constraint, or
coordination boundary. Splitting data between actors does not solve invariants
that still require shared authority.

## Valid states and failure

Try a success, a relevant invalid input, and a failure after partial work when
effects are involved. For retries, ask what happens if the same operation runs
again. For permissions, establish when access must still be valid.

Check whether types express the intended states or merely conceal missing
decisions behind optional fields, casts, and flags. Distinguish invalid internal
state from legitimate conditions that need explicit handling.

## Change and integration

Trace every affected contract to its consumers. Include stored data and work
already in progress where relevant. Can old and new code coexist during the
intended delivery? What happens if only part of the change takes effect?

Look for duplicated paths that will keep gaining callers. Prefer one clear
supported path, while retaining compatibility that current consumers or delivery
constraints actually require. Preserve unrelated and dormant features.

## Reader effort and evidence

Read the caller and follow the important decision to its owner. Identify jumps
that hide behavior, forwarding layers without a purpose, or dependencies implicit
in a large context object. Judge the reasoning required rather than file count.

For the strongest assumption, name a result that could disprove it. Establish
which claims the sketch supports and which need execution. If competing designs
share that assumption, test it rather than treating agreement as confirmation.
