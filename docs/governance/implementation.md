# Implementation Governance

This file governs design, implementation, refactoring, dependencies, comments,
schemas, queries, and generated sources. Clause numbering continues from
`ownership.md`.

## Modeling

- **G074** Model current business needs only. Every field, operation, and public
  capability needs a present reader, writer, or caller.
- **G075** CRUD is caller-driven. Do not scaffold operations, fields,
  interfaces, or directories for hypothetical future use.
- **G076** Choose keys from real queries and invariants rather than defaulting
  every operation to a synthetic identifier.
- **G077** Relational schemas enforce structural correctness with applicable
  foreign keys, composite keys, uniqueness, defaults, and triggers. Application
  code owns only rules the schema cannot express.
- **G078** Do not duplicate constraints or side effects already guaranteed by
  an authoritative component.

## Reuse and Decomposition

- **G079** Inspect existing capabilities, foundation libraries, and similar
  logic before adding code or dependencies.
- **G080** Similar syntax prompts a reuse review, but shared code must represent
  a real common operation, not a universal flow controlled by mode parameters.
- **G081** Share validation, conversion, query, or assembly steps only when
  their contracts are genuinely identical.
- **G082** Split code by responsibility, not length or directory symmetry. Keep
  closely cooperating code nearby and avoid mixed-responsibility modules.
- **G083** A new layer must validate, convert, orchestrate, or isolate a
  dependency. Do not add forwarding-only wrappers.
- **G084** Abstract for current callers; prefer a later focused refactor over a
  speculative framework.

## Capability Boundaries

- **G085** Organize business code by domain. Consumers define minimal
  capability interfaces; callers own dependency injection, assembly, and
  lifecycles.
- **G086** Protocol entry points own routing, binding, authentication protocol
  handling, and responses, not business rules.
- **G087** Adapters contain dependency-specific types and perform real
  translation or isolation. Business code does not depend on transport,
  assembly, or concrete infrastructure.
- **G088** Use functions when no state or dependency is retained. Do not force
  every domain into the same service pattern.
- **G089** Follow upstream interfaces and expose only capabilities current
  callers need.

## Sources and Compatibility

- **G090** Keep one authoritative source for each schema, query, protocol, and
  generated representation. Use the official generation workflow.
- **G091** Verify dependency behavior against the contract and source of the
  version in use, not documentation for another release.
- **G092** Internal changes update declarations, callers, tests, examples,
  documentation, and generated bindings together. Do not preserve internal
  aliases or thin forwarding layers solely for old names.
- **G093** A compatibility layer requires a real external boundary, a named
  consumer, and a removal condition.

## Comments

- **G094** Follow the language's documentation conventions and document public
  contracts. Document non-public contracts when names and code are insufficient.
- **G095** Use block headers only for genuinely multi-stage functions. Their
  exact form is the comment marker followed by `=== verb-object phrase ===`,
  with no number or trailing punctuation.
- **G096** Explanatory comments state reasons, constraints, or non-obvious
  tradeoffs; they do not narrate clear code.
- **G097** Comments describe current code only, not issues, pull requests,
  governance clauses, migration history, or future work.
