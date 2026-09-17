# Interface Governance

This file governs errors, HTTP behavior, transport DTOs, and external protocol
integration. Clause numbering continues from `implementation.md`.

## Errors

- **G098** Reuse established error identities, constructors, chains, and
  response mechanisms. Do not introduce a parallel general error system.
- **G099** Keep application-owned error identities with their actual owner;
  distinct conditions must not accidentally share codes or identifiers.
- **G100** Translate dependency failures into domain meaning at adapter
  boundaries. Business logic must not inspect database or transport strings.
- **G101** Preserve meaningful cause chains and add operation context only when
  useful. Lower layers return errors and the final handling boundary logs once.
- **G102** Keep internal causes out of client messages. Explicitly map binding,
  validation, authentication, authorization, absence, and conflict failures.
- **G103** Do not turn corrupt input or business failure into success. Reserve
  panic or termination for unrecoverable programming or static assembly errors.

## HTTP and DTOs

- **G104** Owned resource paths use plural kebab-case nouns. Use POST for
  genuine domain commands rather than disguising them as reads.
- **G105** Preserve GET safety and PUT/DELETE idempotence. PATCH distinguishes
  missing fields from explicit null, false, zero, and empty strings.
- **G106** Use transport statuses according to operation semantics. Do not
  encode a business failure as a successful response.
- **G107** Keep transport DTOs and mappings outside domain entities. Determine
  request requiredness, response presence, and nullability separately.
- **G108** Runtime responses and generated API documentation share structural
  sources.
- **G109** Use RFC 3339 timestamps with timezones unless an upstream contract
  differs. Empty collections are arrays rather than accidental null values.
