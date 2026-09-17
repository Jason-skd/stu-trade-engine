---
name: diagnose-and-fix
description: Diagnose and, when authorized, fix a bug, regression, failing test, or incorrect runtime behavior. Do not use for feature design or for validating code-review comments.
---

# Diagnose and Fix

Establish the defect from current evidence, identify its owner and cause, and
close the repair with a regression scenario and proportionate verification.
Diagnosis alone does not authorize file changes.

## Establish the Defect

Describe separately:

1. Current observable behavior.
2. Expected observable behavior and its authoritative source.
3. The smallest reliable reproduction on the current checkout.

Do not infer a defect solely from an old issue, historical output, or a failing
mock whose assumptions do not match the real dependency.

## Locate the Cause

- Trace from the observable entry point through the owning rule, adapter, and
  side effects.
- Compare nearby implementations and tests before adding a new path.
- Distinguish the root cause from downstream symptoms.
- Check whether the failure belongs to application logic, schema constraints,
  protocol mapping, generated state, configuration, or an upstream dependency.
- Identify the smallest complete repair; do not attach unrelated cleanup.

If the request is diagnosis-only, stop after reporting evidence, cause,
impact, and the recommended solution.

## Repair When Authorized

- Add or extend a regression scenario that reproduces the defect at the
  closest responsible layer.
- Implement the complete fix across affected declarations, callers, adapters,
  generated artifacts, documentation, and contracts without speculative work.
- Preserve useful error identity and causes. Do not degrade corrupt input or a
  failed side effect into success.
- Validate the reproduction first, then expand checks according to risk.

Consider invalid inputs and limits, error mapping and recovery, transaction and
side-effect order, retries and cancellation, concurrency and idempotence,
caching, and external protocol compatibility only when relevant to the defect.

## Report the Result

Use this order:

1. Current situation and root cause.
2. Expected situation.
3. Recommended or implemented solution.
4. Regression coverage and commands actually run.
5. Relevant checks not run or residual risk.

Never claim that a test failed before the repair or passed afterward unless it
was actually observed on the applicable checkout.
