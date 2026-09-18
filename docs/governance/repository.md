# Repository Governance Supplement

This file supplements the shared governance baseline with ownership, handoff,
gitlink, and validation rules specific to StuTrade Engine.

## Repository Actions

- **R-A001** Treat implementation authority as limited to this repository
  unless the task explicitly names another repository. Outside repositories,
  including parent and sibling worktrees, may be inspected read-only but are
  not writable merely because they share a product or workspace.
- **R-A002** Perform cross-repository changes as an ownership handoff. Update
  the authoritative source in its owning repository under separate explicit
  authority; after that revision is merged, update only this repository's
  generated artifacts, pinned revision, and integration checks.
- **R-A003** Before completion, run
  `test -z "$(gofmt -l .)"`, `go vet ./...`, `go test ./...`,
  `go test -race ./...`, `go mod tidy -diff`, and `git diff --check`.

## Product and Component Ownership

- **R-G001** The parent repository's `docs/project-description.md` owns
  current product scope, user-visible transaction semantics, deployment
  topology, and cross-component boundaries. Each repository's
  `docs/implementation_plan.md` owns its unfinished targets and stages without
  proving current behavior.
- **R-G002** This repository's exported Go code and behavioral tests own
  generic engine models, entry points, logical persistence protocols,
  restoration rules, errors, and state-machine behavior.
- **R-G003** The web repository's `openapi/stu-trade.yaml` owns `/api/v1`
  paths, transport DTOs, statuses, and error structures. Web maintainers
  propose caller-facing HTTP structures; host maintainers review product
  semantics, security, idempotence, and feasibility before implementation.
- **R-G004** Parent-repository migrations own the physical PostgreSQL schema,
  constraints, and migration order, including engine-protocol tables and
  host-extension foreign keys, without transferring engine state semantics to
  the host.
- **R-G005** Parent-repository query files own sqlc query text. Generated
  database bindings remain derived and stay inside adapters implementing engine
  protocols or host-owned persistence.
- **R-G006** Web implementation and behavioral tests own pages, interaction,
  frontend routing, accessibility, and frontend release behavior.
- **R-G007** The host owns `/health`, `/ready`, process configuration,
  logging, dependency assembly, and deployment lifecycle; these are not web
  business API operations.

## Integration Revisions

- **R-G008** The parent repository records exact engine and web commits. A
  child merge does not affect product integration until the corresponding
  parent gitlink is updated and verified.
