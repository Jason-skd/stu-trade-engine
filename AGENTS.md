# Repository Governance

This file is the shared baseline for repository maintenance. Explicit task
instructions determine the current goal and scope. Repository-specific rules
in `docs/governance/repository.md` supplement this baseline where they are more
specific; they cannot silently weaken it.

## Authority and Evidence

- **A001** Treat consultation, investigation, diagnosis, and review as
  read-only. Change files or external state only after a clear implementation
  instruction.
- **A002** Once implementation is authorized, continue through investigation,
  changes, focused validation, repair, broader verification, and final
  inspection without using routine status updates as approval gates.
- **A003** Make low-risk choices from established contracts without repeatedly
  asking for confirmation.
- **A004** Ask before selecting incompatible user-visible behavior, crossing a
  product or repository boundary, performing irreversible work, breaking a
  public protocol, or relying on unavailable credentials or external input.
- **A005** Treat task-specific skills as workflows, not as authority to expand
  scope or perform additional mutations.
- **A006** Verify conclusions against the current branch, revision, code,
  tests, generated state, and locked dependencies. Use history, plans, and
  memory only to guide investigation.
- **A007** Read applicable repository instructions and routed governance, then
  inspect the nearest implementation, callers, and tests before editing.
- **A008** Follow existing patterns only when they conform to the current
  authoritative contracts.
- **A009** Change only what the task requires. Preserve unrelated and
  uncommitted work, including overlapping edits from other contributors.
- **A010** Do not use destructive Git operations or broad overwrites to create
  a clean worktree.
- **A011** Do not claim behavior, test results, generated state, or external
  outcomes that were not observed on the applicable checkout.

## Rule Identities

- **A012** Give each normative rule exactly one permanent identity. Common
  action rules use `A`; common acceptance rules use `G`; repository action and
  acceptance supplements use `R-A` and `R-G`.
- **A013** Add a rule with the next unused maximum number in its family. Keep
  its ID when moving it, never reuse a deleted ID, and never continue common
  numbering in a repository-specific family.

Action rules describe how work is performed. Acceptance rules describe facts
that can be checked afterward from the complete pull-request diff, commits,
tests, and recorded verification. Pull-request checklists cover `G` and `R-G`
rules only; violations of `A` or `R-A` remain concrete review findings.

## Governance Routing

- **A014** For authorization, investigation, Git, issues, pull requests, or
  documentation workflow, read
  [`docs/governance/workflow.md`](docs/governance/workflow.md).
- **A015** For modeling, reuse, architecture, dependencies, interfaces, errors,
  compatibility, generated sources, or comments, read
  [`docs/governance/engineering.md`](docs/governance/engineering.md).
- **A016** For tests, CI, asynchronous work, resources, or completion evidence,
  read [`docs/governance/verification.md`](docs/governance/verification.md).
- **A017** Read
  [`docs/governance/repository.md`](docs/governance/repository.md) for local
  ownership, architecture, protocols, tools, validation, and release rules.
- **A018** Load every applicable governance file for a cross-cutting task, but
  do not load unrelated material merely because it exists.

## Skill Routing

- **A019** Use
  [`.agents/skills/design-change`](.agents/skills/design-change/SKILL.md) for a
  feature or architecture decision with material design tradeoffs.
- **A020** Use
  [`.agents/skills/diagnose-and-fix`](.agents/skills/diagnose-and-fix/SKILL.md)
  for a bug, regression, failing test, or incorrect runtime behavior.
- **A021** Use
  [`.agents/skills/resolve-review`](.agents/skills/resolve-review/SKILL.md) to
  evaluate supplied review findings against a specific revision.
- **A022** Use
  [`.agents/skills/pr-self-review`](.agents/skills/pr-self-review/SKILL.md) when
  a pull request is first opened and again immediately before the work is
  delivered to the user as its terminal review.
- **A023** Use
  [`.agents/skills/cross-repository-escalation`](.agents/skills/cross-repository-escalation/SKILL.md)
  before a required change or handoff crosses repository ownership.
