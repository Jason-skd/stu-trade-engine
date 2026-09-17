# Repository Governance

This file is the shared maintenance contract for this repository. It contains
repository-facing rules, not personal preferences or agent identity. Project
context, architecture, commands, and validation matrices belong here only after
they become established facts.

Explicit task instructions determine the current goal and scope. This contract
governs work inside that scope. Tools own mechanical formatting; code, schemas,
tests, generated contracts, and upstream protocols own concrete behavior.
Exceptions must state their scope and reason.

# Authority

- Consultation, investigation, diagnosis, and review do not authorize
  implementation. Change files only after a clear start-work instruction.
- Once implementation is authorized, continue through investigation, changes,
  focused validation, repair, broader verification, and final inspection.
- Skills under `.agents/skills/` add task workflows. They do not override this
  contract, broaden scope, or authorize writes.
- Ask before choosing incompatible user-visible behavior, crossing a product
  boundary, performing irreversible work, or breaking a public protocol.

# Evidence and Workspace

- Verify conclusions against the current branch, code, tests, and locked
  dependencies. History is context, not proof of current behavior.
- Read the nearest implementation, callers, and tests before editing. Follow
  existing patterns only when they conform to current contracts.
- Change only what the task requires. Preserve unrelated and uncommitted work.
- Never use destructive Git operations or broad overwrites to manufacture a
  clean worktree.
- Never claim unobserved behavior, test results, generated state, or external
  outcomes.

# Governance Routing

Read only the governance relevant to the current task before making the
corresponding change:

- For commits, branches, issues, pull requests, or documentation, read
  [`docs/governance/repository-workflow.md`](docs/governance/repository-workflow.md).
- For design, implementation, refactoring, dependencies, comments, schemas, or
  generated sources, read
  [`docs/governance/implementation.md`](docs/governance/implementation.md).
- For errors, HTTP, DTOs, or external protocol boundaries, read
  [`docs/governance/interfaces.md`](docs/governance/interfaces.md).
- For behavior changes, tests, CI, asynchronous work, or managed resources,
  read [`docs/governance/verification.md`](docs/governance/verification.md).

Read every applicable file for a cross-cutting task, but do not load unrelated
governance merely because it exists. More specific directory governance may
add local rules but cannot weaken this root contract.

# Standing Engineering Invariants

- Model current needs only. Every field, operation, abstraction, and public
  capability needs a present caller or owner.
- Inspect existing capabilities before adding code or dependencies. Reuse real
  common operations; do not build mode-driven universal flows.
- Split by responsibility. New layers must validate, convert, orchestrate, or
  isolate a dependency rather than merely forward calls.
- Consumers define minimal capabilities. Business code does not depend on
  process assembly, transport entry points, or concrete infrastructure.
- Keep one authoritative source for schemas, protocols, queries, and generated
  representations. Never hand-edit generated output.
- Translate dependency failures at adapter boundaries, preserve meaningful
  causes, and log once at the final handling boundary.
- Behavior changes require behavioral tests; bug fixes require a regression
  scenario reproducing the defect.
- Documentation and comments describe current facts, contracts, and reasons,
  not implementation history or speculative future work.

# Completion

- Keep the final diff focused and preserve unrelated workspace state.
- Run formatting, static checks, tests, generation, and contract validation in
  proportion to the change and repository capabilities.
- Report actual changes, design boundaries, commands run, omitted checks, and
  residual risk. Old or partial results are not evidence for the current
  checkout.

# Language

- Final governance is English; Chinese drafts are acceptable during review.
- Comments and non-governance maintainer documentation are Chinese.
- Console output, log templates, internal error context, and commits are English.
- Pull request and issue language is unrestricted. User-visible text follows
  the product and upstream language contracts.
