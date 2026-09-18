---
name: diagnose-and-fix
description: Diagnose and, when authorized, repair a bug, regression, failing test, or incorrect runtime behavior. Do not use for feature design or review-comment validation.
---

# Diagnose and Fix

Apply the routed governance in `AGENTS.md`. Diagnosis does not authorize file
changes.

Establish separately the current observable behavior, expected behavior and
its authoritative source, and the smallest reliable reproduction on the
current checkout. Trace from the observable entry point through the owning
rule, adapters, and side effects. Distinguish the root cause from symptoms and
identify the smallest complete repair.

If the request is diagnosis-only, stop after reporting evidence, cause, impact,
and a recommended solution. When repair is authorized, add or extend a
regression scenario at the closest responsible layer, implement the complete
fix across affected contracts and derived artifacts, validate the reproduction,
then expand checks according to risk.

Report the current situation and cause, expected situation, solution,
regression coverage, commands actually run, and residual risk. Never claim a
before-or-after test result that was not observed.
