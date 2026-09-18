# Repository Governance Supplement

This file contains only rules specific to StuTrade Engine. The shared baseline
already governs authorization, cross-repository handoffs, engineering quality,
verification, GitHub work, and delivery evidence.

## Repository Actions

- **R-A003** Before completion, run
  `test -z "$(gofmt -l .)"`,
  `find .agents/skills -type f -name '*.sh' -print0 | xargs -0 bash -n`,
  `go vet ./...`, `go test ./...`, `go test -race ./...`,
  `go mod tidy -diff`, and `git diff --check`.

## Authority and Ownership

- **R-G001** The parent repository's `docs/project-description.md` owns
  product scope, user-visible transaction semantics, deployment topology, and
  cross-component boundaries. This repository's `docs/project-description.md`
  owns the engine component boundary, while `docs/implementation_plan.md` owns
  unfinished targets without proving current behavior.
- **R-G002** This repository's exported Go code and behavioral tests own
  generic engine models, entry points, logical persistence protocols,
  restoration rules, errors, and state-machine behavior.
- **R-G003** The web repository's `openapi/stu-trade.yaml` owns `/api/v1`
  paths, transport DTOs, statuses, and error structures; the parent host owns
  their runtime implementation and translation to engine operations.
- **R-G004** Parent-repository migrations own the physical PostgreSQL schema,
  constraints, and migration order. This repository owns the logical state and
  restoration semantics exposed through its persistence protocols.
- **R-G005** Parent-repository query files own sqlc query text and generated
  database bindings; those bindings stay in parent adapters implementing engine
  protocols or host persistence.

## Technical Profile and Delivery

- **R-G008** The parent repository records the exact engine commit as a
  gitlink. An engine merge does not affect product integration until that
  gitlink is updated and verified in the parent repository.
- **R-G009** `github.com/Jason-skd/stu-trade-engine` is a Go 1.27.1 library
  whose public API remains in the root package and whose production module has
  no non-standard-library dependencies.
- **R-G010** The engine remains synchronous and deterministic and does not own
  HTTP, physical database access, process configuration, logging, background
  goroutines, host domain types, executable artifacts, deployment, or product
  release.
