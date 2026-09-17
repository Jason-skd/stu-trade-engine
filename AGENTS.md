# Repository Governance

This file is the shared maintenance contract for this repository. It contains
repository-facing rules, not personal preferences or agent identity. Explicit
task instructions determine the current goal and scope; this contract governs
work inside that scope.

# Authority

- **G001** Consultation, investigation, diagnosis, and review do not authorize
  implementation. Change files only after a clear start-work instruction.
- **G002** Once implementation is authorized, continue through investigation,
  changes, focused validation, repair, broader verification, and final
  inspection.
- **G003** Skills under `.agents/skills/` add task workflows. They do not
  override this contract, broaden scope, or authorize writes.
- **G004** Implementation authority is limited to this repository. It does not
  extend to submodules, sibling or nested repositories, or other repositories
  in the same product or workspace. Outside repositories may be inspected
  read-only for evidence; the only permitted cross-repository mutation is
  filing an Emergency Issue in the owning repository for handoff.

# Evidence and Workspace

- **G005** Verify conclusions against the current branch, code, tests, and
  locked dependencies. History is context, not proof of current behavior.
- **G006** Read the nearest implementation, callers, and tests before editing.
  Follow existing patterns only when they conform to current contracts.
- **G007** Change only what the task requires. Preserve unrelated and
  uncommitted work.
- **G008** Never use destructive Git operations or broad overwrites to
  manufacture a clean worktree.
- **G009** Never claim unobserved behavior, test results, generated state, or
  external outcomes.

# Numbered Governance — Important

- **G010** Every governance review must account for the clause IDs rather than
  claim vaguely that the change looks compliant. Every normative clause in
  this file and `docs/governance/*.md` has
  one globally unique ID. IDs increase in this canonical file order:
  `AGENTS.md`, `repository-workflow.md`, `ownership.md`, `implementation.md`,
  `interfaces.md`, then `verification.md`. Adding, removing, or moving a clause
  must renumber later clauses and update every repository reference in the same
  change.

# Pull Request Gate

- **G011** After a pull request exists, every completed push starts a new
  checklist phase. Use `.agents/skills/pr-self-review`, challenge the change
  from the strongest opposing position, and review all code authored in the
  pull request rather than only the latest commit.
- **G012** Record every governance clause with exactly one status: `[x]`
  compliant with evidence, `[ ]` missed or not yet checked, `[-]` inapplicable
  with a reason, or `[~]` checked and found incomplete. A checklist containing
  `[ ]` or `[~]` is not complete.
- **G013** Resolve checklist gaps with one or more focused `fix:` or `refactor:`
  commits, push them, and repeat the checklist against the new pushed head.

# Governance Routing

- **G014** For commits, branches, issues, pull requests, or documentation, read
  [`docs/governance/repository-workflow.md`](docs/governance/repository-workflow.md).
- **G015** For ownership, authoritative sources, or cross-repository changes,
  read [`docs/governance/ownership.md`](docs/governance/ownership.md). When a
  required change is owned elsewhere, use
  [`.agents/skills/cross-repository-escalation`](.agents/skills/cross-repository-escalation/SKILL.md)
  and stop before modifying the owning repository.
- **G016** For design, implementation, refactoring, dependencies, comments,
  schemas, or generated sources, read
  [`docs/governance/implementation.md`](docs/governance/implementation.md).
- **G017** For errors, HTTP, DTOs, or external protocol boundaries, read
  [`docs/governance/interfaces.md`](docs/governance/interfaces.md).
- **G018** For behavior changes, tests, CI, asynchronous work, or managed
  resources, read
  [`docs/governance/verification.md`](docs/governance/verification.md).
- **G019** Read every applicable file for a cross-cutting task, but do not load
  unrelated governance merely because it exists. More specific directory
  governance may add local rules but cannot weaken this contract.

# Standing Engineering Invariants

- **G020** Model current needs only. Every field, operation, abstraction, and
  public capability needs a present caller or owner.
- **G021** Inspect existing capabilities before adding code or dependencies.
  Reuse real common operations; do not build mode-driven universal flows.
- **G022** Split by responsibility. New layers must validate, convert,
  orchestrate, or isolate a dependency rather than merely forward calls.
- **G023** Consumers define minimal capabilities. Business code does not depend
  on process assembly, transport entry points, or concrete infrastructure.
- **G024** Keep one authoritative source for schemas, protocols, queries, and
  generated representations. Never hand-edit generated output.
- **G025** Translate dependency failures at adapter boundaries, preserve
  meaningful causes, and log once at the final handling boundary.
- **G026** Behavior changes require behavioral tests; bug fixes require a
  regression scenario reproducing the defect.
- **G027** Documentation and comments describe current facts, contracts, and
  reasons, not implementation history or speculative future work.

# Completion

- **G028** Keep the final diff focused and preserve unrelated workspace state.
- **G029** Run formatting, static checks, tests, generation, and contract
  validation in proportion to the change and repository capabilities.
- **G030** Report actual changes, design boundaries, commands run, omitted
  checks, and residual risk. Old or partial results are not evidence for the
  current checkout.

# Language

- **G031** Final governance is English; Chinese drafts are acceptable during
  review.
- **G032** Comments and non-governance maintainer documentation are Chinese.
- **G033** Console output, log templates, internal error context, and commits
  are English.
- **G034** Pull request and issue language is unrestricted. User-visible text
  follows the product and upstream language contracts.
