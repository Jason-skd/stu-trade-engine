---
name: pr-self-review
description: Perform adversarial review of the complete pull-request diff when the pull request opens and immediately before terminal delivery, checking every common and repository acceptance rule against the reviewed pushed head.
---

# Pull Request Self-Review

Use this workflow only when a pull request exists. Run it exactly when the pull
request is first opened and again immediately before delivering the completed
work to the user. Intermediate pushes do not trigger a cycle. The remote pushed
head is the revision under review.

Confirm the base, pushed head, commit list, and complete base-to-head diff.
Review every authored line, including earlier commits, tests, documentation,
configuration, and generated artifacts. Load `AGENTS.md`, all applicable common
governance, and the repository supplement.

Challenge each changed responsibility or behavior from the strongest
evidence-based opposing position. Test the challenge against current code,
callers, tests, authoritative contracts, generated state, and locked dependency
versions. Record findings only when supported by evidence.

Create exactly one checklist row for every `G` and `R-G` rule, without
collapsing ranges or inheriting status from the opening cycle. Every row is a
Markdown list item using one marker:

- `- [x]` compliant, with concrete evidence;
- `- [ ]` not checked or evidence unavailable;
- `- [-]` inapplicable, with a reason;
- `- [~]` incomplete, with the required repair.

Do not create checklist rows for `A` or `R-A` rules. If the review discovers an
action-rule violation, record it as a concrete finding and resolve its effect.
Publish the revision-anchored checklist as pull-request evidence when repository
tooling is available.

A cycle containing `- [ ]` or `- [~]` cannot pass. Resolve supported gaps with
focused commits and run the relevant checks. If the opening cycle required
repairs, the terminal cycle re-anchors the full review after those repairs are
pushed. The pull request is ready only when the terminal cycle contains
exclusively evidenced `- [x]` and justified `- [-]` rows and no unresolved
finding remains.
