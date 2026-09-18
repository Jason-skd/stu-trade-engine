---
name: resolve-review
description: Evaluate and, when authorized, resolve supplied code-review findings against a specific revision. Do not use for independent general review, ordinary bug reports, or feature design.
---

# Resolve Review

Treat each comment as a hypothesis. Apply the routed governance in `AGENTS.md`;
evaluation alone does not authorize fixes.

Anchor the comment to its revision, branch, paths, and symbols, then inspect the
implementation, callers, contracts, and tests. Classify each finding as a
current defect, already resolved, false positive, or needing a product or
ownership decision. Record concrete evidence rather than changing code to
satisfy wording.

When repair is authorized, group related current defects by root cause, add
behavioral or regression coverage where behavior changes, and run the closest
checks before expanding verification. Keep unrelated cleanup out of the diff.

Lead the report with unresolved findings ordered by severity and grounded in
file references. Then report fixes, rejected or obsolete comments with reasons,
and verification evidence.
