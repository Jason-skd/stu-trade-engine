---
name: resolve-review
description: Evaluate and, when authorized, resolve code-review findings or pull-request comments against a specific revision. Do not use for an independent general code review without supplied findings, ordinary bug reports, or feature design.
---

# Resolve Review

Treat every review comment as a hypothesis. Determine whether it still applies
to the referenced revision before changing code. Review evaluation alone does
not authorize fixes.

## Anchor the Review

- Identify the revision, branch, paths, and symbols the comments concern.
- Confirm the working checkout corresponds to that evidence or account for the
  difference explicitly.
- Preserve unrelated local work and avoid broad cleanup.

## Classify Each Finding

For every comment, inspect the implementation, callers, contracts, and tests,
then classify it as:

- **Current defect:** the issue is reproducible or supported by current code.
- **Already resolved:** later changes removed the issue.
- **False positive:** the claimed behavior does not occur or is required by a
  stronger contract.
- **Needs decision:** evidence is insufficient or valid behaviors conflict.

Record the concrete evidence. Do not fix code merely to satisfy the wording of
a comment.

## Resolve When Authorized

- Group related current defects and address their common cause.
- Add behavioral or regression coverage when behavior changes.
- Validate each group with the closest relevant checks before expanding the
  matrix.
- Keep changes focused. Do not rewrite unaffected code or preserve obsolete
  paths to make the review easier to close.

## Report the Result

Lead with unresolved findings ordered by severity and grounded in file and line
references. Then report fixed findings, rejected or obsolete comments with
reasons, and verification evidence. If no issue remains, say so directly and
identify any test gap or residual risk.
