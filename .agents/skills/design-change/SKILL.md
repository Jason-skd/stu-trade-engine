---
name: design-change
description: Research and design a feature, architecture change, dependency, or abstraction with material tradeoffs. Do not use for routine implementation with an established design, bug diagnosis, or review findings.
---

# Design Change

This workflow does not authorize implementation. Apply the routed governance
in `AGENTS.md`, especially the engineering rules relevant to the design.

Start from observable end-to-end acceptance and separate the current need from
non-goals. Establish current behavior, ownership, affected callers and
contracts, reusable capabilities, locked dependency constraints, and materially
different options from current evidence.

Form an internal task contract with the goal, non-goals, owner, affected code
and data boundaries, and evidence required for completion. Ask only when valid
product behaviors conflict, authority is missing, or an irreversible choice
cannot be inferred.

Present the result in this order:

1. Intuitive acceptance.
2. Relevant existing capabilities, dependencies, and referenceable patterns.
3. The proposed design, ownership boundaries, tradeoffs, and verification.

State unresolved decisions and missing evidence plainly. Do not present a
speculative choice as an established project decision.
