# Engineering Governance

This file governs modeling, reuse, responsibility boundaries, authoritative
sources, dependencies, errors, compatibility, generated artifacts, and
comments.

## Modeling

- **G021** The implementation models complete current requirements without
  fields, operations, abstractions, directories, or public capabilities for
  hypothetical future needs.
- **G022** Every public capability and persisted field has a present caller,
  reader, writer, or explicitly named owner.
- **G023** Keys, constraints, and operations follow real queries and invariants
  rather than a generic CRUD shape.
- **G024** Structural correctness is enforced by the most authoritative layer
  that can express it; other layers do not duplicate its constraints or side
  effects.

## Reuse and Responsibilities

- **G025** Existing capabilities and locked dependencies are reused when they
  already implement the required contract.
- **G026** Shared code represents a real common operation with one contract,
  not merely similar syntax or a universal flow controlled by modes.
- **G027** Code is split by responsibility rather than length or directory
  symmetry, while closely cooperating behavior remains nearby.
- **G028** A new layer validates, converts, orchestrates, or isolates a
  dependency; it does not only forward calls.
- **G029** Abstractions serve current callers and do not form speculative
  frameworks or empty extension points.

## Capability and Dependency Boundaries

- **G030** Consumers define the smallest capability interfaces they need;
  callers own dependency injection, process assembly, and lifecycle wiring.
- **G031** Business rules remain independent of transport entry points, process
  assembly, and concrete infrastructure where the repository architecture
  supports that separation.
- **G032** Protocol entry points own protocol concerns, while adapters contain
  dependency-specific types and perform real translation or isolation.
- **G033** Functions are used when no state or dependency is retained; code is
  not forced into a uniform service pattern.
- **G034** New dependencies provide a present capability that cannot be met
  adequately by the standard library or an established repository dependency.

## Sources, Generation, and Compatibility

- **G035** Each schema, protocol, query, and generated representation has one
  identified authoritative source.
- **G036** Generated artifacts are produced by the established workflow and
  are not edited or synchronized manually.
- **G037** Dependency behavior follows the contract and source of the locked
  version actually in use.
- **G038** Upstream interfaces are not changed for local naming preferences,
  and dependency internals are not copied to bypass a missing capability.
- **G039** Internal changes update declarations, callers, tests, examples,
  documentation, and generated bindings together.
- **G040** A compatibility layer exists only for a real external boundary with
  a named consumer and a removal condition.

## Errors and External Boundaries

- **G041** The implementation reuses the repository's established error
  identities, constructors, chains, and response mechanism rather than adding
  a parallel general error system.
- **G042** Distinct failure conditions retain distinct identities, and
  dependency failures are translated into domain meaning at their adapter
  boundaries.
- **G043** Meaningful causes are preserved and operation context is added only
  when useful; errors are logged once at the final handling boundary.
- **G044** User-facing failures exclude internal causes while mapping relevant
  validation, authentication, authorization, absence, conflict, and dependency
  conditions explicitly.
- **G045** Corrupt input and business or side-effect failures are not reported
  as success. Panic or termination is reserved for unrecoverable programming
  or static assembly errors.
- **G046** Transport methods, status, idempotence, requiredness, nullability,
  and serialization match the owned protocol semantics.
- **G047** Transport DTOs and mappings remain outside domain entities, and
  runtime output and generated protocol documentation share structural sources.

## Comments

- **G048** Public contracts follow the language's documentation conventions;
  non-public contracts are documented when names and code are insufficient.
- **G049** Explanatory comments record reasons, constraints, or non-obvious
  tradeoffs rather than narrating clear code.
- **G064** Code is organized into logical blocks, each introduced by a
  `// === verb-objective phrase ===` header that states what the block does.
- **G050** Comments describe current behavior only, not issue history, pull
  requests, governance clauses, migrations, or speculative future work.
