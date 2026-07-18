# Task 3 Report: Learn soft-gate (replace trial unit lock)

**Status:** DONE  
**Branch:** `feat/guest-settings-sync`  
**Commit:** `feat(learn): guest soft-gate by completed lessons, drop trial unit lock`

## What changed

- Removed `assertGuestTrialScope` (unit == `trialUnitID`) for guests.
- Added `assertGuestSoftGate` + `countCompletedLessons`:
  - Allow `StartLesson` / `CompleteLesson` if lesson already `completed` OR completed count `n < 3`.
  - Else return `ErrGuestSoftGate`.
- `GetUnit` / `GetLesson`: browse allowed for any unit/lesson (no soft-gate).
- `SubmitAnswer`: no soft-gate (in-flight attempts may finish).
- Service maps `ErrGuestSoftGate` (and alias `ErrTrialLimit`) → `PermissionDenied` / `"GUEST_SOFT_GATE"`.
- `trialUnitID` params retained unused for API stability.

## Tests

```text
go test ./internal/biz/ -v -count=1
PASS (all biz tests)
```

Key cases:
- Guest browse any unit/lesson OK
- Soft-gate blocks start/complete when 3 completed and lesson not completed
- Soft-gate allows already-completed lesson restart
- Soft-gate allows when `n < 3`
- Existing guest complete-one-lesson flow still passes

## Files

- `app/learn/internal/biz/curriculum.go`
- `app/learn/internal/biz/attempt.go`
- `app/learn/internal/biz/curriculum_test.go`
- `app/learn/internal/biz/attempt_test.go`
- `app/learn/internal/service/learn.go`

## Concerns / follow-ups

- FE should migrate from `TRIAL_LIMIT` to `GUEST_SOFT_GATE` (BE now only emits `GUEST_SOFT_GATE`).
- `trialUnitID` / config can be removed in a later cleanup once callers drop the unused arg.
- Soft reminder after lesson 1 remains FE-only (not enforced here).
