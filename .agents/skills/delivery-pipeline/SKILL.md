---
name: delivery-pipeline
description: Design or adjust PR, main, and production delivery gates around a repository's real artifact and deployment model. Do not invent deployment stages.
---

# Delivery Pipeline

Read `docs/governance/delivery.md`, the repository supplement, existing
workflows, build definitions, artifact stores, deployment targets, and release
evidence. Establish which stages the repository actually owns before proposing
or changing automation. Use `seeds/delivery/pipeline.md` as a question set, not
as a workflow to copy unchanged.

Ensure the pull-request gate proves merge readiness, the main gate evaluates
every exact `main` revision and certifies its artifact, and later delivery
consumes that immutable revision and digest. Keep development completion,
verification, deployable candidacy, and production release distinct. Rebuilding
creates a new candidate.

Do not add production automation without a real target, identity, artifact
registry, protection model, rollback procedure, and owner. Validate syntax and
the repository's real checks; report any external prerequisite that could not
be exercised.
