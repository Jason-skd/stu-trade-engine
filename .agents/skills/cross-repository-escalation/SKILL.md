---
name: cross-repository-escalation
description: Hand off a required change when its authoritative source or implementation belongs to another repository. Use before any cross-repository write.
---

# Cross-Repository Escalation

Confirm which repository owns the required change from
`docs/governance/ownership.md`. Inspect that repository only as much as needed
to collect current, reproducible evidence. Do not create a worktree or branch,
edit files, commit, open a pull request, or push there. A submodule, nested
checkout, shared workspace, or common product does not extend authority.

File an Emergency Issue in the owning repository. This issue is the sole
permitted cross-repository mutation. Use the repository issue structure and
include:

- the current evidence and affected contract;
- the impact blocking the originating repository;
- the recommended owner-side direction without prescribing unsupported detail;
- observable acceptance conditions and any revision needed for integration.

Report the issue link to the user and ask them to schedule the owning
repository's agent. Stop the blocked portion of the task until that repository
produces a merged revision. Continue unrelated in-scope work only when it does
not assume the pending change.

After handoff completion, verify the merged owner revision. Modify only the
current repository to consume it, update any pinned dependency or gitlink, and
run the relevant integration checks. Explicit user authorization naming every
target repository is required to synchronize governance across them; such an
exception does not create standing cross-repository authority.
