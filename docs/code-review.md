# Code Review — 2026-05-28

Scope: all non-test source files. Reviewed with 7 parallel finder angles (correctness, removed-behavior, cross-file, reuse, simplification, efficiency, altitude) followed by individual verification of each candidate.

Findings are ranked most-severe first. Correctness bugs precede cleanup.

---

## 1. Wrong-branch unstash when operation + return-checkout both fail

**File:** `internal/git/workflows.go` — lines 85, 112  
**Status:** Confirmed

In `MergeInto` and `RebaseOnto`, after the primary operation fails the code attempts `Checkout(original)` to return to the starting branch before `AutoUnstash` runs. The guard `&& mergeErr == nil` (resp. `&& rebaseErr == nil`) means a failure of that checkout is silently discarded. `AutoUnstash` then pops the stash while `HEAD` is still on the wrong branch, applying the user's uncommitted changes to the wrong working tree.

**Scenario:** User on `main` runs MergeInto(`feature`, `release`). Merge conflicts; `merge --abort` leaves the tree dirty. `git checkout main` also fails (dirty index). The guard swallows the checkout error; `AutoUnstash` pops the stash onto `release` instead of `main`.

---

## 2. `refreshAll` loads commits for HEAD, not the displayed branch

**File:** `internal/model/app.go` — line 554  
**Status:** Confirmed

`refreshAll` captures `current := a.client.CurrentBranch()` (the git checkout state) and uses it as the ref for the commits `loadCommitsMsg`. If the user is browsing a branch other than the checked-out one, every workflow completion silently replaces the commit panel with HEAD's commits, overwriting `a.commits.branchRef` with no indication.

**Scenario:** User is checked out on `main` but has navigated the branch panel to `feature`. They trigger Fetch. `refreshAll` loads commits for `main`, replacing the feature commits while the branch panel cursor still highlights `feature`.

---

## 3. Stash fallback can pop the wrong entry under concurrent stash activity

**File:** `internal/git/stash.go` — line 39  
**Status:** Confirmed

`AutoUnstash` finds the operation's stash entry by comparing commit hashes. If the lookup finds no match it falls back to `stash@{0}` (the newest stash). If another process pushes a stash between `AutoStash` and `AutoUnstash`, the fallback pops that external entry instead, silently destroying unrelated saved work while the operation's stash remains stranded.

**Scenario:** AutoStash pushes the operation's stash. A concurrent terminal session pushes another stash, bumping the operation's entry to `stash@{1}`. Hash comparison fails (e.g. whitespace difference). Fallback `stash@{0}` pops the external stash.

---

## 4. `len()` used for display-width calculations on branch names

**File:** `internal/model/branch_picker.go` — lines 66, 235  
**Status:** Confirmed

`listW` is computed with `len(name) + 2` (byte length) and the truncation at line 235 cuts at a byte offset. Both are wrong for multibyte UTF-8 branch names; the correct call is `lipgloss.Width`.

**Scenario:** A remote named `ëco/feature` has a 2-byte first character. `len` counts bytes, not display columns, so `listW` is too wide and the truncation `name[:m.listW-3]` may split a multibyte sequence, producing garbled output in the picker dialog.

---

## 5. `fetchTagMap` called redundantly on every detail view

**File:** `internal/git/client.go` — line 199  
**Status:** Confirmed

`ShowCommit` calls `fetchTagMap()` to populate `CommitDetail.Tags`, spawning a `git tag` subprocess. `ListCommits` independently calls `fetchTagMap()` for the same repo state. Every commit cursor move triggers both, paying the tag-enumeration cost twice.

**Scenario:** On a repo with many tags or a slow filesystem, every keystroke in the commit panel incurs two `git tag` subprocesses: one for the commits list refresh and one for the detail load.

---

## 6. `CurrentBranch()` returns `"HEAD"` in detached state; used as checkout target

**File:** `internal/git/client.go` — line 77  
**Status:** Plausible

`git rev-parse --abbrev-ref HEAD` outputs the literal string `"HEAD"` in detached HEAD state. `MergeInto` and `RebaseOnto` store this as `original` and later call `Checkout("HEAD")` to return to it. After an operation that advances HEAD (e.g. a successful merge), `git checkout HEAD` re-detaches at the new commit rather than the pre-operation one.

**Scenario:** Repo detached at `abc123`. MergeInto runs; merge succeeds and creates a new commit. `git checkout HEAD` attaches to the merge commit, not `abc123`. The user is silently left at a different point in history.

---

## 7. Filter key-handling duplicated between `branches.go` and `commits.go` — already diverged

**File:** `internal/model/branches.go` — line 125 (and `internal/model/commits.go` — line 113)  
**Status:** Confirmed (cleanup)

The Esc/Enter/default filter key-handling block is copy-pasted verbatim across both panel models and has already silently diverged: `branches.go` emits a `selectionCmd()` on Esc (keeping the detail panel in sync); `commits.go` does not. Any future fix applied to one copy will likely miss the other.

---

## 8. Selected-row truncation uses inconsistent strategies for plain vs highlighted rows

**File:** `internal/model/commits.go` — line 296  
**Status:** Confirmed (cleanup)

Subject truncation is duplicated between the unselected and selected rendering paths. The plain path uses `[]rune` length (correct for Unicode); the selected path uses a byte-width variable `w`. A commit subject containing multibyte characters is truncated correctly when unselected but incorrectly when highlighted.

---

## 9. Magic panel-height constant duplicated across four sites with unexplained divergence

**File:** `internal/model/branches.go` — line 190 (and `internal/model/commits.go`)  
**Status:** Confirmed (cleanup)

`innerH = height - 8` is computed independently in `clampOffset()` and `View()` in `branches.go`, and again with a different constant (`height - 6`) in `commits.go`, giving four independent sites. The constant already differs between panels with no explanation; a future chrome change (e.g. adding a status row) requires updating all four sites, and the existing mismatch means scroll and render boundaries are already subtly inconsistent between panels.
