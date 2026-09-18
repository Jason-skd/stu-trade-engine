# Workflow Governance

This file governs authorization-aware investigation, Git operations, issues,
pull requests, and documentation.

## Investigation and Git Actions

- **A024** Resolve the exact repository, worktree, branch, revision, and dirty
  state before editing or running repository mutations.
- **A025** Inspect existing capabilities and dependencies before introducing a
  new path, abstraction, tool, or package.
- **A026** Treat every independently pure documentation update—where the
  task's only requirement is changing documentation—as an emergency repair.
  Commit and push it directly to `main`; never create a `docs/` branch for it.
- **A027** For every other change, work on a branch other than `main` and merge
  through a pull request unless repository governance or an explicit task
  instruction establishes a different workflow.
- **A028** Rebase private work onto `main`; never rebase `main` itself.
- **A029** Stage explicit task-owned paths and inspect the staged diff before
  committing so unrelated workspace changes remain untouched.
- **A030** Use non-interactive, recoverable operations where practical. Resolve
  destructive targets with read-only checks before acting.
- **A031** Before mutating a remote service, verify authentication, repository
  identity, target revision, and the exact resource being changed.

## Commit Acceptance

- **G001** Commits use the repository's explicit convention or, when none is
  defined, the minimal Conventional Commits form
  `<type>[(scope)][!]: <subject>`.
- **G002** Each commit expresses one change type; independent changes are split.
- **G003** The subject is concise English, imperative, has no trailing period,
  and uses a scope only when it names a real repository module.
- **G004** A commit body is omitted unless it explains why. Incompatible
  changes use `!` and explain their impact.

## Issue Acceptance

- **G005** An issue contains `Context`, `Proposal`, and `Acceptance` sections.
- **G006** Context states the problem and relevant evidence; Proposal states
  the recommended direction; Acceptance states observable outcomes without
  prescribing incidental implementation details.
- **G007** Issues and pull requests reference relevant tracking work and state
  explicitly whether they close it. Labels and milestones are used only when
  they convey real project information.

## Pull Request Acceptance

- **G008** A pull request describes the actual change and its important
  boundaries, and its complete diff contains only work within the authorized
  scope.
- **G009** The commit sequence is coherent and contains no accidental generated
  files, local artifacts, unrelated edits, or temporary debugging changes.
- **G010** The opening and terminal review evidence is anchored to the pushed
  head reviewed in that cycle and its complete base-to-head diff.

## Documentation Acceptance

- **G011** Documentation states current facts, contracts, reasons, and
  constraints. Stale statements and superseded instructions are removed.
- **G012** Affected entry points and indexes are updated when documentation is
  added, removed, moved, or materially changed.
- **G013** A README remains a concise newcomer entry point rather than a copy of
  detailed product, protocol, plan, or governance documents.
- **G014** Planned behavior is clearly separated from implemented behavior and
  never presented as proof of the current system.
- **G015** New indexes, plans, compatibility notes, and directory structures
  exist only when they have a current reader or maintenance purpose.
- **G016** Markdown tables, links, formatting, and whitespace are valid in the
  changed documentation.

## Language Acceptance

- **G017** Normative governance rules and skill instructions are English;
  repository introductions and non-normative translations may follow their
  intended audience.
- **G018** Comments and non-governance maintainer documentation are Chinese
  unless repository or product conventions require another language.
- **G019** Console output, log templates, internal error context, and commit
  messages are English.
- **G020** Pull-request and issue language is unrestricted. User-visible text
  follows the product and upstream language contracts.
