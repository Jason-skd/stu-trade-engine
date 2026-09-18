# Governance Meta

This file governs rule identities, promotion into the common baseline, seed
material, and reusable skills.

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

## Promotion and Reuse

- **A039** When governance rules form a reusable multi-step workflow, alert the
  user that they are a skill candidate before creating or extending a skill.
- **A041** Keep project exceptions in repository supplements until current
  evidence demonstrates that they apply across repositories.
- **A042** Treat seeds as optional initialization material, not inherited
  governance. Adapt selected placeholders to the target repository and assign
  `R-A` or `R-G` identities only after a rule becomes a local supplement.
- **A043** Keep common governance, repository supplements, seeds, and skills
  distinct: governance defines rules, supplements define local rules, seeds
  initialize local material, and skills guide reusable workflows.

The standalone `GOVERNANCE` repository is a copy source. A target repository
is governed only by the files present in that repository, including its local
supplements. Skills remain automatically discoverable unless their own
metadata says otherwise, but discovery never grants mutation authority.
