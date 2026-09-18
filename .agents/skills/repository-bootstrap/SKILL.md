---
name: repository-bootstrap
description: Initialize repository-local governance from optional seeds in an explicitly authorized target repository. Do not use to overwrite existing governance or mutate remote services.
---

# Repository Bootstrap

Confirm that the user explicitly named the target repository and authorized
file changes there. Inspect its current governance, technical shape, ownership,
verification, generated artifacts, and delivery responsibilities before
selecting seeds.

Use `scripts/bootstrap.sh SOURCE TARGET PROFILE...` for conflict-safe copying.
It copies the common managed files, creates the repository supplement from the
seed, and adds selected profiles without overwriting any existing path. A
conflict stops the operation before any file is copied.

Resolve every applicable placeholder from current target evidence, remove
unused seed sections, and assign local `R-A` or `R-G` identities only to
normative rules adopted by that repository. Do not modify GitHub, create a
branch, commit, or push unless those actions are separately authorized.
