# Ownership and Authoritative Sources

This file governs responsibility assignment, cross-repository contracts, and
decisions about which artifact owns a fact. Clause numbering continues
from `repository-workflow.md`.

## Product and Planning

- **G061** The parent repository's `docs/project-description.md` owns current
  product scope, user-visible transaction semantics, deployment topology, and
  cross-component boundaries.
- **G062** Each repository's `docs/implementation_plan.md` owns unfinished
  targets and implementation stages for that repository. Plans guide work but
  do not prove current behavior.

## Executable Contracts

- **G063** The engine repository's exported Go code and behavioral tests own
  generic engine models, command and query entry points, logical persistence
  protocols, restoration rules, errors, and state-machine behavior. Other
  repositories consume the pinned engine revision and do not reconstruct,
  mutate, or copy those contracts.
- **G064** `web/openapi/stu-trade.yaml` owns `/api/v1` paths, transport DTOs,
  statuses, and error structures. Generated client and server bindings derive
  from that file and are not edited by hand.
- **G065** Web maintainers propose HTTP structures from caller needs; host
  maintainers review product semantics, security, idempotence, and feasibility,
  then implement only an accepted pinned revision.
- **G066** Parent-repository migrations own the physical PostgreSQL schema,
  constraints, and migration order, including tables that implement engine
  protocols and foreign keys to host-owned extensions. Physical schema
  ownership does not transfer engine state semantics to the host. Application
  models and OpenAPI must not redefine storage.
- **G067** Parent-repository query files own sqlc query text. Generated database
  bindings are derived artifacts and remain inside adapters that implement
  engine protocols or host-owned persistence; host business code does not use
  them to reconstruct or mutate engine entities directly.
- **G068** Web implementation and behavioral tests own pages, interaction,
  frontend routing, accessibility, and frontend release behavior.
- **G069** The host owns `/health`, `/ready`, process configuration, logging,
  dependency assembly, and deployment lifecycle; these are not web business
  API operations.

## Cross-Repository Changes

- **G070** Cross-repository work is a handoff sequence, not shared write
  authority. The agent for the owning repository updates the authoritative
  source and its derived implementation first. After that revision is merged,
  the agent for each consuming repository updates only its own generated
  artifacts, pinned revision or parent gitlink, and integration checks. Do not
  maintain parallel transition contracts.
- **G071** The parent repository records exact engine and web commits. A child
  repository merge does not affect integration until the parent gitlink is
  updated and verified.
- **G072** Repository READMEs explain local purpose, current state, development
  entry points, and links. They do not restate product or protocol detail owned
  elsewhere.
- **G073** When sources conflict, resolve the conflict at the owner named here;
  do not treat a generated file, copied prose, or older plan as stronger
  evidence.
