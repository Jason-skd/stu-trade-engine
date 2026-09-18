# Delivery Governance

This file governs evidence and identity across pull-request, main-branch,
artifact, deployment, and release stages.

## Delivery Acceptance

- **G066** Before merge, the pull-request gate proves that required updates are
  complete and the change complies with applicable governance and repository
  checks.
- **G067** The main gate runs for every direct push and merge commit on `main`
  and proves that exact revision and its artifacts are deployable.
- **G068** Development completion, verification success, deployable candidacy,
  and production release remain distinct evidenced states.
- **G069** Later delivery stages consume artifacts certified by an earlier
  stage through immutable revision and digest identities; rebuilding produces
  a new candidate.
