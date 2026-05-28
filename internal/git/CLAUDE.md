# internal/git — shell-out layer

This package is the only place that touches the `git` binary. Nothing above this layer calls `exec.Command`.

## Design responsibility

This layer owns the core UX promise: **the user never has to manually stash or unstash before triggering an operation.** Every `Workflows` method that modifies `HEAD` or the working tree must call `stashAround`, which transparently pushes and pops uncommitted changes around the operation.

The one deliberate exception is operations whose purpose is to interact with uncommitted changes. `AmendLast` must NOT use `stashAround` — the user's intent is to fold staged work into the amended commit. Any future commit-style operations fall into the same category.

## Files

| File | Responsibility |
|------|---------------|
| `client.go` | `Client` struct, typed data structs, raw git operations |
| `stash.go` | `AutoStash` / `AutoUnstash` around destructive ops |
| `workflows.go` | `Workflows` — stash-aware, high-level operations |
| `github.go` | `gh` CLI integration (best-effort, never fatal) |
| `util.go` | `FindRepoRoot` — walks upward to find `.git` |

## Key invariants

**`Client` = raw git, `Workflows` = stash-aware orchestration.**
`Client` methods run a single git command and return a typed result or `error`. `Workflows` methods compose `Client` calls; every destructive operation goes through `stashAround`.

**`stashAround(label, fn)`** pushes a stash if the working tree is dirty, runs `fn`, pops the stash by ref afterward. Use it for all operations that change `HEAD` or working tree state (checkout, merge, rebase, pull, drop, squash). `StashResult.Ref` holds the stash object hash so `AutoUnstash` can locate the right entry even if other stashes exist.

Do **not** use `stashAround` for operations whose purpose is to interact with uncommitted changes (currently: `AmendLast`). Stashing first would undo exactly what the user intends.

**`WorkflowResult`** is always returned by `Workflows` methods — never a bare `error`. Set `Err` for failures; set `Message` for the success status string shown in the UI. The stash note `(uncommitted changes stashed and restored)` is appended automatically by `stashAround` when a stash was needed.

**No model/ui imports.** This package has no knowledge of Bubble Tea, lipgloss, or the UI layer. It must remain importable standalone.

## Adding a new git operation

1. **`client.go`** — add the raw method to `Client`. Keep it to one git command; compose in `Workflows` if multiple steps are needed.
2. **`workflows.go`** — add a `Workflows` method. Ask: does git refuse or produce wrong results when the working tree is dirty? If yes, wrap with `stashAround`. If the operation's purpose is to consume uncommitted changes (amend, commit), do not. Match the message format of existing methods (`"verb noun complete"`).
3. Wire the new `WorkflowResult`-returning method into `model/app.go` (`handleMenuAction`).

## Current stash-coverage gaps

`CherryPick` and `RevertCommit` do not use `stashAround` and will return a git error if the working tree is dirty. They are candidates for the same treatment as `DropCommit` and `SquashCommits`, but the right behaviour (stash-transparent vs. user-controlled) has not yet been decided.

## github.go

`ListOpenPRs` is best-effort: it returns `nil` on any error (no `gh`, no GitHub remote, bad JSON). The UI degrades gracefully when `nil` is returned — never assume the map is non-nil.

## Testing

Tests live in `client_test.go`. They run against a real `git` binary (no mocks). Use `t.TempDir()` + `git init` to create isolated repos. Never mock `exec.Command`.
