---
name: governance-sync
description: Compare or update a repository against the managed common governance baseline while preserving local supplements. Defaults to a read-only report.
---

# Governance Sync

Resolve the governance source and target repository. Run
`scripts/sync.sh SOURCE TARGET` for a read-only status report. The managed-file
manifest is authoritative for common files; `docs/governance/repository.md`,
local rules, seeds, and every unlisted path remain target-owned.

Report missing and differing managed files. Apply changes only when the user
explicitly authorizes writing the named target repository, using
`scripts/sync.sh SOURCE TARGET --apply`. Application replaces only listed
managed files and does not commit, push, or mutate remote services. Inspect the
resulting target diff and reconcile any target-specific references before
claiming synchronization.
