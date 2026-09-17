---
name: pr-self-review
description: Run the mandatory numbered governance checklist after a pull request has been opened and after every subsequent push. Review the complete pull-request diff, challenge the author's own decisions, record clause-level evidence, and resolve gaps with focused fix or refactor commits before merge.
---

# Pull Request Self-Review

Use this workflow only when a pull request already exists. A completed push is
the trigger; the pushed remote head is the revision under review.

## Anchor the Cycle

Confirm the pull request base, pushed head, commit list, and full diff from the
base to that head. Review every file and every authored line in the pull
request, including tests, documentation, configuration, generated artifacts,
and earlier commits. Do not limit the review to the most recent push.

Load `AGENTS.md` and every routed governance file. Confirm that clause IDs are
continuous in the canonical order before using them as the checklist.

## Challenge the Change

For each changed responsibility or behavior, make the strongest evidence-based
case that the author's choice is wrong, incomplete, unsafe, duplicated, or
owned elsewhere. Then test that case against the implementation, callers,
tests, authoritative contracts, generated state, and current dependency
versions.

This is an adversarial review of one's own work, not a request to invent
criticism. Record a finding only when evidence supports it, and record the
counterevidence when the challenge does not hold.

## Complete the Numbered Checklist

Create one checklist line for every governance clause. Do not omit clauses,
collapse ranges, or inherit a status from an earlier push. Use exactly one of
these markers:

- `[x] Gnnn` - compliant: reviewed and supported by concrete evidence.
- `[ ] Gnnn` - missed: the required check was skipped or evidence is still
  unavailable.
- `[-] Gnnn` - inapplicable: the clause does not govern this diff; state why.
- `[~] Gnnn` - incomplete: review found a concrete governance gap or partial
  implementation; state the required repair.

Each `[x]` line cites a path, test, command result, contract, or concise
reasoning. Each `[-]` line states why it is outside the change. A bare marker is
not a completed review. Publish the revision-anchored checklist as pull-request
review evidence when repository tooling is available.

## Repair and Repeat

A cycle containing `[ ]` or `[~]` cannot pass. Group related gaps by cause and
resolve them with one or more focused commits:

- Use `fix:` when correcting behavior, a contract violation, or missing
  required evidence.
- Use `refactor:` when improving structure without changing behavior.

Keep unrelated cleanup out of these commits. Run the closest relevant checks,
push the commits, and start a new checklist cycle against the new pushed head.
Earlier checklist marks are historical evidence, not proof for the new head.

The pull request is ready only when the latest cycle contains exclusively
`[x]` and justified `[-]` entries, all required checks pass, and no unresolved
finding remains. Report the reviewed revision, repair commits, commands run,
omitted checks, and residual risk.
