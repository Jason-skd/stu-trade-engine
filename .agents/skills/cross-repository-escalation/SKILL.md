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
open pull requests, or push in another repository without explicit authority
naming that target.

When issue creation is authorized, file the handoff in the owning repository
using its issue structure. Include the affected contract and current evidence,
blocking impact, owner-side direction without unsupported detail, observable
acceptance, and any revision required by consumers. Report the link and stop the
blocked portion until the owner produces a merged revision.

Afterward, verify that revision and modify only the currently authorized
repository to consume it, update pinned references or integration metadata, and
run relevant integration checks. Continue unrelated in-scope work only when it
does not assume the pending owner change.
