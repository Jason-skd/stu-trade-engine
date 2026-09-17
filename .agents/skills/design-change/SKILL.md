---
name: design-change
description: Research and design a feature, architecture change, new dependency, shared abstraction, or other change with material design tradeoffs. Do not use for routine implementation with an already established design, bug diagnosis, or code-review findings.
---

# Design Change

Produce an evidence-backed design for the current need without expanding the
assignment. This workflow does not authorize implementation unless the user has
clearly requested it.

## Define Acceptance

Start from observable end-to-end outcomes. Write acceptance so a reviewer who
does not know the implementation can confirm the result. Separate current
requirements from adjacent possibilities and explicit non-goals.

## Establish Evidence

Research enough to answer:

- What behavior exists now, and where is it owned?
- Which callers, data, protocols, tests, and documentation are affected?
- Which existing capabilities or foundation libraries can be reused?
- Which upstream contracts, official references, or established external
  patterns constrain the design?
- Which materially different options remain, and what are their tradeoffs?

Use current code and locked dependency versions as evidence. Historical plans
may guide searches but do not establish current facts.

## Form the Task Contract

Internally state:

```text
Goal: observable behavior to add or change
Non-goals: adjacent work excluded from this task
Ownership: component or boundary responsible for the rule
Impact: code, data, API, UI, tests, generation, and documentation
Acceptance: evidence required to consider the work complete
```

Do not ask the user to approve facts that can be established safely from the
repository. Ask when valid product behaviors conflict, authority is missing, or
an irreversible choice cannot be inferred.

## Design the Change

- Prefer the final coherent shape for the current scope, not a temporary design
  that knowingly leaves duplicate paths or empty compatibility layers.
- Keep modeling minimal. Do not reserve fields, interfaces, directories, or
  extension systems for hypothetical callers.
- Assign validation, conversion, orchestration, persistence, and protocol work
  to their actual owners.
- Examine failure behavior, transaction or side-effect ordering, concurrency,
  compatibility, and migration only where the change makes them relevant.
- Extract shared operations only when they have the same contract.

## Present the Result

Report a feature or material design in this order:

1. Intuitive acceptance.
2. Referenceable patterns, available dependencies or vendors, and reusable
   existing capabilities, including useful responsibility separation.
3. Detailed solution, ownership boundaries, tradeoffs, and verification plan.

State unresolved decisions and missing evidence plainly. Do not present a
speculative choice as an established project decision.
