# Interface Governance

Read this file when changing errors, HTTP behavior, transport DTOs, or external
protocol integration. Upstream integrations preserve their actual contracts
even when local defaults differ.

## Errors

- Reuse established error identities, constructors, chains, and response
  mechanisms. Do not introduce a parallel general error system.
- Keep application-owned error identities with their actual owner; distinct
  conditions must not accidentally share codes or identifiers.
- Translate dependency failures into domain meaning at adapter boundaries.
  Business logic must not inspect database or transport error strings.
- Preserve meaningful cause chains and add operation context only when useful.
  Lower layers return errors and the final handling boundary logs once.
- Keep internal causes out of client messages. Explicitly map binding,
  validation, authentication, authorization, absence, and conflict failures.
- Do not turn corrupt input or business failure into success. Reserve panic or
  termination for unrecoverable programming or static assembly errors.

## HTTP and DTOs

- Owned resource paths use plural kebab-case nouns. Use POST for genuine domain
  commands rather than disguising them as reads.
- Preserve GET safety and PUT/DELETE idempotence. PATCH distinguishes missing
  fields from explicit null, false, zero, and empty strings.
- Use transport statuses according to operation semantics. Do not encode a
  business failure as a successful response.
- Keep transport DTOs and mappings outside domain entities. Determine request
  requiredness, response presence, and nullability separately.
- Runtime responses and generated API documentation share structural sources.
- Use RFC 3339 timestamps with timezones unless an upstream contract differs.
  Empty collections are arrays rather than accidental null values.
