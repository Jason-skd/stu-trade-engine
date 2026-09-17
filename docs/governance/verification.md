# Verification Governance

Read this file for behavior changes, tests, CI, asynchronous work, managed
resources, or claims that an implementation is complete.

## Test Ownership

- Added or changed behavior requires behavioral tests. A bug fix adds a
  regression scenario that reproduces the original defect.
- Test closest to the rule owner. Cross-layer tests cover protocols, wiring,
  and critical collaboration rather than repeating every lower-layer case.
- Use the real target database and actual migrations for constraints,
  transactions, locks, and query semantics. Mocks do not prove storage behavior.
- Use substitutes only where they reliably express the behavior under test.

## Test Size and Resources

- Every scenario protects an explainable behavior or failure mode. Extend an
  existing parameterized test when cases share a contract.
- Reuse small fixtures and helpers. Do not build test DSLs or mirror production
  logic inside tests.
- Register cleanup immediately after acquiring a resource. Helpers own ordinary
  resource lifecycles rather than shifting them to every caller.
- Wait for asynchronous results with conditions and explicit timeouts, not
  fixed sleeps. Cleanup uses an independent bounded context where applicable.
- Remove tests that lose independent value when behavior is removed or merged.

## Completion Evidence

- Fast suites may explicitly exclude external systems. Full verification must
  not silently skip unavailable required dependencies.
- Local full verification and CI use the same criteria. Check formatting,
  static analysis, tests, generated synchronization, and contract validity as
  applicable to the change.
- Verification detects drift rather than hiding it through automatic edits.
- Report commands actually run, their current-checkout results, omitted checks,
  and residual risk. Coverage percentages do not replace meaningful scenarios.
