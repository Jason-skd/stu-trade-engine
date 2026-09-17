# Repository Workflow

This file governs commits, branches, issues, pull requests, and documentation
changes. Clause numbering continues from the root `AGENTS.md`.

## Commits

- **G035** Use the minimal Conventional Commits form
  `<type>[(scope)][!]: <subject>`, with an optional body only when it explains
  why.
- **G036** Prefer `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `build`,
  `ci`, and `chore`; use another type only when established in the repository.
- **G037** One commit expresses one change type. Split independent changes.
- **G038** Subjects are English, imperative, lowercase at the start, and have
  no trailing period. Use 50 characters as a soft limit and 72 as a hard limit.
- **G039** Scope is optional and comes from the actual module layout. Omit it
  for cross-cutting work or when no precise scope exists.
- **G040** Omit the body by default. When needed, explain why without restating
  the diff.
- **G041** Mark incompatible changes with `!` and explain the impact. Do not add
  attribution or issue-link trailers. Preserve standard revert messages.

## Branches and Pull Requests

- **G042** Pure documentation changes may be committed and pushed directly to
  `main` in the repository where the task is authorized. Synchronizing the same
  change across repositories requires explicit authorization naming every
  target repository; otherwise hand each target off independently. Parent
  gitlink updates remain changes in the parent repository.
- **G043** All other changes use a branch outside `main` and merge through a
  pull request unless an explicit task instruction states otherwise.
- **G044** Rebase private work onto `main` and resolve conflicts in that
  branch's commits. Never rebase `main`.
- **G045** Pull requests reference relevant issues and state whether they close
  them. Add labels and milestones only when useful.
- **G046** Once a pull request exists, every completed push triggers the
  numbered governance checklist in `.agents/skills/pr-self-review`; review the
  complete pull-request diff at the pushed head.
- **G047** Do not treat a pull request as ready while any checklist clause is
  `[ ]` or `[~]`. Apply focused `fix:` or `refactor:` commits, push, and run a
  fresh checklist cycle before merge.

## Issues

- **G048** Issues use the sections Context, Proposal, and Acceptance.
- **G049** Context states the problem and relevant background.
- **G050** Proposal states the recommended direction.
- **G051** Acceptance states observable outcomes independent of implementation
  details.
- **G052** Filing a short issue does not require a full implementation study
  unless the proposal depends on unresolved technical facts. A required change
  owned by another repository is filed there as an Emergency Issue with the
  current evidence, blocking impact, proposed direction, and observable
  acceptance; notify the user to schedule that repository's agent.

## Documentation

- **G053** Documentation states currently valid facts. Remove stale statements
  and maintain affected indexes and entry points.
- **G054** The root README is the canonical index for every file under `docs/`.
  Add, remove, or move a document and update that index in the same change.
- **G055** A repository README is a concise newcomer entry point: purpose,
  current state, boundaries, setup, validation commands, and document links.
- **G056** Stable product or component explanation belongs in a dedicated
  project-description document rather than accumulating in the README.
- **G057** Planned behavior, target structures, future dependencies, stages,
  and acceptance targets belong in `docs/implementation_plan.md`.
- **G058** Responsibility allocation, authoritative sources, and cross-repository
  change duties belong in `docs/governance/ownership.md`.
- **G059** Do not create indexes, plans, or compatibility notes merely for
  symmetry. Format Markdown tables and check whitespace before committing.
- **G060** Runtime structures and generated documentation share authoritative
  sources; do not maintain manually synchronized copies.
