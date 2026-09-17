# Ownership and Authoritative Sources

Read this file when assigning responsibilities, changing a cross-repository
contract, or deciding which artifact owns a fact. Clause numbering continues
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
  engine models, protocols, errors, and state-machine behavior. Other
  repositories consume the pinned engine revision and do not copy its APIs.
- **G064** `web/openapi/stu-trade.yaml` owns `/api/v1` paths, transport DTOs,
  statuses, and error structures. Generated client and server bindings derive
  from that file and are not edited by hand.
- **G065** Web maintainers propose HTTP structures from caller needs; host
  maintainers review product semantics, security, idempotence, and feasibility,
  then implement only an accepted pinned revision.
- **G066** Parent-repository migrations own PostgreSQL schema, constraints, and
  migration order. Application models and OpenAPI must not redefine storage.
- **G067** Parent-repository query files own sqlc query text; generated database
  bindings are derived artifacts and remain inside the database adapter.
- **G068** Web implementation and behavioral tests own pages, interaction,
  frontend routing, accessibility, and frontend release behavior.
- **G069** The host owns `/health`, `/ready`, process configuration, logging,
  dependency assembly, and deployment lifecycle; these are not web business
  API operations.

## Cross-Repository Changes

- **G070** A cross-repository change updates its authoritative source first,
  then generated artifacts and implementations, then the parent gitlinks and
  integration checks. Do not maintain parallel transition contracts.
- **G071** The parent repository records exact engine and web commits. A child
  repository merge does not affect integration until the parent gitlink is
  updated and verified.
- **G072** Repository READMEs explain local purpose, current state, development
  entry points, and links. They do not restate product or protocol detail owned
  elsewhere.
- **G073** When sources conflict, resolve the conflict at the owner named here;
  do not treat a generated file, copied prose, or older plan as stronger
  evidence.
