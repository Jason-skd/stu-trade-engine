---
name: github-work-management
description: Audit and coordinate GitHub labels, SemVer milestones, issue forms, and pull-request templates for an explicitly named repository.
---

# GitHub Work Management

Inspect the target repository governance and the seed definitions under
`seeds/github/`. Verify authentication and repository identity, then audit the
current labels, milestones, issue forms, and pull-request template without
mutation. Use `scripts/audit-labels.sh OWNER/REPO LABELS_YML` to compare labels
through authenticated `gh` reads and YAML parsing managed by `uv`.

Report the exact additions, updates, conflicts, and removals proposed. Obtain
explicit user authorization naming the target repository and remote action
before changing GitHub. Preserve work-item role exclusivity, SemVer milestone
semantics, leaf closure through `Closes`, and `flow:emergency` as a state in
addition to the item's role. Do not infer authorization from a prior audit.
