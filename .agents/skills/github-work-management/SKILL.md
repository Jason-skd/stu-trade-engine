---
name: github-work-management
description: Audit and coordinate GitHub labels, SemVer milestones, issue forms, and pull-request templates for an explicitly named repository.
---

# GitHub Work Management

Inspect the target repository governance, then resolve the common-governance
source when seed comparison is requested. If that source provides
`seeds/github/`, treat those definitions as optional starting points rather
than target requirements. If no source is available, audit the target directly
against its current governance and report that seed comparison was omitted.

Verify authentication and repository identity, then audit the current labels,
milestones, issue forms, and pull-request template without mutation. Use the
managed `scripts/audit-labels.sh OWNER/REPO LABELS_YML` when a label definition
has been explicitly selected; it compares labels through authenticated `gh`
reads and YAML parsing managed by `uv`.

Report the exact additions, updates, conflicts, and removals proposed. Obtain
explicit user authorization naming the target repository and remote action
before changing GitHub. Preserve work-item role exclusivity, SemVer milestone
semantics, leaf closure through `Closes`, and `flow:emergency` as a state in
addition to the item's role. Do not infer authorization from a prior audit.
