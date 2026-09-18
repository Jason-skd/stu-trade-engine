# Verification Governance

This file governs behavioral coverage, real-system semantics, asynchronous
work, resource management, and completion evidence.

## Verification Actions

- **A032** Run the nearest relevant checks first, repair failures, then expand
  to the risk-appropriate validation matrix.
- **A033** Use the real system when its semantics are the subject of the test.
  Use substitutes only when they reliably express the behavior being verified.
- **A034** Register cleanup as soon as a resource is acquired and wait for
  asynchronous outcomes using conditions and explicit timeouts rather than
  fixed sleeps.
- **A035** Before finishing, inspect the complete diff and applicable
  formatting, static analysis, tests, generation, contracts, and documentation.

## Test Acceptance

- **G051** Added or changed behavior has behavioral coverage at the closest
  responsible layer.
- **G052** A repaired defect has a regression scenario that reproduces the
  original failure mode.
- **G053** Cross-layer tests cover protocols, wiring, and critical
  collaboration without duplicating every lower-layer case.
- **G054** Tests of constraints, transactions, locks, query semantics, or
  another real-system contract use that real system and its actual setup.
- **G055** Every scenario protects an explainable behavior or failure mode;
  cases sharing one contract reuse an existing parameterized test where useful.
- **G056** Fixtures and helpers remain small, own ordinary resource lifecycles,
  and do not reproduce production decision logic in a test DSL.
- **G057** Acquired resources have bounded cleanup, and asynchronous tests use
  condition-based waits with explicit timeouts.
- **G058** Tests that no longer protect independent behavior are removed when
  their behavior is removed or merged.

## Completion Acceptance

- **G059** Required external dependencies are not silently skipped by full
  verification; unavailable checks are reported explicitly.
- **G060** Verification covers formatting, static analysis, tests, generated
  synchronization, contract validity, and documentation as applicable to the
  change and repository capabilities.
- **G061** Verification detects generated or formatted drift rather than hiding
  it through unreported automatic edits.
- **G062** Reported evidence names the commands actually run and their results
  on the current checkout, together with omitted checks and residual risk.
- **G063** The final diff is focused and contains the complete requested
  behavior without unrelated changes.
