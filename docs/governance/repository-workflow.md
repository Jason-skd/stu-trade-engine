# Repository Workflow

Read this file for commits, branches, issues, pull requests, and documentation
changes. Repository-specific automation may replace mechanical steps, but not
the decisions and boundaries below.

## Commits

Use the minimal Conventional Commits form:

```text
<type>[(scope)][!]: <subject>

[optional body]
```

- Common types are `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `build`,
  `ci`, and `chore`. Use an ecosystem-specific type only when established here.
- One commit expresses one change type. Split independent changes.
- Subjects are English, imperative, lowercase at the start, and have no
  trailing period. Use 50 characters as a soft limit and 72 as a hard limit.
- Scope is optional and comes from the actual module layout. Omit it for
  cross-cutting work or when no precise scope exists.
- Omit the body by default. When needed, explain why without restating the diff.
- Mark incompatible changes with `!` and explain the impact. Do not add
  attribution or issue-link trailers. Preserve standard revert messages.

## Branches and Pull Requests

- Work outside the main branch and merge through a pull request unless an
  explicit repository rule authorizes otherwise.
- Rebase private work onto main and resolve conflicts in that branch's commits.
  Never rebase main.
- Pull requests reference relevant issues and state whether they close them.
  Add labels and milestones only when useful.

## Issues

Issues use three concise sections:

1. Context: the problem and relevant background.
2. Proposal: the recommended direction.
3. Acceptance: observable outcomes independent of implementation details.

Filing a short issue does not require a full implementation study unless the
proposal itself depends on unresolved technical facts.

## Documentation

- Documentation states currently valid facts. Remove stale statements and
  maintain affected indexes and entry points.
- Do not create indexes, plans, or compatibility notes merely for symmetry.
- Format Markdown tables and check whitespace before committing.
- Runtime structures and generated documentation share authoritative sources;
  do not maintain manually synchronized copies.
