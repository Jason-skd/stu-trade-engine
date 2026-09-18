---
name: cross-repository-escalation
description: Hand off a required change whose authoritative source or implementation belongs to another repository. Use before any cross-repository write.
---

# Cross-Repository Escalation

Apply `AGENTS.md` and the ownership topology in
`docs/governance/repository.md`. A nested checkout, submodule, shared workspace,
or common product does not extend write authority.

Identify the owning repository and collect enough current, reproducible
read-only evidence for a handoff. Do not create branches, edit files, commit,
open issues or pull requests, comment, message maintainers, or push in another
repository without explicit authority naming that target and action.

For repositories in the same organization and system, inspect read-only
evidence across the topology. If target-repository writing is not authorized,
prepare an Emergency Handoff containing the affected contract, evidence,
blocking impact, compatibility constraints, owner-side direction, observable
acceptance, and consumer revision requirement. Confirm the target repository
and issue-creation action before filing it. If the user has already given
equally specific target write authority, implement only within that scope.

For an external upstream, instead prepare a user-facing formal statement of
evidence, impact, compatibility, and viable strategy options. Wait for the user
to decide whether and how to contact upstream; never proactively create an
issue, comment, or message.

Stop the blocked portion until the owner produces a merged revision. Only after
verifying that revision may work in a currently authorized consumer update its
pin, gitlink, adapter, or integration metadata and run the relevant integration
checks. Continue unrelated in-scope work only when it does not assume the
pending owner change.
