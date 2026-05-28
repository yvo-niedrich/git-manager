# internal/model — Bubble Tea model layer

This package contains all application state and message routing. It sits between the `git` shell-out layer and the `ui` rendering layer.

## Files

| File | Responsibility |
|------|---------------|
| `app.go` | Root `App` model — owns panels, dialogs, focus, git clients |
| `branches.go` | `BranchesModel` — branch list with filter and virtual «new branch» button |
| `commits.go` | `CommitsModel` — commit list with multi-select mode |
| `detail.go` | `DetailModel` — scrollable commit detail via `bubbles/viewport` |
| `contextmenu.go` | `ContextMenuModel` — full-screen action menu overlay |
| `confirm.go` | `ConfirmModel`, `AmendModel`, `NewBranchModel` — single-purpose dialog overlays |
| `branch_picker.go` | `BranchPickerModel` — filterable branch list for merge/rebase target selection |
| `dialog.go` | `DialogContent` interface + `dialogStack` (priority-based overlay stack) |

## Key invariants

### Message flow

```
KeyMsg → app.Update
  → dialogs.Update   (if any dialog is open — dialogs capture all input)
  → handlePanelKey   (delegates to focused panel)
    → panel.Update → tea.Cmd (goroutine)
      → git.Workflows.SomeOp() → WorkflowResultMsg → app.Update
        → setStatus + refreshAll
```

Global keys (`q`, `tab`, `ctrl+c`) are handled in `app.Update` before delegation. The `isFiltering()` check suppresses global shortcuts while a panel's text filter is active.

### Dialog system

`dialogStack` holds multiple open dialogs simultaneously. On each update, the dialog with the highest `Priority()` captures input and is rendered as the overlay. All current dialogs use priority 20; assign a higher value for a dialog that must interrupt an already-open one.

**`DialogContent.DialogUpdate`** signals close by returning `(nil, cmd)`. The optional `cmd` carries a follow-up message (e.g. `BranchPickerSubmitMsg`, `AmendSubmitMsg`) back into `app.Update`. Returning `(self, nil)` keeps the dialog open with no side-effects.

**The stack never exceeds depth 1 in practice.** Every action closes the current dialog first (returns `nil`), then `app.handleMenuAction` optionally pushes a new one. `Priority()` is therefore never called with more than one item and has no exercised cases yet — but the mechanism is correct and available if true stacking is needed.

**Never** push a dialog from inside `DialogUpdate` — push from `app.handleMenuAction` instead, so the stack stays under `App`'s control.

### Git operations

Stash transparency is handled entirely in `internal/git` — the model layer does not need to check for uncommitted changes before dispatching any operation.

All git calls are dispatched as `tea.Cmd` goroutines so the UI stays responsive:

```go
return a.runWorkflow(func() git.WorkflowResult { return a.wf.SomeOp(...) })
```

The goroutine returns `WorkflowResultMsg`, which `app.Update` handles by calling `setStatus` + `refreshAll`. Call `refreshAll()` after any operation that mutates repo state. Do not call git methods directly in `Update`.

### Panel sizing

`relayout()` is called on every `WindowSizeMsg` and focus change. It calls `ui.Widths()` and then `SetSize(w, h)` on each panel. Panels must not store terminal dimensions themselves; they receive them through `SetSize`.

### `BranchesModel` cursor

The branch list includes a virtual «new branch» button at cursor position `localCount()`. `IsNewBranchSelected()` and `Selected()` account for this: `Selected()` returns `nil` when the button is active. Always check `IsNewBranchSelected()` before reading `Selected()` in the branches panel context.

## Adding a new dialog

1. Implement `DialogContent` (`Priority()`, `DialogUpdate`, `View()`).
2. Push it from `app.handleMenuAction` via `a.dialogs.Push(...)`.
3. Handle any submit message in `app.Update` (add a new `case`).
4. Add the menu action constant to `contextmenu.go` and the `MenuAction` const list.

## Adding a new menu action

1. Add a `MenuAction` constant in `contextmenu.go`.
2. Add a `MenuItem` in `BranchMenuItems` or `CommitMenuItems` (with `Key` for direct shortcut).
3. Handle the action in `app.handleMenuAction`.
4. Add a direct-shortcut guard in `branchOwnedKeys` / `commitOwnedKeys` if the key must not be shadowed.

## Testing

Tests use message injection: construct the model, call `Update(msg)`, assert on the returned model state or commands. Do not start a real `tea.Program` in tests. Git operations in tests must use a real `git.Client` pointed at a temp repo — do not mock.
